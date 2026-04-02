# Changelog

## 0.0.1

### Added

- **Enforce Layer**: `carabiner enforce --all` runs configured static analysis tools with deterministic pass/fail
  - 3-state exit codes: 0=pass, 1=enforcement fail, 2=config error
  - Tools: golangci-lint, gofmt, staticcheck, gotestsum
  - Sequential execution with stop-on-first-fail option
  - `carabiner enforce --tool <name>` for single tool execution

- **Event Log Layer**: SQLite-backed workflow observability
  - Auto-logs every carabiner invocation
  - `carabiner events list` with filtering by command, branch, run-id
  - `carabiner events init` for database initialization

- **Templates**: Project templates for tool configuration
  - React+TypeScript: eslint, tsc, prettier, vitest
  - Go: golangci-lint, gofmt, staticcheck, gotestsum

- **Quality Layer**: Existing quality patterns system
  - `carabiner quality check --files <paths>` for pattern retrieval
  - `carabiner quality record --gate-id <id>` for capturing learnings

- **Init**: `carabiner init --mode repo|local` scaffolds .carabiner/ directory
