# TODOs

## MVP: Join Layer (Current)

### CLI Commands
- [x] `carabiner why <file>:<line>` — forensic dossier for a line of code. Traces line → commit → git-ai note → agentlytics session. Reports confidence per hop.
- [x] `carabiner doctor` — diagnose available data sources (git, git-ai, agentlytics). Reports readiness with install instructions for missing sources.
- [x] `--json` flag for both commands
- [x] `--rev` flag for `carabiner why` (blame against specific revision)

### Data Source Readers
- [x] git-ai note parser (v3.0.0 format: attestation + JSON metadata, separated by `---`)
- [x] agentlytics reader (direct read-only query against `~/.agentlytics/cache.db`, `chats` table)
- [x] git blame parser (porcelain format with `-C` copy detection)
- [x] Typed git layer replacing old `gitValue()` (returns errors instead of swallowing them)

### Join Algorithm
- [x] Dossier builder: blame → git-ai note → session lookup → agentlytics enrichment
- [x] Confidence per hop: high (deterministic join) or missing (no data)
- [x] Overall confidence = weakest link
- [x] Graceful handling: no note, line not in attested range, session not in agentlytics

### Architecture
- [x] git-ai is a HARD REQUIREMENT (no fallback correlation)
- [x] agentlytics is recommended (enriches session data but git-ai alone gives high confidence)
- [x] Work-item linkage is a WORKFLOW RECOMMENDATION, not a code feature
- [x] Confidence model: high or missing only (no medium/low tiers)
- [x] agentlytics opened read-only with busy_timeout for concurrent safety

## Post-MVP: Approach B (Index + Query)

### `carabiner index`
- [ ] Build a local SQLite index joining git-ai notes + agentlytics data
- [ ] Incremental: only process new commits since last index
- [ ] Enables aggregate queries without re-reading external DBs each time

### `carabiner audit <work-item>`
- [ ] Given a work item (e.g., ENG-42), show all sessions that touched code for it
- [ ] Requires indexed data from `carabiner index`

### `carabiner report`
- [ ] Aggregate view: which agents wrote how much code, confidence distribution
- [ ] Session activity timeline
- [ ] Useful for engineering leads doing AI code audits

### Subsequent Touches
- [ ] `git log --follow` for files to show subsequent edits after initial attribution
- [ ] Limited to N commits (default 10, configurable via `--depth`)
- [ ] Each touch checked for its own git-ai attribution

## Post-MVP: Distribution & DX

### Shell Completions
- [ ] `carabiner completion bash/zsh/fish` (cobra built-in)

### VS Code Extension
- [ ] Inline attribution display (like GitLens for AI)
- [ ] Hover on a line to see session info
- [ ] Would reach significantly more users than CLI alone

### CI/CD Integration
- [ ] GitHub Action that runs `carabiner why` on changed lines in a PR
- [ ] Annotate PRs with AI attribution data

### One-Command Setup
- [ ] Install script that sets up carabiner + git-ai + agentlytics
- [ ] Demo repo with pre-populated git-ai notes for instant trial

## Future: Cross-Vendor Validation

- [ ] Validate join key works with OpenCode (`ses_*` prefix IDs)
- [ ] Validate with Codex sessions
- [ ] Validate with Gemini sessions
- [ ] Document any ID format differences across agents

## Reference: Key Decisions

### Why git-ai is a hard requirement
git-ai provides deterministic line-level attribution via Git Notes. Without it, attribution falls back to fuzzy timestamp correlation which risks false attribution. False attribution is worse than no attribution. Install git-ai first, carabiner joins on top.

### Why no fallback correlation
The design originally included timestamp + file Jaccard as a "medium confidence" fallback. This was dropped because: (1) false attribution risk in shared repos, (2) simpler confidence model (high or missing), (3) git-ai adoption is the right forcing function.

### Why work-item linkage is workflow, not code
Branch name parsing (`feature/ENG-42`) and commit message grepping are fragile. The right approach: use Linear/Jira's existing PR-to-ticket integration. The data flows through platforms, not through carabiner parsing heuristics. Setup docs recommend this workflow.

### Why CLI, not plugin
A plugin only works in one agent. A CLI works from any agent that can run `sh -c`. Universal interface.

### Why compose, don't compete
git-ai handles attribution. agentlytics handles collection. Carabiner handles the join. Each tool is independently maintained and solves its problem well.
