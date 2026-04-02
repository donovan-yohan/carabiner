package events

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestAppendEvent_Success(t *testing.T) {
	// Setup: create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Create test event
	now := time.Now().UTC().Truncate(time.Millisecond)
	event := &Event{
		ID:           "test-id-1",
		Timestamp:    now,
		Command:      "init",
		Args:         `{"arg1": "value1"}`,
		ExitCode:     0,
		DurationMs:   1234,
		FilesTouched: `["file1.go", "file2.go"]`,
		RunID:        "run-123",
		Branch:       "main",
		Commit:       "abc123",
		Metadata:     `{"key": "value"}`,
	}

	// Execute: append event
	err = AppendEvent(db, event)
	if err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	// Verify: query back and check fields
	var retrieved Event
	query := `SELECT id, timestamp, command, args, exit_code, duration_ms, files_touched, run_id, branch, "commit", metadata FROM events WHERE id = ?`
	err = db.QueryRow(query, event.ID).Scan(
		&retrieved.ID,
		&retrieved.Timestamp,
		&retrieved.Command,
		&retrieved.Args,
		&retrieved.ExitCode,
		&retrieved.DurationMs,
		&retrieved.FilesTouched,
		&retrieved.RunID,
		&retrieved.Branch,
		&retrieved.Commit,
		&retrieved.Metadata,
	)
	if err != nil {
		t.Fatalf("Failed to query event: %v", err)
	}

	// Verify all fields match
	if retrieved.ID != event.ID {
		t.Errorf("ID mismatch: got %q, want %q", retrieved.ID, event.ID)
	}
	if !retrieved.Timestamp.Equal(event.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", retrieved.Timestamp, event.Timestamp)
	}
	if retrieved.Command != event.Command {
		t.Errorf("Command mismatch: got %q, want %q", retrieved.Command, event.Command)
	}
	if retrieved.Args != event.Args {
		t.Errorf("Args mismatch: got %q, want %q", retrieved.Args, event.Args)
	}
	if retrieved.ExitCode != event.ExitCode {
		t.Errorf("ExitCode mismatch: got %d, want %d", retrieved.ExitCode, event.ExitCode)
	}
	if retrieved.DurationMs != event.DurationMs {
		t.Errorf("DurationMs mismatch: got %d, want %d", retrieved.DurationMs, event.DurationMs)
	}
	if retrieved.FilesTouched != event.FilesTouched {
		t.Errorf("FilesTouched mismatch: got %q, want %q", retrieved.FilesTouched, event.FilesTouched)
	}
	if retrieved.RunID != event.RunID {
		t.Errorf("RunID mismatch: got %q, want %q", retrieved.RunID, event.RunID)
	}
	if retrieved.Branch != event.Branch {
		t.Errorf("Branch mismatch: got %q, want %q", retrieved.Branch, event.Branch)
	}
	if retrieved.Commit != event.Commit {
		t.Errorf("Commit mismatch: got %q, want %q", retrieved.Commit, event.Commit)
	}
	if retrieved.Metadata != event.Metadata {
		t.Errorf("Metadata mismatch: got %q, want %q", retrieved.Metadata, event.Metadata)
	}
}

func TestAppendEvent_Concurrent(t *testing.T) {
	// Setup: create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Execute: 50 concurrent inserts
	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event := &Event{
				ID:           "concurrent-id-" + string(rune('A'+idx%26)) + string(rune('0'+idx%10)),
				Timestamp:    time.Now().UTC(),
				Command:      "concurrent-test",
				ExitCode:     0,
				DurationMs:   int64(idx),
				FilesTouched: "[]",
				Metadata:     "{}",
			}
			if err := AppendEvent(db, event); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Verify: all inserts succeeded
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}
	if len(errorList) > 0 {
		t.Errorf("Concurrent inserts failed with %d errors:", len(errorList))
		for _, err := range errorList {
			t.Errorf("  - %v", err)
		}
	}

	// Verify: count all rows
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count events: %v", err)
	}
	if count != numGoroutines {
		t.Errorf("Row count mismatch: got %d, want %d", count, numGoroutines)
	}
}

