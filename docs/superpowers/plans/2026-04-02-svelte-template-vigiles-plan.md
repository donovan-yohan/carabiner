# Svelte Template + Vigiles Add-on Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Svelte TypeScript templates (SvelteKit + Svelte+Vite) to carabiner, plus universal Vigiles add-on for lint rule annotation validation.

**Architecture:** Three new Go template files in `internal/carabiner/templates/`: one shared config used by both variants, and two variant-specific files. The CLI's `init` command gains a `--template` flag and add-on prompt. Vigiles add-on is template-agnostic scaffolding that runs `npx skills add zernie/vigiles` and writes minimal CLAUDE.md, GitHub Action, and Claude Code hook.

**Tech Stack:** Go (carabiner CLI), Node.js (vigiles, svelte-check, eslint, prettier), Svelte 5 (runes)

---

## File Structure

```
internal/carabiner/templates/
  svelte_shared.go      # Shared: enforce.yaml, eslint config, prettier config (exported vars)
  svelte_kit.go         # SvelteKit: uses shared + routes/ example files + kit-specific files
  svelte_vite.go        # Svelte+Vite: uses shared + component library files + vite-specific files
  templates.go          # MODIFY: register svelte-kit and svelte-vite

internal/cmd/
  init.go               # MODIFY: add --template flag and add-on prompts

internal/carabiner/
  init.go               # MODIFY: add ApplyVigilesAddOn() function

docs/superpowers/plans/   # This plan
docs/superpowers/specs/    # Already written design spec
```

---

## Task 1: Create svelte_shared.go

**Files:**
- Create: `internal/carabiner/templates/svelte_shared.go`

```go
package templates

// SharedEnforceYAML is the enforce.yaml for all Svelte templates.
const SharedEnforceYAML = `version: 1
tools:
  svelte-check:
    enabled: true
    command: "npx"
    args: ["svelte-check", "--fail-on-warnings"]
    files: ["src/**/*.{svelte,ts}"]

  eslint:
    enabled: true
    command: "npx"
    args: ["eslint", "--max-warnings=0", "."]
    files: ["src/**/*.{svelte,ts,js}"]

  prettier:
    enabled: true
    command: "npx"
    args: ["prettier", "--check", "."]
    files: ["src/**/*.{svelte,ts,js,css}"]

  vitest:
    enabled: true
    command: "npx"
    args: ["vitest", "run"]
    files: ["src/**/*.test.{ts,tsx}"]

behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`

// SharedEslintrc is the .eslintrc.json for all Svelte templates.
const SharedEslintrc = `{
  "extends": [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:svelte/recommended"
  ],
  "plugins": ["@typescript-eslint", "svelte"],
  "rules": {
    "no-console": "error",
    "no-unused-vars": "off",
    "@typescript-eslint/no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
    "max-lines-per-function": ["error", 40],
    "complexity": ["error", 10],
    "max-depth": ["error", 3],
    "max-params": ["error", 3],
    "svelte/valid-compile": "error",
    "svelte/no-at-html-tags": "error"
  }
}
`

// SharedPrettierrc is the .prettierrc for all Svelte templates.
const SharedPrettierrc = `{
  "semi": true,
  "singleQuote": true,
  "trailingComma": "es5"
}
`

// SharedPrettierIgnore is the .prettierignore for all Svelte templates.
const SharedPrettierIgnore = `node_modules
build
dist
.svelte-kit
package
.vite
.env
.env.*
!.env.example
`

// SharedTsconfig is the tsconfig.json for all Svelte templates.
const SharedTsconfig = `{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "allowJs": true,
    "checkJs": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "skipLibCheck": true,
    "sourceMap": true,
    "strict": true,
    "moduleResolution": "bundler"
  }
}
`

// SharedVitestConfig is the vitest.config.ts for all Svelte templates.
const SharedVitestConfig = `import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.{test,spec}.{js,ts}'],
  },
});
`

// SharedPackageScripts is the scripts section for package.json (template uses full package.json).
const SharedPackageScripts = `"scripts": {
  "dev": "vite dev",
  "build": "vite build",
  "preview": "vite preview",
  "check": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json",
  "check:watch": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json --watch",
  "test": "vitest run",
  "test:watch": "vitest"
}
`
```

