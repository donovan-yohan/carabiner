# carabiner

Agent-agnostic harness for coding agents. Two jobs: **quality** (patterns from review failures) and **knowledge** (domain semantics). CLI-first, any agent that can run `sh -c` can use it.

## Quick Reference

| Action | Command |
|--------|---------|
| Build | `go build -o carabiner ./cmd/carabiner` |
| Test | `go test ./...` |
| Run | `./carabiner` |
| Init | `carabiner init` |

## Documentation Map

| Category | Path | When to look here |
|----------|------|-------------------|
| Philosophy | `docs/PHILOSOPHY.md` | Three roles, H-as-feature, why CLI not plugin, false positive contamination |
| TODOs | `docs/TODOS.md` | MVP scope, post-MVP features, knowledge layer design questions, backlog |
| Origin | `docs/ORIGIN.md` | How carabiner emerged from the Slate analysis session, key references |

## Key Patterns

- **Quality patterns**: YAML files in `.carabiner/quality/learnings/`. Path-prefix matching for retrieval. Append-only signals for concurrent safety.
- **CLI-first**: the binary IS the interface. Lightweight agent plugins are convenience wrappers that call the CLI.
- **Separate from belayer**: carabiner is the harness, belayer is the orchestrator. Frameworks (shipped with belayer) compose both. Either works alone.
- **No evaluator self-evolution yet**: gate failures are not ground truth. Need bug-traced-to-run external signal before enabling auto-evolution of quality standards.

## Relationship to Belayer

Belayer (separate repo: github.com/donovan-yohan/belayer) orchestrates YAML pipelines. Carabiner provides quality/knowledge. Frameworks shipped with belayer (e.g., claude-codex-carabiner) call both CLIs. The gate contract is belayer's concern. What happens on failure (calling `carabiner quality record`) is the framework's decision.
