# Agent Validations — Design Doc

**Date:** 2026-04-02
**Status:** Approved for implementation

## Overview

Add a new `carabiner validate` command that forces agents to consciously consider reflection questions before committing. Unlike hard enforcement tools (eslint, svelte-check), validations are non-blocking forcing functions — they ask questions, record answers, and always pass. The value is in the record, not the gate.

## The Problem

Agents skip reflective tasks when not blocked. "Did you check the README?" isn't enforced, so it doesn't happen. Advisory CLAUDE.md suggestions can be ignored. We need a mechanism that:
1. Forces the question to be considered
2. Records whether the agent did the thing
3. Creates heuristics for when validations become noise

## Design

### Two Commands

```bash
# 1. Execute validations — creates pending records, prints questions
carabiner validate

# Output:
# [VALIDATION] README-impact
# Did you check if your changes affect any README sections?
# Run ID: abc123

# [VALIDATION] test-coverage
# Did you add or update tests for your changes?
# Run ID: abc123

# 2. Record answer — updates pending record
carabiner validate pass --name README-impact --run-id abc123
carabiner validate fail --name test-coverage --run-id abc123
carabiner validate irrelevant --name some-validation --run-id abc123
```

### enforce.yaml Schema

```yaml
agent_validations:
  - name: README-impact
    script: "echo 'Did you check if your changes affect any README sections?'"
    files: ["README.md", "docs/**/*"]

  - name: test-coverage
    script: "echo 'Did you add or update tests for your changes?'"
    files: ["src/**/*", "**/*.test.ts"]
```

### Data Model

```sql
CREATE TABLE validation_events (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    name TEXT NOT NULL,
    script TEXT NOT NULL,
    status TEXT NOT NULL,  -- pending|responded|orphaned
    result TEXT,           -- pass|fail|irrelevant (if responded)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    responded_at DATETIME,
    orphaned_at DATETIME
);

CREATE INDEX idx_validation_run ON validation_events(run_id);
CREATE INDEX idx_validation_name ON validation_events(name);
CREATE INDEX idx_validation_status ON validation_events(status);
```

### Orphaning Logic

When a new `carabiner validate` call runs:
1. Generate new run_id (UUID)
2. Execute all validations with this run_id
3. Mark all pending records from previous run_ids as orphaned:
   ```sql
   UPDATE validation_events 
   SET status = 'orphaned', orphaned_at = CURRENT_TIMESTAMP
   WHERE status = 'pending' AND run_id != ?
   ```

Concurrency: If two agents run validate simultaneously, orphaning is "last write wins" — fine for MVP. The goal is detecting skipped validations, not perfect sequencing.

### Stats

```bash
carabiner validate stats

# Output:
# VALIDATION         | PENDING | RESPONDED | ORPHANED | LAST_RUN
# README-impact     | 0       | 12        | 2        | 2m ago
# test-coverage     | 1       | 8         | 0        | just now
```

### Integration with Quality Layer

TODOs plan `carabiner quality signal` for recording pass/fail against learnings. Validations feed into the same signal model:
- Validation fails 3x in a row → suggest converting to hard check via vigiles
- Validation passes 10x with no changes → consider removing (noise)
- Validation orphaned often → prompt injection isn't working

This integration is deferred. MVP focuses on the orphaning + stats loop.

### Self-Reporting Limitation

Validations are forcing functions, not truth verification. The answer is only as honest as the agent's interpretation. Stats should be treated as leading indicators, not ground truth. Observe over time and adjust.

## File Structure

```
internal/carabiner/
  validate/
    execute.go      — ExecuteValidations()
    record.go       — RecordResult(), MarkOrphaned()
    stats.go       — ValidationStats()
  events/
    db.go          — ADD: CREATE TABLE validation_events
  cmd/
    validate.go     — CLI: validate [--all] + validate <result> --name <id> --run-id <id> + validate stats
```

## TODO

- [ ] Add `validation_events` table to events DB schema
- [ ] Implement `validate/execute.go` — ExecuteValidations()
- [ ] Implement `validate/record.go` — RecordResult(), MarkOrphaned()
- [ ] Implement `validate/stats.go` — ValidationStats()
- [ ] Add `validate` command to CLI
- [ ] Add tests for validate package
- [ ] Add agent_validations to shared Svelte templates (README-impact validation)
- [ ] Document self-reporting limitation in docs/
