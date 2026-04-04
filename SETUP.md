# Carabiner Setup Guide

## What is Carabiner?

Carabiner is a forensic query layer for AI-coded repos. It answers: "What agent session wrote this code, what model was it using, and what was the session doing?"

It works by joining two existing tools:
- **git-ai**: tracks which agent wrote which lines (Git Notes at `refs/notes/ai`)
- **agentlytics**: records session data from all AI agents (SQLite cache)

Carabiner reads both and connects them. One command, zero config.

## Prerequisites

### 1. git-ai (required)

git-ai provides line-level AI attribution via Git Notes. Carabiner needs this to know which agent session wrote which code.

```bash
curl -sSL https://usegitai.com/install.sh | bash
git-ai install-hooks
```

After `git-ai install-hooks`, future commits made with AI agents will automatically get attribution notes. Existing commits won't have notes (git-ai only tracks going forward).

### 2. agentlytics (recommended)

agentlytics records session metadata (name, timestamps, source) from all major AI agents into a local SQLite cache.

```bash
npx agentlytics
```

This scans your agent session stores and builds `~/.agentlytics/cache.db`. Run it periodically to keep the cache current.

## Installing Carabiner

```bash
go install github.com/donovan-yohan/carabiner/cmd/carabiner@latest
```

## Verify Your Setup

```bash
carabiner doctor
```

Doctor checks all three data sources and tells you what's available:

```
carabiner doctor
================

  [OK] Git repository
  [OK] git-ai notes (refs/notes/ai)
  [OK] agentlytics cache (7325 sessions indexed)

Ready. Run: carabiner why <file>:<line>
```

If anything is missing, doctor gives you the install command.

## Using Carabiner

### Forensic query

```bash
carabiner why src/auth/handler.go:47
```

This traces line 47 back through:
1. `git blame` to find which commit introduced it
2. git-ai notes to find which agent session wrote it
3. agentlytics to enrich with session name, model, timestamps

### JSON output

```bash
carabiner why src/auth/handler.go:47 --json
carabiner doctor --json
```

Both commands support `--json` for structured output. Useful for scripts, CI, and agents calling carabiner programmatically.

### Historical queries

```bash
carabiner why src/auth/handler.go:47 --rev abc1234
```

Blame against a specific commit instead of the working tree.

## Work-Item Linkage (Workflow Recommendation)

Carabiner doesn't parse branch names or grep commit messages for ticket numbers. Instead, use your platform's existing integration:

- **Linear**: Enable the GitHub integration. PRs linked to Linear issues automatically.
- **Jira**: Use smart commits or the GitHub/GitLab integration.
- **GitHub Issues**: Reference issues in PR descriptions.

The chain becomes: line (carabiner) → commit → PR → ticket → spec. Each hop is handled by the tool that owns it.

## For Agents

Any agent that can run `sh -c` can use carabiner:

```
Run `carabiner doctor` to check data source availability.
Run `carabiner why <file>:<line>` to trace a line back to its agent session.
Use --json for structured output.
```
