# TODOs

## MVP: Quality Layer

### CLI Commands
- [ ] `carabiner init` — scaffold `.carabiner/` directory, walk user through setup, optionally write carabiner context to AGENTS.md (repo or local)
- [ ] `carabiner quality check --files <paths>` — retrieve relevant quality patterns. Path-prefix matching against learning paths. Return top N active learnings as markdown. Target: < 500ms (file I/O only, no model call).
- [ ] `carabiner quality record --gate-id <id>` — capture a learning from gate failure. Reads gate output (score + rationale), calls cheap model (haiku-class) to extract structured pattern, saves learning YAML + initial fail signal. `--skip-extraction` flag for environments without model CLI.
- [ ] `carabiner quality stats` — analytics. Per-learning: signal count, pass/fail ratio, last triggered. Computed from signal files at read time.

### Storage
- [ ] Learning YAML format: id (UUID), created, source_gate, status (computed, not stored), paths (directory prefixes), tags, pattern, recommendation
- [ ] Signal files: append-only, one per gate result. `{timestamp}-{branch}-{learning}-{result}.yaml`. Never mutated.
- [ ] `.carabiner/quality/learnings/` for pattern files
- [ ] `.carabiner/quality/signals/` for gate result signals
- [ ] `.carabiner/config.yaml` for settings

### Learning Extraction
- [ ] Extraction prompt spec: inputs (gate-result summary, rationale content, diff-stat file list), output (JSON with pattern, recommendation, paths as directory prefixes, tags), fallback (raw rationale if parsing fails)
- [ ] Model call via `claude -p --model haiku` (or configurable via `--model` flag)

### Agent Plugins (lightweight)
- [ ] Claude Code plugin — teaches Claude that `carabiner` exists and how to call quality check/record
- [ ] Codex skill — same for Codex
- [ ] OpenCode/Gemini equivalents as demand warrants

### Init Experience
- [ ] `carabiner init` creates `.carabiner/` with quality/ subdirs and config.yaml
- [ ] Asks: repo-level (`.carabiner/` committed) or local (`~/.carabiner/<repo-slug>/`)?
- [ ] Optionally writes carabiner usage instructions to AGENTS.md or .claude/CLAUDE.md

## Post-MVP: Quality Improvements

### Staleness and Decay
- [ ] Recurrence tracking: if gate passes and learning was relevant, track pass signals. After N consecutive passes, mark as dormant (pattern internalized).
- [ ] Age decay: learnings not triggered in 90 days marked stale, excluded from check.
- [ ] Codebase drift: if referenced paths no longer exist, mark orphaned.
- [ ] `carabiner quality compact` — roll up old signals into summaries on the learning file.

### Evolve (reduced scope)
- [ ] `carabiner evolve` — housekeeping mode. Prune stale patterns, flag ineffective ones (high recurrence despite learning existing), suggest removals. Does NOT auto-modify quality standards (descoped after false positive contamination analysis).
- [ ] Session launcher mode: user picks which agent runs the evolve session. Saved to config.yaml.
- [ ] Auto-apply as upgrade path: off by default, users opt in after seeing the system work.

### Bug Tracing (external signal for evolve)
- [ ] Tie bugs to specific design docs, branches, commits via git history
- [ ] When a bug fix lands for code that passed all gates, trace back to the original run
- [ ] This becomes the ground truth signal that gate failures alone can't provide
- [ ] Prerequisite for evaluator self-evolution (currently descoped)

## Future: Knowledge Layer

The knowledge layer is aspirational. It deserves its own brainstorm session. Open questions captured here for when we get there.

### What Is Knowledge
- Domain semantics agents can't derive from code (e.g., "total_spend subtracts CREDIT rows")
- Architectural invariants and conventions that look arbitrary without context
- Business logic rules that produce wrong results if an agent guesses
- Project philosophy and design principles

