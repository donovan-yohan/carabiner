package agentlytics

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cache.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE chats (
			id TEXT PRIMARY KEY,
			source TEXT,
			folder TEXT,
			created_at TEXT,
			last_updated_at TEXT,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO chats (id, source, folder, created_at, last_updated_at, name)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"655eb4a6-822f-4d52-8c06-cade4afdcd8d",
		"claude-code",
		"/Users/test/project",
		"2026-03-28T14:32:00Z",
		"2026-03-28T15:47:00Z",
		"Implement auth handler",
	)
	if err != nil {
		t.Fatal(err)
	}

	return dbPath
}

func TestQuerySession_Found(t *testing.T) {
	dbPath := setupTestDB(t)

	session, err := QuerySession(dbPath, "655eb4a6-822f-4d52-8c06-cade4afdcd8d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.ID != "655eb4a6-822f-4d52-8c06-cade4afdcd8d" {
		t.Errorf("id = %q", session.ID)
	}
	if session.Source != "claude-code" {
		t.Errorf("source = %q", session.Source)
	}
	if session.Name != "Implement auth handler" {
		t.Errorf("name = %q", session.Name)
	}
	if session.Folder != "/Users/test/project" {
		t.Errorf("folder = %q", session.Folder)
	}
}

func TestQuerySession_NotFound(t *testing.T) {
	dbPath := setupTestDB(t)

	session, err := QuerySession(dbPath, "nonexistent-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session != nil {
		t.Errorf("expected nil, got %+v", session)
	}
}

func TestQuerySession_DBMissing(t *testing.T) {
	_, err := QuerySession("/nonexistent/path/cache.db", "some-id")
	if err == nil {
		t.Error("expected error for missing DB")
	}
}

func TestCheckHealth_Found(t *testing.T) {
	dbPath := setupTestDB(t)

	h := CheckHealth(dbPath)
	if !h.Found {
		t.Error("expected Found=true")
	}
	if !h.SchemaValid {
		t.Errorf("expected SchemaValid=true, error: %s", h.Error)
	}
	if h.ChatCount != 1 {
		t.Errorf("chat count = %d, want 1", h.ChatCount)
	}
	if h.Error != "" {
		t.Errorf("unexpected error: %s", h.Error)
	}
}

func TestCheckHealth_Missing(t *testing.T) {
	h := CheckHealth("/nonexistent/path/cache.db")
	if h.Found {
		t.Error("expected Found=false")
	}
	if h.Error == "" {
		t.Error("expected error message")
	}
}

func TestCheckHealth_BadSchema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cache.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("CREATE TABLE wrong_table (id TEXT)")
	db.Close()

	h := CheckHealth(dbPath)
	if !h.Found {
		t.Error("expected Found=true")
	}
	if h.SchemaValid {
		t.Error("expected SchemaValid=false for wrong schema")
	}
}

func TestDefaultCachePath(t *testing.T) {
	path := DefaultCachePath()
	if path == "" {
		t.Skip("no home directory")
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".agentlytics", "cache.db")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}
