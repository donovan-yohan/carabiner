# carabiner

Forensic query layer for AI-coded repos. Joins git-ai (line attribution), agentlytics (session data), and work-item systems (Linear/Jira/GitHub) into a single query: `carabiner why <file>:<line>`. CLI-first, any agent that can run `sh -c` can use it.

## Quick Reference

| Action | Command |
|--------|---------|
| Build | `go build -o carabiner ./cmd/carabiner` |
| Test | `go test ./...` |
| Run | `./carabiner` |
| Why | `carabiner why <file>:<line>` |
| Doctor | `carabiner doctor` |

## Documentation Map

| Category | Path | When to look here |
|----------|------|-------------------|
| Philosophy | `docs/PHILOSOPHY.md` | Join-layer thesis, confidence model, compose-don't-compete, retroactive attribution |
| TODOs | `docs/TODOS.md` | MVP status, post-MVP features (index, audit, report), backlog |
| Origin | `docs/ORIGIN.md` | How carabiner emerged from the Slate analysis session, key references |
| Setup | `SETUP.md` | Prerequisites (git-ai, agentlytics), installation, usage guide |
| Changelog | `CHANGELOG.md` | Release history |
| Design (current) | `~/.gstack/projects/donovan-yohan-carabiner/donovanyohan-master-design-20260404-133209.md` | Join-layer architecture, proof-of-join gate, git-ai + agentlytics integration |
| Design (superseded) | `docs/design/2026-04-02-feed-forward-enforcement.md` | Old enforce layer design (superseded by join-layer pivot) |

## Key Patterns

- **Join layer**: Carabiner reads git-ai Git Notes, agentlytics cache.db, and git history. It doesn't collect data or write attribution. It connects them.
- **Confidence per hop**: Every step in the attribution chain (line->commit, commit->session, session->transcript) carries a confidence label: high (deterministic join) or missing (no data). Honest ambiguity over false certainty.
- **git-ai required**: git-ai is a hard requirement for `carabiner why`. No fallback correlation. Either the deterministic join works or it reports missing.
- **CLI-first**: the binary IS the interface. Lightweight agent plugins are convenience wrappers that call the CLI.
- **Compose, don't compete**: git-ai handles attribution. agentlytics handles collection. Carabiner handles the join. Each tool is independently maintained.
- **Separate from belayer**: carabiner is the forensic layer, belayer is the orchestrator. Either works alone.

## Relationship to Belayer

Belayer (separate repo: github.com/donovan-yohan/belayer) orchestrates YAML pipelines. Carabiner provides forensic context. A belayer framework can call `carabiner why` to enrich gate failure reports with session context, but neither requires the other.

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
