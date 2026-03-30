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

## Skill routing

When the user's request matches an available skill, ALWAYS invoke it using the Skill
tool as your FIRST action. Do NOT answer directly, do NOT use other tools first.
The skill has specialized workflows that produce better results than ad-hoc answers.

Key routing rules:
- Product ideas, "is this worth building", brainstorming → invoke office-hours
- Bugs, errors, "why is this broken", 500 errors → invoke investigate
- Ship, deploy, push, create PR → invoke ship
- QA, test the site, find bugs → invoke qa
- Code review, check my diff → invoke review
- Update docs after shipping → invoke document-release
- Weekly retro → invoke retro
- Design system, brand → invoke design-consultation
- Visual audit, design polish → invoke design-review
- Architecture review → invoke plan-eng-review
