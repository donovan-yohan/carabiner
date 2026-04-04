# Workflow Observability MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reframe carabiner from an enforcement wrapper into a repo-local workflow observability layer by adding branch-scoped context, commit-time attribution, agentlytics ingestion, and raw data query/export surfaces.

**Architecture:** Carabiner owns repo-local attribution and reflection, not the underlying enforcement toolchain. Work context is stored in git worktree-local config and enforced by generated hooks. Existing CLI events remain in `.carabiner/carabiner.db`, which is extended with workflow telemetry tables populated from agentlytics' local SQLite cache. A minimal raw query/export surface ships before higher-level `doctor`/`health` commands.

**Tech Stack:** Go (cobra CLI, modernc SQLite), git hooks, git worktree-local config, agentlytics local SQLite cache at `~/.agentlytics/cache.db`

---

## Eng Review Deltas (2026-04-03)

Accepted changes from `/plan-eng-review`:

- **Keep full MVP scope** as the next experiment, but tighten risky edges rather than reducing scope.
- **agentlytics integration stays SQLite-first**, but all upstream schema knowledge must live behind one importer adapter boundary. Preserve raw upstream metadata.
- **Context validity is branch-scoped plus freshness-aware**, not branch-only. A stale context on the same branch must fail validation.
- **Do not auto-install git hooks in `carabiner init`.** Replace that with repo onboarding docs (`README.md` + `SETUP.md`) that tell agents how to install/adapt the hooks themselves.
- **Raw SQL query stays**, but it must be clearly labeled **experimental/internal** so it does not accidentally become the long-term public contract.
- **agentlytics rows should land in a raw import table first**, not be eagerly normalized into `workflow_events` with guessed workflow names.
- **Add a real git E2E integration test** covering branch switch, stale context rejection, and trailer injection.
- **Add an agentlytics schema-drift fixture test** so upstream cache changes fail explicitly.
- **Add import watermark / upsert behavior** so repeated imports do not rescan and duplicate forever.

These deltas override any conflicting details later in this plan.

---

## File Structure

```text
internal/cmd/
  init.go                         # MODIFY: advertise new context + hook scaffolding during init
  context.go                      # CREATE: `carabiner context set|show|clear`
  telemetry.go                    # CREATE: `carabiner telemetry import agentlytics`
  data.go                         # CREATE: `carabiner data export` + `carabiner data query`
  events.go                       # MODIFY: optionally expose new workflow-oriented list filters or keep legacy intact

internal/carabiner/
  init.go                         # MODIFY: scaffold AGENTS.md/CLAUDE.md instructions and git hooks
  context.go                      # CREATE: worktree-local context read/write/validate helpers
  context_test.go                 # CREATE
  hooks.go                        # CREATE: render pre-commit + commit-msg hook scripts
  hooks_test.go                   # CREATE

internal/carabiner/events/
  db.go                           # MODIFY: add workflow/context/telemetry tables and indexes
  query.go                        # MODIFY or extend for raw query/export support
  insert.go                       # MODIFY or extend with insert helpers for new tables
  workflow.go                     # CREATE: typed helpers for workflow events / imports / attribution rows
  workflow_test.go                # CREATE

internal/carabiner/telemetry/
  agentlytics.go                  # CREATE: read agentlytics cache.db and normalize upstream rows
  agentlytics_test.go             # CREATE

docs/superpowers/plans/
  2026-04-03-workflow-observability-mvp-plan.md   # this plan
```

**External contract details to preserve:**
- agentlytics advertises a local cache at `~/.agentlytics/cache.db`
- relay data lives separately at `~/.agentlytics/relay.db`
- the Deno edition bypasses SQLite, so MVP should only target the Node/SQLite cache path
- agentlytics exposes read-only REST endpoints, but MVP should consume SQLite directly to stay local-first and avoid server startup dependency

---

## MVP Product Contract

### User-visible commands

```bash
carabiner context set --work-item <ref> [--spec <ref>]
carabiner context show
carabiner context clear

carabiner telemetry import agentlytics [--source <path>]

carabiner data export [--format json]
carabiner data query --sql <query>
```

### Git hook behavior

- `pre-commit`
  - fail if work context is missing
  - fail if stored context branch does not match current branch
  - print exact remediation command
- `commit-msg`
  - append trailers automatically from current valid context
  - fail if context missing/invalid

### Commit trailers

```text
Carabiner-Work-Item: linear/ENG-42
Carabiner-Spec: doc:docs/designs/workflow-health.md
Carabiner-Context-Branch: feature/workflow-health
```

