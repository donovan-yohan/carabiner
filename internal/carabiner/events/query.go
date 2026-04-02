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
