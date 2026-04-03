# Svelte TypeScript Template + Vigiles Add-on

**Date:** 2026-04-02
**Status:** Approved

## Overview

Add Svelte TypeScript templates to carabiner (SvelteKit and Svelte+Vite variants) plus a universal Vigiles add-on for lint rule annotation validation. Carabiner's role is minimal scaffolding — it sets up tooling, not rule content.

## Templates

### Files to Add

```
internal/carabiner/templates/
  svelte.go          # Shared: enforce.yaml, eslint, prettier, vitest configs
  svelte_kit.go      # SvelteKit: routing structure, +page.svelte, +layout.svelte, etc.
  svelte_vite.go     # Svelte+Vite: component library structure, App.svelte, main.ts
```

### Template Registration

In `templates.go`, add:
- `svelte-kit` → `getSvelteKitTemplate()`
- `svelte-vite` → `getSvelteViteTemplate()`
- `svelte` → aliases to `svelte-kit` (or prompts user to choose)

---

## Shared Tooling Stack (enforce.yaml)

All Svelte templates share identical tooling:

```yaml
version: 1
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
```

**Key decision:** `--fail-on-warnings` on `svelte-check` makes compiler warnings blockers. For agents, this forces attention to deprecation warnings, a11y issues, and missing `lang="ts"` attributes immediately rather than accumulating.

---

## ESLint Config

Shared `.eslintrc.json` for all Svelte templates:

```json
{
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
```

**Complexity rules enforce decomposition.** Per Ernie's feedback loop article: without caps, agents generate 150-line functions with 6 levels of nesting. Set `max-lines-per-function: 40`, `complexity: 10`, `max-depth: 3` and the agent *has* to extract helpers.

---

## SvelteKit Variant (src/lib/components/)

### Example Files

**`src/routes/+page.svelte`**
```svelte
<script lang="ts">
  import Counter from '$lib/components/Counter.svelte';
</script>

<h1>Welcome</h1>
<Counter />
```

**`src/routes/+page.server.ts`**
```ts
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
  return { timestamp: Date.now() };
};
```

**`src/routes/+layout.svelte`**
```svelte
<script lang="ts">
  import '../app.css';
</script>

<slot />
```

**`src/lib/components/Counter.svelte`** — Svelte 5 runes example:
```svelte
<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  function increment() {
    count++;
  }
</script>

<button onclick={increment}>
  Count: {count}, Doubled: {doubled}
</button>
```

### File Structure

```
svelte-kit/
  .eslintrc.json
  .prettierrc
  svelte.config.js
  vite.config.ts
  tsconfig.json
  package.json
  src/
    app.html
    app.css
    routes/
      +page.svelte
      +layout.svelte
      +page.server.ts
    lib/
      components/
        Counter.svelte
      server/
        db.ts
```

---

## Svelte+Vite Variant (src/lib/components/)

Same tooling configs as SvelteKit. Different structure (no routing):

```
svelte-vite/
  .eslintrc.json
  .prettierrc
  svelte.config.js
  vite.config.ts
  tsconfig.json
  package.json
  src/
    app.html
    app.css
    lib/
      components/
        Counter.svelte
      utils/
        format.ts
    App.svelte
    main.ts
```

**`src/App.svelte`** — Root component:
```svelte
<script lang="ts">
  import Counter from './lib/components/Counter.svelte';
</script>

<Counter />
```

**`src/main.ts`** — Mount point:
```ts
import App from './App.svelte';

const app = new App({ target: document.getElementById('app')! });

export default app;
```

---

## Vigiles Add-on (Universal)

Template-agnostic. Works with Go, React+TS, SvelteKit, Svelte+Vite — any language.

### What It Does

Vigiles validates that rules in CLAUDE.md have `**Enforced by:**` or `**Guidance only:**` annotations. It does NOT generate rule content — that accumulates over time via vigiles' `/pr-to-lint-rule` skill.

### Activation

```
carabiner init --template svelte-kit --add-ons vigiles
```

Or prompted after template selection:
```
? Add optional add-ons:
  ○ Vigiles — lint rule annotations for CLAUDE.md (recommended: No)
```

### Setup Steps

1. **Install:** `npx skills add zernie/vigiles`
2. **Scripts:** Add `"vigiles": "vigiles"` to package.json
3. **GitHub Action:** Create `.github/workflows/vigiles.yml`
4. **Claude Code hook:** Create `.claude/settings.json` with PostToolUse hook (if `.claude/` detected)
5. **Skills:** Run `npx skills add zernie/vigiles` to install vigiles skills
6. **CLAUDE.md:** Create minimal scaffold (see below)
7. **enforce.yaml:** Add vigiles tool entry

### GitHub Action (`.github/workflows/vigiles.yml`)

```yaml
name: Validate agent instructions
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: zernie/vigiles@main
```

### Claude Code Hook (`.claude/settings.json`)

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "command": "npx vigiles CLAUDE.md"
      }
    ]
  }
}
```

### Minimal CLAUDE.md

```markdown
# Agent Guidance

## Before committing
Run `carabiner enforce --all` to verify linting passes.

## Feedback loops
When you notice a recurring mistake in code review, run `/pr-to-lint-rule` to convert it into an enforced lint rule.

## Quality patterns
Run `carabiner quality check --files <files>` before implementation to see relevant learnings from past gate failures.
```

**Carabiner does NOT write template-specific rules into CLAUDE.md.** Those accumulate via the `/pr-to-lint-rule` workflow when patterns recur.

### enforce.yaml Entry

```yaml
vigiles:
  enabled: true
  command: "npx"
  args: ["vigiles"]
  files: ["CLAUDE.md", "AGENTS.md", ".cursorrules", ".windsurfrules"]
```

---

## init Flow

```
carabiner init
  → Select template:
      1. go
      2. react-typescript
      3. svelte-kit
      4. svelte-vite
  → Optional add-ons:
      Vigiles? (recommended: No)
  → Scaffold files
  → Run npm install (if package.json modified)
  → Run npx skills add zernie/vigiles (if vigiles add-on selected)
```

---

## Relationship to Quality/Enforcement Loop

```
Agent ships code → carabiner gate fails → carabiner quality record
    ↓
Pattern noticed → vigiles /pr-to-lint-rule "we keep importing from antd"
    ↓
Lint rule generated → CLAUDE.md annotated → enforced forever
    ↓
Next agent hits same mistake → lint blocks it before gate
```

Vigiles and carabiner are complementary:
- **carabiner** detects gate failures and records patterns
- **vigiles** converts recurring patterns into permanent lint rules
- **enforce.yaml** blocks on lint failures

---

## TODO

- [ ] Implement `templates/svelte.go` with shared tooling configs
- [ ] Implement `templates/svelte_kit.go` with routing example files
- [ ] Implement `templates/svelte_vite.go` with component library example
- [ ] Update `templates.go` to register `svelte-kit` and `svelte-vite`
- [ ] Add Vigiles add-on detection in `cmd/init.go`
- [ ] Implement Vigiles add-on scaffolding in `carabiner/init.go` or `templates/`
- [ ] Add tests for template generation
- [ ] Update CLAUDE.md docs to reflect new templates