- [ ] **Step 1: Create svelte_shared.go with shared config constants**

Write the file above.

- [ ] **Step 2: Run tests to verify it compiles**

Run: `go build ./internal/carabiner/templates/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/carabiner/templates/svelte_shared.go
git commit -m "feat(templates): add shared Svelte config constants"
```

---

## Task 2: Create svelte_kit.go

**Files:**
- Create: `internal/carabiner/templates/svelte_kit.go`

```go
package templates

func getSvelteKitTemplate() *Template {
	packageJSON := `{
  "name": "svelte-kit-app",
  "version": "0.0.1",
  "private": true,
  "scripts": ` + SharedPackageScripts + `,
  "devDependencies": {
    "@sveltejs/adapter-auto": "^3.0.0",
    "@sveltejs/kit": "^2.0.0",
    "@sveltejs/vite-plugin-svelte": "^4.0.0",
    "svelte": "^5.0.0",
    "svelte-check": "^4.0.0",
    "typescript": "^5.0.0",
    "vite": "^6.0.0",
    "eslint": "^9.0.0",
    "prettier": "^3.0.0",
    "prettier-plugin-svelte": "^3.0.0",
    "@typescript-eslint/eslint-plugin": "^8.0.0",
    "@typescript-eslint/parser": "^8.0.0",
    "eslint-plugin-svelte": "^2.0.0",
    "vitest": "^2.0.0",
    "jsdom": "^25.0.0"
  },
  "type": "module"
}
`

	svelteConfig := `import adapter from '@sveltejs/adapter-auto';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter(),
  },
};

export default config;
`

	viteConfig := `import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
});
`

	appHTML := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="icon" href="%sveltekit.assets%/favicon.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    %sveltekit.head%
  </head>
  <body data-sveltekit-preload-data="hover">
    <div style="display: contents">%sveltekit.body%</div>
  </body>
</html>
`

	appCSS := `/* Global styles */
:root {
  font-family: system-ui, -apple-system, sans-serif;
  line-height: 1.5;
}

body {
  margin: 0;
  padding: 1rem;
}
`

	routesPageSvelte := `<script lang="ts">
  import Counter from '$lib/components/Counter.svelte';
</script>

<h1>Welcome to SvelteKit</h1>
<Counter />
`

	routesLayoutSvelte := `<script lang="ts">
  import '../app.css';
</script>

<slot />
`

	routesPageServer := `import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
  return {
    timestamp: Date.now(),
  };
};
`

	counterComponent := `<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  function increment() {
    count++;
  }
</script>

<button onclick={increment}>
  Count: {count}, Doubled: {doubled}
</button>
`

	dbServer := `// Server-only database utilities
// Never import this file from client-side code

export async function query<T>(sql: string, params?: unknown[]): Promise<T[]> {
  // TODO: implement database query
  return [] as T[];
}
`

	return &Template{
		Name:        "svelte-kit",
		EnforceYAML: SharedEnforceYAML,
		ConfigFiles: map[string]string{
			"package.json":        packageJSON,
			"svelte.config.js":    svelteConfig,
			"vite.config.ts":      viteConfig,
			"tsconfig.json":       SharedTsconfig,
			".eslintrc.json":      SharedEslintrc,
			".prettierrc":        SharedPrettierrc,
			".prettierignore":    SharedPrettierIgnore,
			"vitest.config.ts":    SharedVitestConfig,
			"src/app.html":        appHTML,
			"src/app.css":         appCSS,
			"src/routes/+page.svelte":        routesPageSvelte,
			"src/routes/+layout.svelte":      routesLayoutSvelte,
			"src/routes/+page.server.ts":    routesPageServer,
			"src/lib/components/Counter.svelte": counterComponent,
			"src/lib/server/db.ts":            dbServer,
		},
	}
}
```

- [ ] **Step 1: Create svelte_kit.go with SvelteKit template**

Write the file above.

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/carabiner/templates/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/carabiner/templates/svelte_kit.go
git commit -m "feat(templates): add svelte-kit template with routing examples"
```

