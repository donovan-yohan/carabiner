# carabiner

Agent-agnostic harness for coding agents. Two jobs: **quality** (patterns learned from review failures) and **enforcement** (deterministic feed-forward checks).

Companion to [belayer](https://github.com/donovan-yohan/belayer) (orchestrator). You can use carabiner without belayer, and belayer without carabiner. They compose but don't depend on each other.

## Why

Coding agents are ephemeral. Every session starts fresh. Claude doesn't remember that auth route changes need middleware registry updates. Codex doesn't know that your billing system stores absolute values and CREDIT rows must be subtracted.

Carabiner is the repo's institutional memory for agents. It persists quality patterns across sessions and encodes domain knowledge that no amount of code reading reveals. Any agent that can run a shell command can use it.

## Two Jobs

### Enforcement (feed-forward, deterministic)

Before work starts, carabiner ensures conditions exist. Static analysis tools (eslint, golangci-lint, tsc) must pass. No bypass.

```bash
carabiner enforce --all              # run all configured tools
carabiner enforce --tool eslint     # run specific tool
```

Exit codes: 0=pass, 1=enforcement fail, 2=config error.

### Quality (reactive, learned from failures)

When an adversarial review gate catches a bug, carabiner records the pattern. Next time an agent touches those files, carabiner surfaces the relevant patterns. Every failed review makes future implementations better.

```bash
carabiner quality check --files src/auth/routes.ts    # what patterns apply here?
carabiner quality record --gate-id <id>               # capture a learning from a gate failure
```

### Knowledge (proactive, human-authored)

Domain knowledge agents need but can't derive from code. Business logic semantics, architectural invariants, conventions that look arbitrary without context.

```bash
carabiner knowledge query --context "calculating total spend"   # what do I need to know?
```

**Status: aspirational.** Enforcement and Quality are MVP. Knowledge is the next horizon. See [docs/TODOS.md](docs/TODOS.md).

## Quick Start

```bash
go install github.com/donovan-yohan/carabiner/cmd/carabiner@latest
cd your-repo
carabiner init
carabiner enforce --all   # verify your tools pass
```

## For Agents

Paste this into your session to install and initialize carabiner:

```
Install carabiner: run `go install github.com/donovan-yohan/carabiner/cmd/carabiner@latest`
Then initialize: `carabiner init --template <template> --add-ons vigiles`
Templates: go | react-typescript
Recommended: `carabiner init --template <your-framework> --add-ons vigiles`
After init: run `carabiner enforce --all` before starting work
```

Non-interactive (pick one):
```bash
carabiner init --template go
carabiner init --template react-typescript
```

## Event Log

Every carabiner invocation is logged to SQLite for workflow observability.

```bash
carabiner events list                  # recent invocations
carabiner events list --limit 20      # with limit
carabiner events list --command enforce  # filter by command
```

## How It Works With Belayer

Belayer orchestrates pipelines (Intake → Implementation → Output). When a belayer gate node fails, the framework's failure handler calls `carabiner quality record`. When the implementation node starts, the framework script calls `carabiner quality check` and injects relevant patterns into the agent's prompt.

Belayer doesn't know about quality patterns. Carabiner doesn't know about pipelines. The framework (e.g., `claude-codex-carabiner`) is the glue that composes both.

## Philosophy

See [docs/PHILOSOPHY.md](docs/PHILOSOPHY.md) for the full reasoning behind carabiner's design: three-role separation (Agent/Harness/Orchestrator), H-as-feature for adversarial review, why quality patterns are not codebase documentation, and the false positive contamination problem.

## Status

Active development. Enforcement and Quality are MVP. Templates available for Go and React+TypeScript. Universal vigiles add-on for lint rule annotation validation.
