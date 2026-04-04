# Changelog

## 0.1.0 — Join Layer

Carabiner is now a forensic query layer for AI-coded repos. One command traces any line of code back to the agent session that wrote it.

### Added

- **`carabiner why <file>:<line>`**: Forensic dossier for a line of code. Traces line → commit (via git blame) → agent session (via git-ai notes) → session metadata (via agentlytics). Reports confidence per hop: high (deterministic join) or missing (no data).
- **`carabiner doctor`**: Diagnoses data source availability. Checks git repo, git-ai notes ref, and agentlytics cache. Reports readiness with install instructions for anything missing.
- **`--json` flag** on both commands for structured output
- **`--rev` flag** on `carabiner why` for blaming against a specific revision
- **git-ai note parser**: Reads v3.0.0 format (attestation + JSON metadata separated by `---`). Extracts agent_id, tool, model, and line ranges.
- **agentlytics reader**: Direct read-only query against `~/.agentlytics/cache.db` with busy_timeout for concurrent safety. Queries `chats` table by conversation ID.
- **Typed git layer**: `internal/carabiner/git/` package with `RunGit`, `Blame`, `ShowNote`, `HasNotesRef`. Returns errors instead of swallowing them.
- **Dossier builder**: `internal/carabiner/dossier/` assembles the full attribution chain with confidence per hop and weakest-link overall confidence.

### Removed

- Enforce layer (`carabiner enforce`)
- Quality layer (`carabiner quality check`, `carabiner quality record`)
- Work context layer (`carabiner context set/show/clear`)
- Templates (go, react-typescript)
- Validation layer (`carabiner validate`)
- Telemetry import layer (agentlytics session import)
- Init scaffolding (`carabiner init --template`)
- ~6,300 lines of code from the previous harness design

### Changed

- git-ai is now a hard requirement. No fallback correlation.
- Confidence model simplified to binary: high or missing. No medium/low tiers.
- agentlytics cache opened in read-only mode with busy_timeout.
- Work-item linkage is a workflow recommendation, not a code feature.