---

## Task 3: Create svelte_vite.go

**Files:**
- Create: `internal/carabiner/templates/svelte_vite.go`

```go
package templates

func getSvelteViteTemplate() *Template {
	packageJSON := `{
  "name": "svelte-vite-app",
  "version": "0.0.1",
  "private": true,
  "scripts": ` + SharedPackageScripts + `,
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^4.0.0",
    "svelte": "^5.0.0",
    "typescript": "^5.0.0",
    "vite": "^6.0.0",
    "eslint": "^9.0.0",
    "prettier": "^3.0.0",
    "prettier-plugin-svelte": "^3.0.0",
    "@typescript-eslint/eslint-plugin": "^8.0.0",
    "@typescript-eslint/parser": "^8.0.0",
    "eslint-plugin-svelte": "^2.0.0",
    "vitest": "^2.0.0",
    "jsdom": "^25.0.0"
  },
  "type": "module"
}
`

	svelteConfig := `import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  preprocess: vitePreprocess(),
};
`

	viteConfig := `import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
});
`

	indexHTML := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="icon" "/favicon.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Svelte App</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
`

	appCSS := `:root {
  font-family: system-ui, -apple-system, sans-serif;
  line-height: 1.5;
}

body {
  margin: 0;
  padding: 1rem;
}
`

	appSvelte := `<script lang="ts">
  import Counter from './lib/components/Counter.svelte';
</script>

<Counter />
`

	mainTS := `import App from './App.svelte';
import './app.css';

const app = new App({
  target: document.getElementById('app')!,
});

export default app;
`

	counterComponent := `<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  function increment() {
    count++;
  }
</script>

<button onclick={increment}>
  Count: {count}, Doubled: {doubled}
</button>
`

	formatUtils := `// Utility functions
export function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}
`

	return &Template{
		Name:        "svelte-vite",
		EnforceYAML: SharedEnforceYAML,
		ConfigFiles: map[string]string{
			"package.json":              packageJSON,
			"svelte.config.js":          svelteConfig,
			"vite.config.ts":            viteConfig,
			"tsconfig.json":             SharedTsconfig,
			".eslintrc.json":           SharedEslintrc,
			".prettierrc":             SharedPrettierrc,
			".prettierignore":         SharedPrettierIgnore,
			"vitest.config.ts":         SharedVitestConfig,
			"index.html":               indexHTML,
			"src/app.css":              appCSS,
			"src/App.svelte":           appSvelte,
			"src/main.ts":              mainTS,
			"src/lib/components/Counter.svelte": counterComponent,
			"src/lib/utils/format.ts":  formatUtils,
		},
	}
}
```

- [ ] **Step 1: Create svelte_vite.go with Svelte+Vite template**

Write the file above.

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/carabiner/templates/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/carabiner/templates/svelte_vite.go
git commit -m "feat(templates): add svelte-vite template with component library examples"
```

---

## Task 4: Register templates in templates.go

**Files:**
- Modify: `internal/carabiner/templates/templates.go`

- [ ] **Step 1: Update GetTemplate switch to include svelte-kit and svelte-vite**

Change:
```go
case "svelte-kit":
    return getSvelteKitTemplate(), nil
case "svelte-vite":
    return getSvelteViteTemplate(), nil
case "svelte":  // alias for svelte-kit
    return getSvelteKitTemplate(), nil
```

Add after existing cases, before default.

- [ ] **Step 2: Update ListTemplates to include new templates**

Change the return array from:
```go
return []string{
    "react-typescript",
    "go",
}
```
to:
```go
return []string{
    "go",
    "react-typescript",
    "svelte-kit",
    "svelte-vite",
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/carabiner/templates/`
Expected: No errors

