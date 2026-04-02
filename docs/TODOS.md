# TODOs

## MVP: Quality Layer

### CLI Commands (3 commands — eng review reduced scope from 5 to 3)
- [ ] `carabiner init --mode repo|local` — scaffold `.carabiner/` directory. Flags only, no interactive prompts. Default mode=repo. Signals dir added to .gitignore.
- [ ] `carabiner quality check --files <paths>` — retrieve relevant quality patterns. Path-prefix matching against learning paths. Return top N active learnings as markdown. Target: < 500ms (file I/O only, no model call).
- [ ] `carabiner quality record --gate-id <id> [--gate-result <path>] [--skip-extraction]` — capture a learning from gate failure. Reads gate output via --gate-result JSON file or stdin. Calls cheap model to extract structured pattern. Saves learning YAML (with raw_input audit trail) + initial fail signal to JSONL. `--skip-extraction` for environments without model CLI.

### Storage (Hybrid: YAML Learnings + JSONL Signals)
- [ ] Learning YAML format: id (UUID), created, source, paths (directory prefixes), tags, pattern, recommendation, raw_input (audit trail), artifacts (branch, commit, design_doc)
- [ ] Signal JSONL: append-only log at `.carabiner/quality/signals/signals.jsonl`. Local-only (.gitignore'd). Learnings committed, signals per-machine.
- [ ] `.carabiner/quality/learnings/` for pattern files
- [ ] `.carabiner/quality/signals/` for JSONL signal log (gitignored)
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

### Signal Tracking + Stats (deferred from MVP by eng review)
- [ ] `carabiner quality signal --learning-id <id> --result <pass|fail> --gate-id <id>` — record pass/fail against existing learning. Enables signal accumulation over time.
- [ ] `carabiner quality stats` — analytics per learning: signal count, pass/fail ratio, last triggered, confidence. Computed from JSONL at read time.
- [ ] Signal count display in `quality check` output (e.g., "4 failures, 2 passes")
- [ ] Dormant learning filtering in `quality check` (requires signal history)
- [ ] Duplicate detection in `quality record` (fuzzy match on pattern text or same source + paths)

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

### Architecture Inspiration: Claude Code's Memory System

Claude Code's memory is a 3-layer bandwidth-aware design worth studying for our knowledge layer:

**Index = pointers, not storage.** MEMORY.md is always loaded (~150 chars/line) but only contains one-line hooks to topic files. Actual content lives in separate files, fetched on demand. This prevents context pollution — the model sees what exists without paying the token cost of everything stored.

**What they DON'T store is the real insight.** No debugging logs, no code structure, no PR history, no git-derivable facts. If it can be derived from the current state of the repo, it's not persisted. This is directly relevant to us: quality patterns capture what agents get WRONG, knowledge should capture what agents CAN'T DERIVE — not what they can grep for.

**Strict write discipline.** Write to file → then update index. Never dump content into the index. Two-step process prevents entropy. We should consider the same: knowledge entries are files, the index is generated/maintained separately.

**Staleness is first-class.** Memory that contradicts current reality is treated as wrong, not as historical context. Code-derived facts are never stored because they go stale. For our knowledge layer: domain rules ("CREDIT rows are subtracted") rarely go stale, but code-path-tied knowledge ("the auth middleware validates tokens in X") drifts constantly. Different staleness models for different knowledge types.

**Background consolidation (autoDream).** A forked subagent periodically merges, dedupes, removes contradictions, converts vague → absolute, aggressively prunes. Runs with limited tools to prevent corruption. We could do something similar: `carabiner knowledge compact` that re-evaluates entries against current codebase state.

**Retrieval is skeptical.** Memory is a hint, not truth. The model must verify before using. This aligns with our philosophy — knowledge entries should be treated as strong suggestions that the agent validates against the current code, not as gospel.

**Design questions this raises for carabiner:**
- [ ] Should we adopt the index-as-pointers pattern? A `knowledge/INDEX.md` that's always injected, with detailed entries in separate files?
- [ ] What's our equivalent of "don't store derivable facts"? Quality patterns have clear criteria (gate failures). Knowledge needs similar criteria for what's worth persisting vs. what the agent should just read from the code.
- [ ] Do we need a consolidation pass? Quality patterns are append-only with signal tracking. Knowledge might need periodic rewriting to stay coherent as the codebase evolves.
- [ ] How do we handle the bandwidth budget? If an agent calls `carabiner knowledge check`, how much context can we return before we're hurting more than helping?

### Architecture Inspiration: PageIndex (VectifyAI/PageIndex)

PageIndex is a "vectorless, reasoning-based RAG" framework. Instead of chunking documents and embedding them into a vector DB, it builds a hierarchical tree index (like a table of contents) and uses LLM reasoning to navigate to relevant sections. Claims 98.7% accuracy on FinanceBench, outperforming traditional vector RAG.

**Key ideas relevant to carabiner's knowledge layer:**

**Hierarchical tree index, not flat chunks.** Each node has a title, page range, LLM-generated summary, and children. This mirrors how a human expert navigates a document — top-down reasoning, not similarity search. For us: knowledge entries could form a tree (domain → subdomain → specific rule) rather than a flat list.

**No embeddings, no vector DB.** The tree index fits in the LLM's context window. The model reasons over summaries to decide which branches to descend into. Only then is the actual content fetched. This is the same bandwidth trick as Claude Code's memory — index is cheap, content is fetched on demand.

**Summaries as routing metadata.** Each node carries a natural-language summary that lets the LLM decide relevance without reading the full content. For carabiner: each knowledge entry could have a one-line summary used for retrieval, with the full rule/explanation fetched only when matched.

**Design questions this raises:**
- [ ] Could `carabiner knowledge check` build a lightweight tree index over knowledge entries (domain hierarchy + summaries) and let the calling agent reason over which entries to fetch in full?
- [ ] Is the tree structure overkill for our scale? Quality patterns use simple path-prefix matching. Knowledge might need something between flat-tag-matching and full tree search.
- [ ] PageIndex builds the index once (offline) and queries it many times. We could do the same: `carabiner knowledge index` rebuilds the tree, `carabiner knowledge check` queries it. The index is a cached artifact, not rebuilt per query.
- [ ] Could we use this approach for the quality layer too? As learnings accumulate, a hierarchical index (by domain/path/tag) might retrieve better than flat path-prefix matching.

### Synthesis: A 3-Layer Knowledge Architecture?

Combining both inspirations, a possible architecture:

```
Layer 1: Index (always injected, ~token-cheap)
  - One-line summaries per knowledge entry
  - Hierarchical: domain → subdomain → entry
  - Generated/cached by `carabiner knowledge index`

Layer 2: Knowledge entries (fetched on demand)
  - Full rules, context, rationale
  - YAML or markdown files in `.carabiner/knowledge/`
  - Each has: id, domain, paths, summary, content, last_verified

Layer 3: Source material (never fetched, only referenced)
  - Links to PRs, design docs, commit hashes where the knowledge originated
  - Audit trail, not retrieval content
```

**Retrieval flow:** Agent calls `carabiner knowledge check --context <paths|domain>` → carabiner returns the Layer 1 index (or relevant subtree) → agent reasons over summaries → agent requests specific entries by ID → carabiner returns Layer 2 content.

**Or simpler:** carabiner does the tree-search internally (path + tag + summary matching) and returns the top-N entries directly. Agent doesn't need to do multi-step retrieval. This is the MVP path — upgrade to agent-driven tree search later if flat matching proves insufficient.

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