### Context validity rules

Context is valid iff:
- `workItemRef` exists
- `contextBranch` equals current `git rev-parse --abbrev-ref HEAD`

Stored values:

```text
git config --worktree carabiner.workItemRef "linear/ENG-42"
git config --worktree carabiner.specRef "doc:docs/designs/workflow-health.md"
git config --worktree carabiner.contextBranch "feature/workflow-health"
git config --worktree carabiner.contextSetAt "2026-04-03T12:30:00Z"
git config --worktree carabiner.contextSource "explicit"
```

---

## Task 1: Add branch-scoped context commands

**Files:**
- Create: `internal/cmd/context.go`
- Create: `internal/carabiner/context.go`
- Create: `internal/carabiner/context_test.go`

- [ ] **Step 1: Define the context state model**

Add this type to `internal/carabiner/context.go`:

```go
package carabiner

type WorkContext struct {
	WorkItemRef  string
	SpecRef      string
	ContextBranch string
	SetAt        string
	Source       string
}
```

- [ ] **Step 2: Add worktree-local config helpers**

Implement helpers using `git config --worktree`:

```go
func SetWorkContext(ctx WorkContext) error
func GetWorkContext() (WorkContext, error)
func ClearWorkContext() error
func CurrentBranch() string
func ValidateWorkContext(ctx WorkContext) error
```

Validation errors should be explicit:
- missing work item
- missing branch
- branch mismatch

- [ ] **Step 3: Expose `carabiner context` subcommands**

Create `internal/cmd/context.go` with Cobra commands:

```go
var contextCmd = &cobra.Command{Use: "context", Short: "Manage branch-scoped work context"}
var contextSetCmd = &cobra.Command{Use: "set", RunE: ...}
var contextShowCmd = &cobra.Command{Use: "show", RunE: ...}
var contextClearCmd = &cobra.Command{Use: "clear", RunE: ...}
```

Flags for `context set`:
- `--work-item` required
- `--spec` optional

- [ ] **Step 4: Write tests for valid and stale context**

Test cases:
- set + get round trip
- clear removes all keys
- validate passes on matching branch
- validate fails on branch mismatch
- validate fails on missing work item

Suggested test shape:

```go
func TestValidateWorkContext_BranchMismatch(t *testing.T) {
	ctx := WorkContext{WorkItemRef: "linear/ENG-42", ContextBranch: "feature/a"}
	// stub current branch to feature/b
	// expect mismatch error
}
```

- [ ] **Step 5: Run tests**

Run:

```bash
go test ./internal/carabiner ./internal/cmd
```

Expected: passing tests, no compile errors

---

## Task 2: Scaffold hooks and repo instructions during init

**Files:**
- Modify: `internal/carabiner/init.go`
- Modify: `internal/cmd/init.go`
- Create: `internal/carabiner/hooks.go`
- Create: `internal/carabiner/hooks_test.go`

- [ ] **Step 1: Add hook renderers**

Create `internal/carabiner/hooks.go` with two functions:

```go
func RenderPreCommitHook() string
func RenderCommitMsgHook() string
```

`pre-commit` shell script must do:

```sh
#!/bin/sh
carabiner context show >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "carabiner: no valid work context for current branch"
  echo "Run: carabiner context set --work-item <ref> [--spec <ref>]"
  exit 1
fi
```

`commit-msg` script must do:
- read current valid context via `carabiner context show --format json` or a dedicated machine-readable mode if added
- append trailers if missing
- reject commit if context invalid

- [ ] **Step 2: Add AGENTS.md and CLAUDE.md instruction blocks**

Modify `internal/carabiner/init.go` to write or append instructions like:

```md
Before meaningful implementation work, run:

`carabiner context set --work-item <ref> [--spec <ref>]`

Before spawning subagents, pass `workItemRef` and `specRef`.
Do not commit without valid carabiner context for the current branch.
```

Prefer:
- prepend a small section to `CLAUDE.md` if present
- create `AGENTS.md` if absent

- [ ] **Step 3: Install hooks during init**

Modify `Init()` in `internal/carabiner/init.go` so repo mode creates:

```text
.git/hooks/pre-commit
.git/hooks/commit-msg
```

and marks them executable.

- [ ] **Step 4: Add init tests**

Cover:
- hooks are written
- hooks are executable
- instruction text is added
- existing files are not clobbered destructively

- [ ] **Step 5: Run tests**

Run:

```bash
go test ./internal/carabiner ./internal/cmd
```

---

## Task 3: Extend the SQLite schema for attribution and workflow telemetry

