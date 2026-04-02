package events

import (
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestListEvents_NoFilter(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []*Event{
		{ID: "1", Timestamp: now.Add(-2 * time.Hour), Command: "init", ExitCode: 0, DurationMs: 100},
		{ID: "2", Timestamp: now.Add(-1 * time.Hour), Command: "quality", ExitCode: 0, DurationMs: 200},
		{ID: "3", Timestamp: now, Command: "query", ExitCode: 0, DurationMs: 300},
	}

	for _, e := range events {
		if err := AppendEvent(db, e); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	results, err := ListEvents(db, &EventFilter{})
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 events, got %d", len(results))
	}
}

func TestListEvents_FilterByCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []*Event{
		{ID: "1", Timestamp: now.Add(-2 * time.Hour), Command: "init", ExitCode: 0, DurationMs: 100},
		{ID: "2", Timestamp: now.Add(-1 * time.Hour), Command: "quality", ExitCode: 0, DurationMs: 200},
		{ID: "3", Timestamp: now, Command: "init", ExitCode: 1, DurationMs: 300},
	}

	for _, e := range events {
		if err := AppendEvent(db, e); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	results, err := ListEvents(db, &EventFilter{Command: "init"})
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 events with command 'init', got %d", len(results))
	}

	for _, r := range results {
		if r.Command != "init" {
			t.Errorf("expected command 'init', got %q", r.Command)
		}
	}
}

func TestListEvents_FilterByBranch(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []*Event{
		{ID: "1", Timestamp: now.Add(-2 * time.Hour), Command: "init", Branch: "main", ExitCode: 0, DurationMs: 100},
		{ID: "2", Timestamp: now.Add(-1 * time.Hour), Command: "quality", Branch: "feature", ExitCode: 0, DurationMs: 200},
		{ID: "3", Timestamp: now, Command: "query", Branch: "main", ExitCode: 0, DurationMs: 300},
	}

	for _, e := range events {
		if err := AppendEvent(db, e); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	results, err := ListEvents(db, &EventFilter{Branch: "main"})
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 events with branch 'main', got %d", len(results))
	}

	for _, r := range results {
		if r.Branch != "main" {
			t.Errorf("expected branch 'main', got %q", r.Branch)
		}
	}
}

func TestListEvents_FilterByRunID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []*Event{
		{ID: "1", Timestamp: now.Add(-2 * time.Hour), Command: "init", RunID: "run-1", ExitCode: 0, DurationMs: 100},
		{ID: "2", Timestamp: now.Add(-1 * time.Hour), Command: "quality", RunID: "run-2", ExitCode: 0, DurationMs: 200},
		{ID: "3", Timestamp: now, Command: "query", RunID: "run-1", ExitCode: 0, DurationMs: 300},
	}

	for _, e := range events {
		if err := AppendEvent(db, e); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	results, err := ListEvents(db, &EventFilter{RunID: "run-1"})
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 events with run_id 'run-1', got %d", len(results))
	}

	for _, r := range results {
		if r.RunID != "run-1" {
			t.Errorf("expected run_id 'run-1', got %q", r.RunID)
		}
	}
}

func TestListEvents_Limit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	for i := 0; i < 10; i++ {
		event := &Event{
			ID:         string(rune('A' + i)),
			Timestamp:  now.Add(time.Duration(i) * time.Hour),
			Command:    "test",
			ExitCode:   0,
			DurationMs: int64(i * 100),
		}
		if err := AppendEvent(db, event); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	results, err := ListEvents(db, &EventFilter{Limit: 5})
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 events with limit, got %d", len(results))
	}
}

func TestListEvents_Ordering(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []*Event{
		{ID: "oldest", Timestamp: now.Add(-2 * time.Hour), Command: "init", ExitCode: 0, DurationMs: 100},
		{ID: "middle", Timestamp: now.Add(-1 * time.Hour), Command: "quality", ExitCode: 0, DurationMs: 200},
		{ID: "newest", Timestamp: now, Command: "query", ExitCode: 0, DurationMs: 300},
	}

	for _, e := range events {
		if err := AppendEvent(db, e); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	results, err := ListEvents(db, &EventFilter{})
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 events, got %d", len(results))
	}

	if results[0].ID != "newest" {
		t.Errorf("expected first event to be 'newest', got %q", results[0].ID)
	}
	if results[1].ID != "middle" {
		t.Errorf("expected second event to be 'middle', got %q", results[1].ID)
	}
	if results[2].ID != "oldest" {
		t.Errorf("expected third event to be 'oldest', got %q", results[2].ID)
	}
}
