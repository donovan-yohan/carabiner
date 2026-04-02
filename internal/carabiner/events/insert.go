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