- [ ] **Step 4: Run existing tests**

Run: `go test ./internal/carabiner/templates/...`
Expected: All existing tests pass

- [ ] **Step 5: Commit**

```bash
git add internal/carabiner/templates/templates.go
git commit -m "feat(templates): register svelte-kit and svelte-vite templates"
```

---

## Task 5: Wire --template flag into init command

**Files:**
- Modify: `internal/cmd/init.go`

The `init` command needs:
1. Add `--template` flag (`string`, default `""`)
2. Add `--add-ons` flag (`stringSlice`, default `[]`)
3. When `--template` is provided, scaffold template files to current directory
4. Prompt interactively if no flags provided (template selection + add-ons)

- [ ] **Step 1: Read current init.go and carabiner/init.go**

Already done above. Review `Init` function signature and what `Init` does.

- [ ] **Step 2: Add flags to init command**

Add to `init.go`:
```go
var initTemplate string
var initAddOns []string

initCmd.Flags().StringVar(&initTemplate, "template", "", "Template to scaffold (go, react-typescript, svelte-kit, svelte-vite)")
initCmd.Flags().StringSliceVar(&initAddOns, "add-ons", nil, "Add-ons to install (e.g., vigiles)")
```

- [ ] **Step 3: Handle --template flag in RunE**

If `initTemplate != ""`, call `carabiner.InitWithTemplate(initTemplate, initAddOns)` instead of just `carabiner.Init(initMode)`.

- [ ] **Step 4: Implement InitWithTemplate in carabiner/init.go**

```go
// InitWithTemplate scaffolds .carabiner/ AND applies a template to the current directory.
func InitWithTemplate(mode, templateName string, addOns []string) error {
    configDir, err := Init(mode)
    if err != nil {
        return err
    }

    // Apply template
    tmpl, err := templates.GetTemplate(templateName)
    if err != nil {
        return fmt.Errorf("template: %w", err)
    }

    cwd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("getting working directory: %w", err)
    }

    // Write enforce.yaml
    enforcePath := filepath.Join(configDir, "enforce.yaml")
    if err := os.WriteFile(enforcePath, []byte(tmpl.EnforceYAML), 0644); err != nil {
        return fmt.Errorf("writing enforce.yaml: %w", err)
    }

    // Write template config files to current directory
    for filename, content := range tmpl.ConfigFiles {
        path := filepath.Join(cwd, filename)
        if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
            return fmt.Errorf("creating directory for %s: %w", filename, err)
        }
        if err := os.WriteFile(path, []byte(content), 0644); err != nil {
            return fmt.Errorf("writing %s: %w", filename, err)
        }
    }

    // Apply add-ons
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

    return nil
}
```

