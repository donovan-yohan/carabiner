# carabiner

**git blame for intent, not just authorship.**

Carabiner answers the question no other tool can: "A bug shipped last week. What agent session wrote this code, what model was it using, and what was the session doing?"

The data already exists. [git-ai](https://github.com/git-ai-project/git-ai) tracks which agent wrote which lines. [agentlytics](https://github.com/f/agentlytics) records full session transcripts. But nobody joins them. Carabiner is the join.

## How It Works

```bash
$ carabiner why src/auth/handler.go:47

LINE: src/auth/handler.go:47 (introduced in commit abc1234)
AUTHOR: Claude Code
CONFIDENCE: high

SESSION: claude_code session 655eb4a6-822f-4d52-8c06-cade4afdcd8d
MODEL: claude-sonnet-4-5-20250514
NAME: Implement token refresh flow
SOURCE: claude-code
PERIOD: 2026-03-28 14:32 to 2026-03-28 15:47

ATTRIBUTION CHAIN:
  [OK] line_to_commit: git blame (commit abc1234)
  [OK] commit_to_session: git-ai note (session a1b2c3d4e5f6g7h8)
  [OK] session_to_transcript: agentlytics match (session "Implement token refresh flow")
```

One command. Zero config. Works with any AI coding agent.

## The Stack

Carabiner doesn't reinvent collection or attribution. It connects tools that already exist:

| Layer | Tool | What it does |
|-------|------|--------------|
| **Attribution** | [git-ai](https://github.com/git-ai-project/git-ai) | Line-level authorship via Git Notes. Which agent wrote which lines. |
| **Collection** | [agentlytics](https://github.com/f/agentlytics) | Session data from all agents. Full transcripts, tool calls, timestamps. |
| **Join** | **carabiner** | Connects attribution to session data. The forensic query layer. |

The join key: git-ai stores `agent_id.id` (the raw conversation UUID) in each Git Note. agentlytics indexes sessions by the same conversation ID. Carabiner reads both and connects them.

## Quick Start

```bash
# 1. Install the foundations
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
carabiner why <file>:<line> [--rev <commit>] [--json]   # forensic dossier for a line
carabiner doctor [--json]                                # detect available data sources
```

## Confidence Model

Every hop in the attribution chain carries a confidence label: **high** (deterministic join via git-ai) or **missing** (no data for this hop). If any hop is missing, the overall confidence is missing. Carabiner never tells a clean story when the evidence is messy.

| Hop | High | Missing |
|-----|------|---------|
| Line to commit | git blame (deterministic) | — |
| Commit to session | git-ai note (hash join) | No note for commit |
| Session to transcript | agentlytics match | Session not in cache |

## Philosophy

See [docs/PHILOSOPHY.md](docs/PHILOSOPHY.md) for the reasoning behind carabiner's design: why the join is the product, why git-ai is required, why confidence is binary (high or missing), and why we compose with existing tools instead of competing.

## Relationship to Belayer

[Belayer](https://github.com/donovan-yohan/belayer) is the orchestrator (YAML pipelines, node execution, gate contracts). Carabiner is the forensic query layer. They serve different purposes and work independently. A belayer framework can call `carabiner why` to enrich gate failure reports with session context, but neither requires the other.

## Status

Active development. Join-layer MVP complete: `carabiner why` and `carabiner doctor` are working. See [docs/TODOS.md](docs/TODOS.md) for the roadmap.
