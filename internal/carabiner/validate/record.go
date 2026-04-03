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

// SQL queries as constants for consistency between code and tests
const (
	insertPendingQuery   = `INSERT INTO validation_events (id, run_id, name, script, status, created_at) VALUES (?, ?, ?, ?, 'pending', ?)`
	updateResultQuery    = `UPDATE validation_events SET status = 'responded', result = ?, responded_at = ? WHERE name = ? AND run_id = ? AND status = 'pending'`
	markOrphanedQuery    = `UPDATE validation_events SET status = 'orphaned', orphaned_at = ? WHERE status = 'pending' AND run_id != ?`
	validationStatsQuery = `SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`
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
	_, err := db.Exec(insertPendingQuery, event.ID, event.RunID, event.Name, event.Script, event.CreatedAt)
	return err
}

// RecordResult updates a pending validation event with the agent's response.
func RecordResult(db *sql.DB, name, runID, result string) error {
	res, err := db.Exec(updateResultQuery, result, time.Now(), name, runID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected for validation result update: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no pending validation found for name=%s run_id=%s", name, runID)
	}
	return nil
}

// MarkOrphaned marks all pending validation events from previous runs as orphaned.
func MarkOrphaned(db *sql.DB, currentRunID string) error {
	_, err := db.Exec(markOrphanedQuery, time.Now(), currentRunID)
	return err
}
