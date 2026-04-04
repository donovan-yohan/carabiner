package carabiner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/donovan-yohan/carabiner/internal/carabiner/templates"
	"gopkg.in/yaml.v3"
)

// Init scaffolds the .carabiner/ directory structure.
func Init(mode string) (string, error) {
	var configDir string

	switch mode {
	case "repo":
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting working directory: %w", err)
		}
		configDir = filepath.Join(cwd, ".carabiner")
	case "local":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home directory: %w", err)
		}
		slug := RepoSlug()
		configDir = filepath.Join(home, ".carabiner", "projects", slug)
	default:
		return "", fmt.Errorf("invalid mode %q: must be 'repo' or 'local'", mode)
	}

	if info, err := os.Stat(configDir); err == nil && info.IsDir() {
		fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", configDir)
		return configDir, nil
	}

	dirs := []string{
		configDir,
		filepath.Join(configDir, "quality", "learnings"),
		filepath.Join(configDir, "quality", "signals"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	cfg := DefaultConfig(mode)
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("marshaling default config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0644); err != nil {
		return "", fmt.Errorf("writing config.yaml: %w", err)
	}

	gitignore := "# Signals are local-only operational data\nsignals.jsonl\n"
	gitignorePath := filepath.Join(configDir, "quality", "signals", ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
		return "", fmt.Errorf("writing .gitignore: %w", err)
	}

	db, err := events.InitDB(filepath.Join(configDir, "carabiner.db"))
	if err != nil {
		return "", fmt.Errorf("initializing events database: %w", err)
	}
	if err := db.Close(); err != nil {
		return "", fmt.Errorf("closing events database: %w", err)
	}

	return configDir, nil
}

func InitWithTemplate(mode, templateName string, addOns []string) error {
	configDir, err := Init(mode)
	if err != nil {
		return err
	}

	tmpl, err := templates.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	enforcePath := filepath.Join(configDir, "enforce.yaml")
	if err := os.WriteFile(enforcePath, []byte(tmpl.EnforceYAML), 0644); err != nil {
		return fmt.Errorf("writing enforce.yaml: %w", err)
	}

	for filename, content := range tmpl.ConfigFiles {
		path := filepath.Join(cwd, filename)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", filename, err)
		}
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", path)
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
	}

	for _, addOn := range addOns {
		switch addOn {
		case "vigiles":
			if err := ApplyVigilesAddOn(cwd); err != nil {
				return fmt.Errorf("vigiles add-on: %w", err)
			}
		default:
			return fmt.Errorf("unknown add-on: %s", addOn)
		}
	}

	if len(addOns) > 0 {
		fmt.Fprintf(os.Stderr, "\nTo complete setup, run:\n")
		for _, addOn := range addOns {
			switch addOn {
			case "vigiles":
				for _, cmd := range GetVigilesInstallCommands() {
					fmt.Fprintf(os.Stderr, "  %s\n", cmd)
				}
			}
		}
	}

	return nil
}

