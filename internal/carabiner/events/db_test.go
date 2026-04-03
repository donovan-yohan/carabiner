package events

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB_CreatesFile(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "carabiner-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Verify file doesn't exist before
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatalf("Database file should not exist before InitDB")
	}

	// Call InitDB
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Verify file exists after
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file should exist after InitDB")
	}
}

func TestInitDB_WALMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Check journal_mode is WAL
	var mode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}

	if mode != "wal" {
		t.Errorf("Expected journal_mode 'wal', got '%s'", mode)
	}
}

func TestInitDB_BusyTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Check busy_timeout is 5000ms (5 seconds)
	var timeout int
	err = db.QueryRow("PRAGMA busy_timeout").Scan(&timeout)
	if err != nil {
		t.Fatalf("Failed to query busy_timeout: %v", err)
	}

	if timeout != 5000 {
		t.Errorf("Expected busy_timeout 5000ms, got %dms", timeout)
	}
}

func TestInitDB_CreatesTable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Verify events table exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='events'").Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Errorf("Events table should exist")
		} else {
			t.Fatalf("Failed to query sqlite_master: %v", err)
		}
	}

	if name != "events" {
		t.Errorf("Expected table name 'events', got '%s'", name)
	}

	// Verify indexes exist
	indexes := []string{
		"idx_events_timestamp",
		"idx_events_run_id",
		"idx_events_command",
	}

	for _, idx := range indexes {
		var idxName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&idxName)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("Index %s should exist", idx)
			} else {
				t.Fatalf("Failed to query index %s: %v", idx, err)
			}
		}
	}
}

func TestInitDB_Idempotent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// First call
	db1, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("First InitDB failed: %v", err)
	}
	db1.Close()

	// Second call - should not error
	db2, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Second InitDB failed: %v", err)
	}
	defer db2.Close()

	// Verify table still exists
	var name string
	err = db2.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='events'").Scan(&name)
	if err != nil {
		t.Errorf("Events table should still exist after second InitDB: %v", err)
	}
}

func TestInitDB_CreatesValidationEventsTable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='validation_events'").Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Errorf("validation_events table should exist")
		} else {
			t.Fatalf("Failed to query sqlite_master: %v", err)
		}
	}

	if name != "validation_events" {
		t.Errorf("Expected table name 'validation_events', got '%s'", name)
	}

	indexes := []string{
		"idx_validation_run",
		"idx_validation_name",
		"idx_validation_status",
	}

	for _, idx := range indexes {
		var idxName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&idxName)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("Index %s should exist", idx)
			} else {
				t.Fatalf("Failed to query index %s: %v", idx, err)
			}
		}
	}
}
