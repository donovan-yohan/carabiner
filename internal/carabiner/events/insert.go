package events

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type Event struct {
	ID           string
	Timestamp    time.Time
	Command      string
	Args         string
	ExitCode     int
	DurationMs   int64
	FilesTouched string
	RunID        string
	Branch       string
	Commit       string
	Metadata     string
}

type WorkContextEvent struct {
	ID          string
	Timestamp   time.Time
	WorkItemRef string
	SpecRef     string
	Branch      string
	Source      string
	Metadata    string
}

type WorkflowEvent struct {
	ID                string
	Timestamp         time.Time
	Workflow          string
	EventType         string
	ExternalSessionID string
	ExternalRunID     string
	RepoPath          string
	Branch            string
	CommitSHA         string
	Agent             string
	Model             string
	DurationMs        int64
	FailureCategory   string
	Metadata          string
}

type GitAttribution struct {
	CommitSHA      string
	WorkItemRef    string
	SpecRef        string
	Branch         string
	TrailerPayload string
	CreatedAt      time.Time
}

func AppendEvent(db *sql.DB, event *Event) error {
	query := `
		INSERT INTO events (id, timestamp, command, args, exit_code, duration_ms, files_touched, run_id, branch, "commit", metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var lastErr error
	for i := 0; i < 10; i++ {
		_, err := db.Exec(query,
			event.ID,
			event.Timestamp,
			event.Command,
			event.Args,
			event.ExitCode,
			event.DurationMs,
			event.FilesTouched,
			event.RunID,
			event.Branch,
			event.Commit,
			event.Metadata,
		)

		if err == nil {
			return nil
		}

		lastErr = err
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == sqlite3.SQLITE_BUSY {
				time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
				continue
			}
		}
		return fmt.Errorf("inserting event: %w", err)
	}

	return fmt.Errorf("inserting event after retries: %w", lastErr)
}

func AppendWorkContextEvent(db *sql.DB, event *WorkContextEvent) error {
	query := `
		INSERT INTO work_context_events (id, timestamp, work_item_ref, spec_ref, branch, source, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var lastErr error
	for i := 0; i < 10; i++ {
		_, err := db.Exec(query,
			event.ID,
			event.Timestamp,
			event.WorkItemRef,
			event.SpecRef,
			event.Branch,
			event.Source,
			event.Metadata,
		)

		if err == nil {
			return nil
		}

		lastErr = err
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == sqlite3.SQLITE_BUSY {
				time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
				continue
			}
		}
		return fmt.Errorf("inserting work context event: %w", err)
	}

	return fmt.Errorf("inserting work context event after retries: %w", lastErr)
}

func AppendWorkflowEvent(db *sql.DB, event *WorkflowEvent) error {
	query := `
		INSERT INTO workflow_events (id, timestamp, workflow, event_type, external_session_id, external_run_id, repo_path, branch, commit_sha, agent, model, duration_ms, failure_category, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var lastErr error
	for i := 0; i < 10; i++ {
		_, err := db.Exec(query,
			event.ID,
			event.Timestamp,
			event.Workflow,
			event.EventType,
			event.ExternalSessionID,
			event.ExternalRunID,
			event.RepoPath,
			event.Branch,
			event.CommitSHA,
			event.Agent,
			event.Model,
			event.DurationMs,
			event.FailureCategory,
			event.Metadata,
		)

		if err == nil {
			return nil
		}

		lastErr = err
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == sqlite3.SQLITE_BUSY {
				time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
				continue
			}
		}
		return fmt.Errorf("inserting workflow event: %w", err)
	}

	return fmt.Errorf("inserting workflow event after retries: %w", lastErr)
}

func UpsertGitAttribution(db *sql.DB, attribution *GitAttribution) error {
	query := `
		INSERT INTO git_attributions (commit_sha, work_item_ref, spec_ref, branch, trailer_payload, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(commit_sha) DO UPDATE SET
			work_item_ref = excluded.work_item_ref,
			spec_ref = excluded.spec_ref,
			branch = excluded.branch,
			trailer_payload = excluded.trailer_payload
	`

	var lastErr error
	for i := 0; i < 10; i++ {
		_, err := db.Exec(query,
			attribution.CommitSHA,
			attribution.WorkItemRef,
			attribution.SpecRef,
			attribution.Branch,
			attribution.TrailerPayload,
			attribution.CreatedAt,
		)

		if err == nil {
			return nil
		}

		lastErr = err
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == sqlite3.SQLITE_BUSY {
				time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
				continue
			}
		}
		return fmt.Errorf("upserting git attribution: %w", err)
	}

	return fmt.Errorf("upserting git attribution after retries: %w", lastErr)
}