**Files:**
- Modify: `internal/carabiner/events/db.go`
- Modify: `internal/carabiner/events/insert.go`
- Modify: `internal/carabiner/events/query.go`
- Create: `internal/carabiner/events/workflow.go`
- Create: `internal/carabiner/events/workflow_test.go`

- [ ] **Step 1: Add new tables**

Extend `db.go` with these tables:

```sql
CREATE TABLE IF NOT EXISTS work_context_events (
  id TEXT PRIMARY KEY,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  work_item_ref TEXT NOT NULL,
  spec_ref TEXT,
  branch TEXT NOT NULL,
  source TEXT NOT NULL,
  metadata TEXT
);

CREATE TABLE IF NOT EXISTS workflow_events (
  id TEXT PRIMARY KEY,
  timestamp DATETIME NOT NULL,
  workflow TEXT NOT NULL,
  event_type TEXT NOT NULL,
  external_session_id TEXT,
  external_run_id TEXT,
  repo_path TEXT,
  branch TEXT,
  commit_sha TEXT,
  agent TEXT,
  model TEXT,
  duration_ms INTEGER,
  failure_category TEXT,
  metadata TEXT
);

CREATE TABLE IF NOT EXISTS git_attributions (
  commit_sha TEXT PRIMARY KEY,
  work_item_ref TEXT NOT NULL,
  spec_ref TEXT,
  branch TEXT NOT NULL,
  trailer_payload TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Add indexes on timestamp, workflow, branch, external_session_id, failure_category.

- [ ] **Step 2: Add insert helpers**

Create typed helpers in `workflow.go`:

```go
type WorkflowEvent struct { ... }
type WorkContextEvent struct { ... }
type GitAttribution struct { ... }

func AppendWorkflowEvent(db *sql.DB, event *WorkflowEvent) error
func AppendWorkContextEvent(db *sql.DB, event *WorkContextEvent) error
func UpsertGitAttribution(db *sql.DB, attribution *GitAttribution) error
```

- [ ] **Step 3: Add query helpers for raw access**

Extend `query.go` or add adjacent helpers for:

```go
func ListWorkflowEvents(db *sql.DB, workflow string, limit int) ([]WorkflowEvent, error)
func ListRecentAttributions(db *sql.DB, limit int) ([]GitAttribution, error)
```

- [ ] **Step 4: Write schema and query tests**

Cover:
- new tables exist
- insert helpers persist rows
- list helpers filter and order correctly

- [ ] **Step 5: Run tests**

Run:

```bash
go test ./internal/carabiner/events
```

---

## Task 4: Implement agentlytics ingestion

**Files:**
- Create: `internal/carabiner/telemetry/agentlytics.go`
- Create: `internal/carabiner/telemetry/agentlytics_test.go`
- Create: `internal/cmd/telemetry.go`

- [ ] **Step 1: Define agentlytics importer contract**

Create a source struct:

```go
type AgentlyticsImportOptions struct {
	SourcePath string
	Limit      int
}