### Open Design Questions
- [ ] Format: structured YAML (like quality patterns) or freeform markdown that carabiner indexes?
- [ ] Retrieval: path+tag matching is enough for quality (file-scoped). Knowledge is domain-scoped. Need semantic matching? Embeddings? Or just good tagging?
- [ ] Bootstrap: can `carabiner init` extract knowledge from existing docs (README, ARCHITECTURE.md, inline comments)? One-time model call to structure what's already written?
- [ ] Authoring: who writes knowledge entries? Human only? Model-assisted? Extracted from PR review comments?
- [ ] Staleness: knowledge entries tied to specific code paths can be checked for drift. Domain rules ("CREDIT rows are subtracted") are true until schema changes. Different decay model than quality patterns.
- [ ] Relationship to AGENTS.md / CLAUDE.md: knowledge entries are richer and more structured than markdown agent rules. Do they replace those files, supplement them, or generate them?

## Backlog: Things From the Harness Plugin Worth Preserving

### PR Plugin
- [ ] Port pr:review (trimmed to 3 agents: code-reviewer, silent-failure-hunter, type-design-analyzer) to carabiner repo
- [ ] Port pr:author, pr:resolve, pr:update to carabiner repo
- [ ] Consider whether these should be CLI commands or stay as Claude Code plugins

### Doc Tier System
- [ ] The CLAUDE.md-as-map with domain docs and deep docs pattern was useful
- [ ] Consider whether `carabiner init` should scaffold this, or if it's out of scope
- [ ] The pruner agent (checking doc health, stale links, bloat) had value

### Self-Improving Runtime Design
- [ ] The `.harness/` design doc (2026-03-28) contains extensive research on self-improving agents
- [ ] Key references: Anthropic's "Harness design for long-running application development" (GAN pattern), Meta's HyperAgents paper
- [ ] Metrics schemas (review-effectiveness, plan-accuracy, learning-efficacy) designed but only partially implemented
- [ ] Codex proposed "harness evolution league" (replay/eval for proposals)
- [ ] Claude proposed "adversarial co-evolution" (fork harness, run both, score outcomes)
- [ ] Both deferred, both worth revisiting when quality layer has real usage data

## Reference: Key Decisions and Rationale

### Why separate from belayer
Belayer is the orchestrator (YAML pipelines, node execution, gate contracts). Carabiner is the harness (quality patterns, domain knowledge). Separate repos forces clean separation. You can use either without the other.

### Why CLI, not plugin
A Claude Code plugin only works in Claude Code. A CLI works from any agent that can run `sh -c`. Carabiner ships lightweight plugins as convenience wrappers, but the real work is in the binary.

### Why descope evaluator self-evolution
Gate failures are not ground truth. False positive contamination means auto-evolving standards from gate failures would encode hallucinations as institutional knowledge. Need external signal (bugs traced back to runs) before enabling self-evolution. See PHILOSOPHY.md for full analysis.

### Why 3 review agents, not 6
Data from 11 PRs across belayer and claude-remote-cli repos:
- silent-failure-hunter: 38 comments, ~90% actionable
- type-design-analyzer: 32 comments, ~80% actionable
- code-simplifier: 28 comments (cut — lower priority than bug/design catches)
- code-reviewer: 20 comments (kept — generalist, catches different axis)
- comment-analyzer: 19 comments (cut — low stakes)
- pr-test-analyzer: 14 comments (cut — lowest volume, often just "add more tests")

Cutting to 3 saves ~50% review tokens while keeping ~87% of actionable findings.

### H-as-Feature
Context loss between pipeline nodes is desirable for adversarial review. Isolated reviewers catch more issues because they lack implementation bias. This means no episode compression needed — each node bootstraps its own understanding. See PHILOSOPHY.md.

### The learning loop patterns from PR data
From silent-failure-hunter across both repos, 4 recurring patterns account for ~65% of findings:
1. Silent error swallowing via bare catch/fallback (12 occurrences)
2. Async fire-and-forget / discarded promises (5 occurrences)
3. JSON file corruption on non-atomic writes (4 occurrences)
4. Indistinguishable success/failure exit codes (4 occurrences)

These are cross-repo, cross-language patterns. Exactly the kind of quality learnings that should be injected into implementation prompts.