- [ ] **Step 5: Verify it compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 6: Run tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/cmd/init.go internal/carabiner/init.go
git commit -m "feat(init): add --template and --add-ons flags"
```

---

## Task 6: Implement ApplyVigilesAddOn

**Files:**
- Modify: `internal/carabiner/init.go`

Add to end of `init.go`:

```go
// ApplyVigilesAddOn scaffolds Vigiles for the project at cwd.
func ApplyVigilesAddOn(cwd string) error {
    // Create .github/workflows/vigiles.yml
    vigilesWorkflow := `name: Validate agent instructions
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: zernie/vigiles@main
`

    wfDir := filepath.Join(cwd, ".github", "workflows")
    if err := os.MkdirAll(wfDir, 0755); err != nil {
        return fmt.Errorf("creating .github/workflows: %w", err)
    }
    if err := os.WriteFile(filepath.Join(wfDir, "vigiles.yml"), []byte(vigilesWorkflow), 0644); err != nil {
        return fmt.Errorf("writing vigiles.yml: %w", err)
    }

    // Create .claude/settings.json with PostToolUse hook (if .claude/ exists)
    claudeDir := filepath.Join(cwd, ".claude")
    if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
        settings := `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "command": "npx vigiles CLAUDE.md"
      }
    ]
  }
}
`
        if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settings), 0644); err != nil {
            return fmt.Errorf("writing .claude/settings.json: %w", err)
        }
    }

    // Create minimal CLAUDE.md
    claudeMD := `# Agent Guidance

## Before committing
Run \`carabiner enforce --all\` to verify linting passes.

## Feedback loops
When you notice a recurring mistake in code review, run \`/pr-to-lint-rule\` to convert it into an enforced lint rule.

## Quality patterns
Run \`carabiner quality check --files <files>\` before implementation to see relevant learnings from past gate failures.
`

    // Only write CLAUDE.md if it doesn't exist
    claudePath := filepath.Join(cwd, "CLAUDE.md")
    if _, err := os.Stat(claudePath); os.IsNotExist(err) {
        if err := os.WriteFile(claudePath, []byte(claudeMD), 0644); err != nil {
            return fmt.Errorf("writing CLAUDE.md: %w", err)
        }
    }

    return nil
}
```

Note: The actual installation of vigiles (`npm install -D vigiles` and `npx skills add zernie/vigiles`) should be printed as instructions to the user, not executed by the Go code (since Go can't run npm commands reliably cross-platform). Add to the `InitWithTemplate` or `ApplyVigilesAddOn` function to return instructions:

```go
// Returns shell commands to run after scaffolding.
func GetVigilesInstallCommands() []string {
    return []string{
        "npm install -D vigiles",
        "npx skills add zernie/vigiles",
    }
}
```

Print these after scaffolding completes.

- [ ] **Step 1: Add ApplyVigilesAddOn and GetVigilesInstallCommands to init.go**

Write the functions above.

- [ ] **Step 2: Update InitWithTemplate to print install commands**

After applying all add-ons, print:
```
To complete setup, run:
  npm install -D vigiles
  npx skills add zernie/vigiles
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/carabiner/init.go
git commit -m "feat(init): add ApplyVigilesAddOn for vigiles scaffolding"
```

---

## Task 7: Add template tests

**Files:**
- Modify: `internal/carabiner/templates/templates_test.go`

- [ ] **Step 1: Add test for svelte-kit template**

```go
func TestGetTemplate_SvelteKit(t *testing.T) {
    tmpl, err := GetTemplate("svelte-kit")
    if err != nil {
        t.Fatalf("GetTemplate(\"svelte-kit\") failed: %v", err)
    }
    if tmpl.Name != "svelte-kit" {
        t.Errorf("Expected Name \"svelte-kit\", got %q", tmpl.Name)
    }
    if tmpl.EnforceYAML == "" {
        t.Error("EnforceYAML should not be empty")
    }
    if len(tmpl.ConfigFiles) == 0 {
        t.Error("ConfigFiles should not be empty")
    }
}
```

- [ ] **Step 2: Add test for svelte-vite template**

```go
func TestGetTemplate_SvelteVite(t *testing.T) {
    tmpl, err := GetTemplate("svelte-vite")
    if err != nil {
        t.Fatalf("GetTemplate(\"svelte-vite\") failed: %v", err)
    }
    if tmpl.Name != "svelte-vite" {
        t.Errorf("Expected Name \"svelte-vite\", got %q", tmpl.Name)
    }
}
```

- [ ] **Step 3: Add test for svelte alias**

```go
func TestGetTemplate_SvelteAlias(t *testing.T) {
    tmpl, err := GetTemplate("svelte")
    if err != nil {
        t.Fatalf("GetTemplate(\"svelte\") failed: %v", err)
    }
    // svelte alias should resolve to svelte-kit
    if tmpl.Name != "svelte-kit" {
        t.Errorf("Expected svelte alias to resolve to \"svelte-kit\", got %q", tmpl.Name)
    }
}
```

- [ ] **Step 4: Add test that svelte-kit template has routing files**

```go
func TestSvelteKitTemplate_HasRoutingFiles(t *testing.T) {
    tmpl, err := GetTemplate("svelte-kit")
    if err != nil {
        t.Fatalf("GetTemplate failed: %v", err)
    }
    requiredFiles := []string{
        "src/routes/+page.svelte",
        "src/routes/+layout.svelte",
        "src/routes/+page.server.ts",
        "src/lib/components/Counter.svelte",
        "svelte.config.js",
        "vite.config.ts",
    }
    for _, f := range requiredFiles {
        if _, ok := tmpl.ConfigFiles[f]; !ok {
            t.Errorf("SvelteKit template missing file: %s", f)
        }
    }
    if !strings.Contains(tmpl.EnforceYAML, "svelte-check") {
        t.Error("enforce.yaml should contain svelte-check")
    }
}
```

- [ ] **Step 5: Add test that svelte-vite template has component files but no routes**

```go
func TestSvelteViteTemplate_HasComponentFiles(t *testing.T) {
    tmpl, err := GetTemplate("svelte-vite")
    if err != nil {
        t.Fatalf("GetTemplate failed: %v", err)
    }
    requiredFiles := []string{
        "src/App.svelte",
        "src/main.ts",
        "src/lib/components/Counter.svelte",
        "vite.config.ts",
    }
    for _, f := range requiredFiles {
        if _, ok := tmpl.ConfigFiles[f]; !ok {
            t.Errorf("SvelteVite template missing file: %s", f)
        }
    }
    // Should NOT have routes
    if _, ok := tmpl.ConfigFiles["src/routes/+page.svelte"]; ok {
        t.Error("SvelteVite template should not have routing files")
    }
}
```

- [ ] **Step 6: Add test that both templates have fail_on_warning: true**

```go
func TestSvelteTemplates_FailOnWarning(t *testing.T) {
    for _, name := range []string{"svelte-kit", "svelte-vite"} {
        tmpl, err := GetTemplate(name)
        if err != nil {
            t.Fatalf("GetTemplate(%q) failed: %v", name, err)
        }
        if !strings.Contains(tmpl.EnforceYAML, "fail_on_warning: true") {
            t.Errorf("%s enforce.yaml should have fail_on_warning: true", name)
        }
        if !strings.Contains(tmpl.EnforceYAML, "svelte-check") {
            t.Errorf("%s enforce.yaml should contain svelte-check", name)
        }
    }
}
```

- [ ] **Step 7: Run tests**

Run: `go test ./internal/carabiner/templates/... -v`
Expected: All new tests pass

- [ ] **Step 8: Commit**

```bash
git add internal/carabiner/templates/templates_test.go
git commit -m "test(templates): add tests for svelte-kit and svelte-vite"
```

---

## Task 8: Update ListTemplates in test

**Files:**
- Modify: `internal/carabiner/templates/templates_test.go`

- [ ] **Step 1: Update TestListTemplates to expect 4 templates**

Change:
```go
if len(templates) < 2 {
    t.Errorf("Expected at least 2 templates, got %d", len(templates))
}
```
to:
```go
if len(templates) != 4 {
    t.Errorf("Expected 4 templates, got %d", len(templates))
}
```

Add checks for svelte-kit and svelte-vite.

- [ ] **Step 2: Run tests**

Run: `go test ./internal/carabiner/templates/...`
Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add internal/carabiner/templates/templates_test.go
git commit -m "test(templates): update TestListTemplates for new templates"
```

---

## Verification

After all tasks:

- [ ] Run: `go build ./... && go test ./...` — all pass
- [ ] Run: `go run ./cmd/carabiner init --help` — shows `--template` and `--add-ons` flags
- [ ] Run: `go run ./cmd/carabiner init --template svelte-kit --add-ons vigiles` in a temp directory — scaffolds files correctly
- [ ] Verify: `enforce.yaml` contains `svelte-check --fail-on-warnings`
- [ ] Verify: `src/routes/+page.svelte` exists for svelte-kit
- [ ] Verify: `src/App.svelte` exists for svelte-vite
- [ ] Verify: `.github/workflows/vigiles.yml` created when vigiles add-on selected
- [ ] Verify: `CLAUDE.md` created when vigiles add-on selected
