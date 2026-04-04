# Carabiner Setup Guide

## What is Carabiner?

Carabiner is an agent-agnostic harness for coding agents. It provides two critical functions:

1. **Quality**: Persists patterns learned from review failures across sessions
2. **Enforcement**: Runs deterministic feed-forward checks before work starts

Every coding agent session starts fresh. Claude doesn't remember that auth route changes need middleware registry updates. Codex doesn't know that your billing system stores absolute values and CREDIT rows must be subtracted.

Carabiner is the repo's institutional memory for agents. It persists quality patterns across sessions and encodes domain knowledge that no amount of code reading reveals.

## Setting Up Context

Before starting meaningful work, set your work context:

```bash
carabiner context set --work-item <ref> [--spec <ref>]
```

This context is branch-scoped and validated on every commit. The context includes:

- **Work Item Reference**: The issue, ticket, or task being worked on
- **Spec Reference** (optional): The specification or design document
- **Context Branch**: The git branch for this work

### Context Commands

```bash
# Set context for current branch
carabiner context set --work-item ISSUE-123 --spec docs/design.md

# Show current context
carabiner context show

# Show context in JSON format
carabiner context show --json

# Clear context for current branch
carabiner context clear
```

## How Hooks Work

Carabiner installs two git hooks during `init`:

### Pre-Commit Hook

Validates that a work context exists before allowing commits:

```bash
#!/bin/sh
carabiner context show >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "carabiner: no valid work context for current branch"
  echo "Run: carabiner context set --work-item <ref> [--spec <ref>]"
  exit 1
fi
```

If no valid context exists, the commit is blocked with instructions to set context.

### Commit-Msg Hook

Appends carabiner trailers to commit messages:

```
Carabiner-Work-Item: ISSUE-123
Carabiner-Spec: docs/design.md
Carabiner-Context-Branch: feature/auth
```

These trailers create an audit trail linking commits to work items and specs.

## Initializing Carabiner in a New Repo

```bash
# Install carabiner
go install github.com/donovan-yohan/carabiner/cmd/carabiner@latest

# Initialize with template
carabiner init --template go
# or
carabiner init --template react-typescript

# Verify enforcement passes
carabiner enforce --all
```

### Templates

- `go`: Go project with golangci-lint, gofmt, go vet
- `react-typescript`: React + TypeScript project with eslint, tsc, prettier

### Add-Ons

- `vigiles`: Validates agent instruction files (CLAUDE.md, AGENTS.md)

```bash
carabiner init --template go --add-ons vigiles
```

## Agent Onboarding Prompt

Paste this into your agent session to install and initialize carabiner:

```
Install carabiner: run `go install github.com/donovan-yohan/carabiner/cmd/carabiner@latest`
Then initialize: `carabiner init --template <template> --add-ons vigiles`
Templates: go | react-typescript
Recommended: `carabiner init --template <your-framework> --add-ons vigiles`
After init: run `carabiner enforce --all` before starting work
```

## Before Committing

Always run:

```bash
carabiner enforce --all
```

This verifies that all configured static analysis tools pass before committing.

## Quality Patterns

Before implementation, check for relevant learnings:

```bash
carabiner quality check --files <files>
```

This surfaces patterns from past gate failures that apply to the files you're about to modify.