func DefaultAgentlyticsCachePath() string
func ImportAgentlytics(db *sql.DB, opts AgentlyticsImportOptions) (int, error)
```

Default path:

```go
filepath.Join(os.Getenv("HOME"), ".agentlytics", "cache.db")
```

- [ ] **Step 2: Read upstream data conservatively**

Do **not** assume the full upstream schema. First inspect available tables at runtime:

```sql
SELECT name FROM sqlite_master WHERE type='table';
```

Then only ingest the minimum fields we need from stable, discoverable columns:
- session/chat identifier
- timestamp or created_at
- editor/agent source
- repo/project path if present
- model if present
- tool/message counts if present

Store anything uncertain in `metadata` JSON.

- [ ] **Step 3: Normalize upstream rows into `workflow_events`**

Map imported sessions into coarse event rows. For MVP, use imported session boundaries as raw workflow events rather than pretending to have perfect started/succeeded/failed lifecycle data.

Example normalization:

```go
WorkflowEvent{
	ID: fmt.Sprintf("agentlytics:%s", upstreamID),
	Workflow: inferWorkflowName(projectPath, editorName),
	EventType: "imported_session",
	ExternalSessionID: upstreamID,
	RepoPath: projectPath,
	Agent: editorName,
	Model: model,
	Metadata: rawJSON,
}
```

- [ ] **Step 4: Expose CLI import command**

Add `internal/cmd/telemetry.go`:

```go
var telemetryCmd = &cobra.Command{Use: "telemetry"}
var telemetryImportCmd = &cobra.Command{Use: "import"}
var telemetryImportAgentlyticsCmd = &cobra.Command{Use: "agentlytics"}
```

CLI:

```bash
carabiner telemetry import agentlytics [--source ~/.agentlytics/cache.db]
```

- [ ] **Step 5: Write importer tests with a fixture SQLite db**

Build a tiny temporary SQLite fixture in test setup and verify:
- importer opens it
- rows normalize successfully
- duplicate imports do not create duplicate workflow rows

- [ ] **Step 6: Run tests**

Run:

```bash
go test ./internal/carabiner/telemetry ./internal/cmd
```

---

## Task 5: Add raw data export/query surfaces before `doctor`

**Files:**
- Create: `internal/cmd/data.go`
- Modify: `internal/carabiner/events/query.go` or supporting helpers

- [ ] **Step 1: Add `carabiner data export`**

CLI contract:

```bash
carabiner data export --format json
```

Return a JSON document containing:
- recent `workflow_events`
- recent `git_attributions`
- recent `work_context_events`

- [ ] **Step 2: Add `carabiner data query --sql`**

CLI contract:

```bash
carabiner data query --sql "SELECT workflow, COUNT(*) FROM workflow_events GROUP BY workflow"
```

Constrain it to read-only SELECT queries only.

- [ ] **Step 3: Add tests for export and read-only SQL enforcement**

Cover:
- JSON export includes all top-level sections
- non-SELECT queries are rejected
- valid SELECT queries return tabular or JSON output consistently

- [ ] **Step 4: Run tests**

Run:

```bash
go test ./internal/cmd ./internal/carabiner/events
```

---

## Task 6: Persist context and attribution into runtime logs

**Files:**
- Modify: `internal/cmd/root.go`
- Modify: `internal/carabiner/events/insert.go`

- [ ] **Step 1: Extend root command event metadata**

When logging normal CLI events in `root.go`, include current work context if available in `Metadata`.

Add fields to metadata JSON:

```json
{
  "os": "darwin",
  "arch": "arm64",
  "pid": 123,
  "workItemRef": "linear/ENG-42",
  "specRef": "doc:docs/designs/workflow-health.md",
  "contextBranch": "feature/workflow-health"
}
```

- [ ] **Step 2: Record work-context changes as events**

Whenever `context set` or `context clear` runs, append a `work_context_event` row.

- [ ] **Step 3: Add tests**

Cover:
- root metadata includes context when set
- work-context events are emitted on set/clear

- [ ] **Step 4: Run tests**

Run:

```bash
go test ./internal/cmd ./internal/carabiner/events
```

---

## Task 7: Final integration and manual verification

**Files:**
- Modify: any touched files above
- Test: existing command integration tests plus new targeted tests

- [ ] **Step 1: Run full test suite**

Run:

```bash
go test ./...
```

Expected: all tests pass

- [ ] **Step 2: Run manual CLI QA for context commands**

In a temporary repo fixture:

```bash
carabiner context set --work-item local:test --spec doc:docs/designs/test.md
carabiner context show
carabiner context clear
```

Expected:
- show prints current branch + refs
- clear removes them

- [ ] **Step 3: Run manual hook QA**

Scenario:
- set context on branch A
- switch to branch B
- attempt commit

Expected:
- pre-commit or commit-msg rejects due to stale branch-scoped context

- [ ] **Step 4: Run manual agentlytics import QA**

Run:

```bash
carabiner telemetry import agentlytics
carabiner data export --format json
```

Expected:
- import succeeds or fails with explicit source-path/schema error
- exported JSON contains imported rows if source db exists

- [ ] **Step 5: Commit in focused slices**

Suggested commit order:

```bash
git commit -m "feat(context): add branch-scoped work context commands"
git commit -m "feat(init): scaffold context hooks and repo instructions"
git commit -m "feat(events): add workflow telemetry and attribution tables"
git commit -m "feat(telemetry): import agentlytics into local workflow history"
git commit -m "feat(data): add raw export and query surfaces"
```

---

## Self-Review Notes

- This plan intentionally keeps `doctor`, `health`, and `workflows` **out** of MVP implementation even though they are the product direction. The raw query/export layer ships first so real data and real user questions can shape those commands.
- This plan uses `carabiner context set` instead of raw `git config` commands so branch stamping and validation logic stay centralized.
- This plan pulls directly from agentlytics' documented local SQLite cache (`~/.agentlytics/cache.db`) and does not depend on its Express server or relay mode for MVP.
- This plan keeps the diff “engineered enough” by reusing the existing SQLite db and command/query patterns already present in carabiner.
