# carabiner

**git blame for intent, not just authorship.**

Carabiner answers the question no other tool can: "A bug shipped last week. What agent session wrote this code, what was it thinking, what spec was it working from, and was the spec even clear?"

The data already exists. [git-ai](https://github.com/git-ai-project/git-ai) tracks which agent wrote which lines. [agentlytics](https://github.com/f/agentlytics) records full session transcripts. Linear and Jira link PRs to work items. But nobody joins them. Carabiner is the join.

## How It Works

```bash
$ carabiner why src/auth/handler.go:47

LINE: src/auth/handler.go:47 (introduced in commit abc1234)
SESSION: Claude Code session 8f3a2b1c (2026-03-28 14:32-15:47)
MODEL: claude-sonnet-4-5-20250514
CONFIDENCE: high (deterministic join via git-ai)

WORK ITEM: ENG-42 "Implement token refresh flow"
  Source: branch name (feature/ENG-42)
  Spec: docs/designs/auth-refresh.md (linked in issue)

SESSION CONTEXT:
  The agent was implementing token refresh per the spec. It chose to
  cache the refresh token in-memory rather than using the session store.
  The spec was ambiguous on storage location (line 34: "persist the token").

SUBSEQUENT TOUCHES:
  - Commit def5678 (2026-03-29): human edit, no agent session
  - Commit ghi9012 (2026-03-30): Codex session 4e7f9a2b, same work item
```

One command. Zero config. Works with any AI coding agent.

## The Stack

Carabiner doesn't reinvent collection or attribution. It connects tools that already exist:

| Layer | Tool | What it does |
|-------|------|--------------|
| **Attribution** | [git-ai](https://github.com/git-ai-project/git-ai) | Line-level authorship via Git Notes. Which agent wrote which lines. |
| **Collection** | [agentlytics](https://github.com/f/agentlytics) | Session data from all agents. Full transcripts, tool calls, timestamps. |
| **Join** | **carabiner** | Connects attribution to session data to work items. The forensic query layer. |
| **Work items** | Linear, Jira, GitHub | PRs linked to issues via branch names and integrations. |

The join key: git-ai stores `agent_id.id` (the raw conversation UUID) in each Git Note. agentlytics indexes sessions by the same conversation ID. Carabiner reads both and connects them.

## Graceful Degradation

Carabiner works with whatever data sources are available:

| What's installed | Confidence | What you get |
|-----------------|------------|--------------|
| git-ai + agentlytics | **High** | Full dossier: session, transcript, model, work item, spec |
| git-ai only | **Partial** | Session attribution from Git Notes, no transcript data |
| agentlytics only | **Medium** | Timestamp + file overlap correlation to match sessions to commits |
| Neither | — | Nothing to join. Install git-ai and agentlytics first. |

## Quick Start

```bash
# 1. Install the foundations (if you haven't already)
curl -sSL https://usegitai.com/install.sh | bash   # git-ai: line attribution
npx agentlytics                                      # agentlytics: session collection

# 2. Install carabiner
go install github.com/donovan-yohan/carabiner/cmd/carabiner@latest

# 3. Check what data sources are available
carabiner doctor

# 4. Ask "why" about any line
carabiner why src/auth/handler.go:47
```

## CLI

```bash
carabiner why <file>:<line> [--rev <commit>]   # forensic dossier for a line of code
carabiner doctor [--json]                       # detect available data sources
```

## Confidence Model

Every hop in the attribution chain carries a confidence label. Carabiner never tells a clean story when the evidence is messy. If the chain is weak, the dossier says so and explains which hop is uncertain.

| Hop | High | Medium | Low |
|-----|------|--------|-----|
| Line to commit | git blame (deterministic) | — | — |
| Commit to session | git-ai note (hash join) | timestamp + file Jaccard | timestamp only |
| Session to transcript | agentlytics match | — | — |
| Commit to work item | explicit trailer or PR link | branch name pattern | commit message grep |

## Philosophy

See [docs/PHILOSOPHY.md](docs/PHILOSOPHY.md) for the reasoning behind carabiner's design: why the join is the product, why confidence is core (not UI garnish), why we compose with existing tools instead of competing, and the relationship between attribution, collection, and forensics.

## Relationship to Belayer

[Belayer](https://github.com/donovan-yohan/belayer) is the orchestrator (YAML pipelines, node execution, gate contracts). Carabiner is the forensic query layer. They serve different purposes and work independently. A belayer framework can call `carabiner why` to enrich gate failure reports with session context, but neither requires the other.

## Status

Active development. Currently validating the proof-of-join between git-ai and agentlytics data. See [docs/TODOS.md](docs/TODOS.md) for the roadmap.