func TestAppendEvent_AllFields(t *testing.T) {
	// Setup: create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Create event with all fields populated
	now := time.Now().UTC().Truncate(time.Millisecond)
	args := map[string]string{"arg1": "value1", "arg2": "value2"}
	argsJSON, _ := json.Marshal(args)
	filesTouched := []string{"file1.go", "file2.go", "file3.go"}
	filesJSON, _ := json.Marshal(filesTouched)
	metadata := map[string]interface{}{"key1": "value1", "key2": 123, "nested": map[string]string{"a": "b"}}
	metadataJSON, _ := json.Marshal(metadata)

	event := &Event{
		ID:           "full-event-id",
		Timestamp:    now,
		Command:      "full-command",
		Args:         string(argsJSON),
		ExitCode:     42,
		DurationMs:   5678,
		FilesTouched: string(filesJSON),
		RunID:        "run-full-123",
		Branch:       "feature-branch",
		Commit:       "def456",
		Metadata:     string(metadataJSON),
	}

	// Execute: append event
	err = AppendEvent(db, event)
	if err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	// Verify: retrieve and check all fields
	var retrieved Event
	query := `SELECT id, timestamp, command, args, exit_code, duration_ms, files_touched, run_id, branch, "commit", metadata FROM events WHERE id = ?`
	err = db.QueryRow(query, event.ID).Scan(
		&retrieved.ID,
		&retrieved.Timestamp,
		&retrieved.Command,
		&retrieved.Args,
		&retrieved.ExitCode,
		&retrieved.DurationMs,
		&retrieved.FilesTouched,
		&retrieved.RunID,
		&retrieved.Branch,
		&retrieved.Commit,
		&retrieved.Metadata,
	)
	if err != nil {
		t.Fatalf("Failed to query event: %v", err)
	}

	// Verify each field
	if retrieved.ID != event.ID {
		t.Errorf("ID: got %q, want %q", retrieved.ID, event.ID)
	}
	if !retrieved.Timestamp.Equal(event.Timestamp) {
		t.Errorf("Timestamp: got %v, want %v", retrieved.Timestamp, event.Timestamp)
	}
	if retrieved.Command != event.Command {
		t.Errorf("Command: got %q, want %q", retrieved.Command, event.Command)
	}
	if retrieved.Args != event.Args {
		t.Errorf("Args: got %q, want %q", retrieved.Args, event.Args)
	}
	if retrieved.ExitCode != event.ExitCode {
		t.Errorf("ExitCode: got %d, want %d", retrieved.ExitCode, event.ExitCode)
	}
	if retrieved.DurationMs != event.DurationMs {
		t.Errorf("DurationMs: got %d, want %d", retrieved.DurationMs, event.DurationMs)
	}
	if retrieved.FilesTouched != event.FilesTouched {
		t.Errorf("FilesTouched: got %q, want %q", retrieved.FilesTouched, event.FilesTouched)
	}
	if retrieved.RunID != event.RunID {
		t.Errorf("RunID: got %q, want %q", retrieved.RunID, event.RunID)
	}
	if retrieved.Branch != event.Branch {
		t.Errorf("Branch: got %q, want %q", retrieved.Branch, event.Branch)
	}
	if retrieved.Commit != event.Commit {
		t.Errorf("Commit: got %q, want %q", retrieved.Commit, event.Commit)
	}
	if retrieved.Metadata != event.Metadata {
		t.Errorf("Metadata: got %q, want %q", retrieved.Metadata, event.Metadata)
	}

	// Verify JSON fields are valid
	var parsedArgs map[string]string
	if err := json.Unmarshal([]byte(retrieved.Args), &parsedArgs); err != nil {
		t.Errorf("Args is not valid JSON: %v", err)
	}
	var parsedFiles []string
	if err := json.Unmarshal([]byte(retrieved.FilesTouched), &parsedFiles); err != nil {
		t.Errorf("FilesTouched is not valid JSON: %v", err)
	}
	var parsedMetadata map[string]interface{}
	if err := json.Unmarshal([]byte(retrieved.Metadata), &parsedMetadata); err != nil {
		t.Errorf("Metadata is not valid JSON: %v", err)
	}
}