func ApplyVigilesAddOn(cwd string) error {
	vigilesWorkflow := "name: Validate agent instructions\n" +
		"on: [push, pull_request]\n" +
		"jobs:\n" +
		"  validate:\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - uses: actions/checkout@v4\n" +
		"      - uses: zernie/vigiles@4eaa5f4\n"

	wfDir := filepath.Join(cwd, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		return fmt.Errorf("creating .github/workflows: %w", err)
	}
	wfPath := filepath.Join(wfDir, "vigiles.yml")
	if _, err := os.Stat(wfPath); err == nil {
		fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", wfPath)
	} else if err := os.WriteFile(wfPath, []byte(vigilesWorkflow), 0644); err != nil {
		return fmt.Errorf("writing vigiles.yml: %w", err)
	}

	claudeDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("creating .claude: %w", err)
	}
	settings := "{\n" +
		"  \"hooks\": {\n" +
		"    \"PostToolUse\": [\n" +
		"      {\n" +
		"        \"matcher\": \"Edit|Write\",\n" +
		"        \"hooks\": [\n" +
		"          {\n" +
		"            \"type\": \"command\",\n" +
		"            \"command\": \"npx vigiles CLAUDE.md\"\n" +
		"          }\n" +
		"        ]\n" +
		"      }\n" +
		"    ]\n" +
		"  }\n" +
		"}\n"
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", settingsPath)
	} else if err := os.WriteFile(settingsPath, []byte(settings), 0644); err != nil {
		return fmt.Errorf("writing .claude/settings.json: %w", err)
	}

	claudeMD := "# Agent Guidance\n\n" +
		"## Before committing\n" +
		"Run `carabiner enforce --all` to verify linting passes.\n\n" +
		"## Feedback loops\n" +
		"When you notice a recurring mistake in code review, run `/pr-to-lint-rule` to convert it into an enforced lint rule.\n\n" +
		"## Quality patterns\n" +
		"Run `carabiner quality check --files <files>` before implementation to see relevant learnings from past gate failures.\n"

	claudePath := filepath.Join(cwd, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", claudePath)
	} else if err := os.WriteFile(claudePath, []byte(claudeMD), 0644); err != nil {
		return fmt.Errorf("writing CLAUDE.md: %w", err)
	}

	return nil
}

func GetVigilesInstallCommands() []string {
	return []string{
		"npx skills add zernie/vigiles",
	}
}

func ScaffoldContextSupport() error {
	if err := scaffoldHooks(); err != nil {
		return err
	}
	if err := scaffoldAgentInstructions(); err != nil {
		return err
	}
	return nil
}

func scaffoldHooks() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	hooksDir := filepath.Join(cwd, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(preCommitPath); err == nil {
		fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", preCommitPath)
	} else {
		if err := os.WriteFile(preCommitPath, []byte(RenderPreCommitHook()), 0755); err != nil {
			return fmt.Errorf("writing pre-commit hook: %w", err)
		}
	}

	commitMsgPath := filepath.Join(hooksDir, "commit-msg")
	if _, err := os.Stat(commitMsgPath); err == nil {
		fmt.Fprintf(os.Stderr, "warning: %s already exists, skipping\n", commitMsgPath)
	} else {
		if err := os.WriteFile(commitMsgPath, []byte(RenderCommitMsgHook()), 0755); err != nil {
			return fmt.Errorf("writing commit-msg hook: %w", err)
		}
	}

	return nil
}

func scaffoldAgentInstructions() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	contextInstructions := `## Carabiner Work Context

Before meaningful work, set your context:
- ` + "`carabiner context set --work-item <ref> [--spec <ref>]`" + `

Before spawning subagents, pass ` + "`workItemRef`" + ` and ` + "`specRef`" + `.
Do not commit without valid carabiner context for the current branch.

Context is branch-scoped and validated on commit.
`

	claudePath := filepath.Join(cwd, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		existing, err := os.ReadFile(claudePath)
		if err != nil {
			return fmt.Errorf("reading CLAUDE.md: %w", err)
		}

		if !bytes.Contains(existing, []byte("Carabiner Work Context")) {
			updated := []byte(contextInstructions + "\n" + string(existing))
			if err := os.WriteFile(claudePath, updated, 0644); err != nil {
				return fmt.Errorf("updating CLAUDE.md: %w", err)
			}
		}
	} else {
		agentsPath := filepath.Join(cwd, "AGENTS.md")
		if _, err := os.Stat(agentsPath); err == nil {
			existing, err := os.ReadFile(agentsPath)
			if err != nil {
				return fmt.Errorf("reading AGENTS.md: %w", err)
			}

			if !bytes.Contains(existing, []byte("Carabiner Work Context")) {
				updated := []byte(contextInstructions + "\n" + string(existing))
				if err := os.WriteFile(agentsPath, updated, 0644); err != nil {
					return fmt.Errorf("updating AGENTS.md: %w", err)
				}
			}
		} else {
			if err := os.WriteFile(agentsPath, []byte(contextInstructions), 0644); err != nil {
				return fmt.Errorf("writing AGENTS.md: %w", err)
			}
		}
	}

	return nil
}
