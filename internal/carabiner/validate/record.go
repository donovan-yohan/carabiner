package validate

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	StatusPending   = "pending"
	StatusResponded = "responded"
	StatusOrphaned  = "orphaned"
)

const (
	ResultPass       = "pass"
	ResultFail       = "fail"
	ResultIrrelevant = "irrelevant"
)

type ValidationEvent struct {
	ID          string
	RunID       string
	Name        string
	Script      string
	Status      string
	Result      string
	CreatedAt   time.Time
	RespondedAt *time.Time
	OrphanedAt  *time.Time
}

// InsertPending inserts a new pending validation event.
func InsertPending(db *sql.DB, event *ValidationEvent) error {
	query := `INSERT INTO validation_events (id, run_id, name, script, status, created_at)
               VALUES (?, ?, ?, ?, 'pending', ?)`
	_, err := db.Exec(query, event.ID, event.RunID, event.Name, event.Script, event.CreatedAt)
	return err
}

// RecordResult updates a pending validation event with the agent's response.
func RecordResult(db *sql.DB, name, runID, result string) error {
	query := `UPDATE validation_events 
               SET status = 'responded', result = ?, responded_at = ?
               WHERE name = ? AND run_id = ? AND status = 'pending'`
	res, err := db.Exec(query, result, time.Now(), name, runID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("no pending validation found for name=%s run_id=%s", name, runID)
	}
	return nil
}

// MarkOrphaned marks all pending validation events from previous runs as orphaned.
func MarkOrphaned(db *sql.DB, currentRunID string) error {
	query := `UPDATE validation_events 
               SET status = 'orphaned', orphaned_at = ?
               WHERE status = 'pending' AND run_id != ?`
	_, err := db.Exec(query, time.Now(), currentRunID)
	return err
}
