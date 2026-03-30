package carabiner

import (
	"fmt"
	"os"
	"path/filepath"

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

	return configDir, nil
}
