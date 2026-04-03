package events

import (
	"database/sql"
	"fmt"
	"strings"
)

type EventFilter struct {
	Command string
	Branch  string
	RunID   string
	Limit   int
}

func ListEvents(db *sql.DB, filter *EventFilter) ([]Event, error) {
	query := `SELECT id, timestamp, command, args, exit_code, duration_ms, files_touched, run_id, branch, "commit", metadata FROM events`
	var conditions []string
	var args []any

	if filter.Command != "" {
		conditions = append(conditions, "command = ?")
		args = append(args, filter.Command)
	}
	if filter.Branch != "" {
		conditions = append(conditions, "branch = ?")
		args = append(args, filter.Branch)
	}
	if filter.RunID != "" {
		conditions = append(conditions, "run_id = ?")
		args = append(args, filter.RunID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Command,
			&e.Args,
			&e.ExitCode,
			&e.DurationMs,
			&e.FilesTouched,
			&e.RunID,
			&e.Branch,
			&e.Commit,
			&e.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating events: %w", err)
	}

	return events, nil
}

func ListWorkflowEvents(db *sql.DB, workflow string, limit int) ([]WorkflowEvent, error) {
	query := `SELECT id, timestamp, workflow, event_type, external_session_id, external_run_id, repo_path, branch, commit_sha, agent, model, duration_ms, failure_category, metadata FROM workflow_events`
	var conditions []string
	var args []any

	if workflow != "" {
		conditions = append(conditions, "workflow = ?")
		args = append(args, workflow)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying workflow events: %w", err)
	}
	defer rows.Close()

	var events []WorkflowEvent
	for rows.Next() {
		var e WorkflowEvent
		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Workflow,
			&e.EventType,
			&e.ExternalSessionID,
			&e.ExternalRunID,
			&e.RepoPath,
			&e.Branch,
			&e.CommitSHA,
			&e.Agent,
			&e.Model,
			&e.DurationMs,
			&e.FailureCategory,
			&e.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning workflow event: %w", err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating workflow events: %w", err)
	}

	return events, nil
}

func ListRecentAttributions(db *sql.DB, limit int) ([]GitAttribution, error) {
	query := `SELECT commit_sha, work_item_ref, spec_ref, branch, trailer_payload, created_at FROM git_attributions ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying git attributions: %w", err)
	}
	defer rows.Close()

	var attributions []GitAttribution
	for rows.Next() {
		var a GitAttribution
		err := rows.Scan(
			&a.CommitSHA,
			&a.WorkItemRef,
			&a.SpecRef,
			&a.Branch,
			&a.TrailerPayload,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning git attribution: %w", err)
		}
		attributions = append(attributions, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating git attributions: %w", err)
	}

	return attributions, nil
}

func ListWorkContextEvents(db *sql.DB, limit int) ([]WorkContextEvent, error) {
	query := `SELECT id, timestamp, work_item_ref, spec_ref, branch, source, metadata FROM work_context_events ORDER BY timestamp DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying work context events: %w", err)
	}
	defer rows.Close()

	var events []WorkContextEvent
	for rows.Next() {
		var e WorkContextEvent
		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.WorkItemRef,
			&e.SpecRef,
			&e.Branch,
			&e.Source,
			&e.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning work context event: %w", err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating work context events: %w", err)
	}

	return events, nil
}
