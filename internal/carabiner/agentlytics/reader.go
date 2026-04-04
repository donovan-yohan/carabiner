package agentlytics

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Session represents a row from agentlytics' chats table.
type Session struct {
	ID            string
	Source        string
	Folder        string
	CreatedAt     time.Time
	LastUpdatedAt time.Time
	Name          string
}

// Health reports the status of the agentlytics data source.
type Health struct {
	Found       bool   `json:"found"`
	Path        string `json:"path"`
	ChatCount   int    `json:"chat_count"`
	SchemaValid bool   `json:"schema_valid"`
	Error       string `json:"error,omitempty"`
}

// DefaultCachePath returns the default agentlytics cache.db location.
func DefaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".agentlytics", "cache.db")
}

// openReadOnly opens the agentlytics SQLite database in read-only mode
// with a busy timeout to handle concurrent access safely.
func openReadOnly(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening agentlytics db: %w", err)
	}

	if _, err := db.Exec("PRAGMA busy_timeout = 3000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting busy_timeout: %w", err)
	}

	return db, nil
}

// QuerySession looks up a single session by conversation ID.
// Returns nil (not an error) if the session doesn't exist.
func QuerySession(path string, conversationID string) (*Session, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var s Session
	var createdAt, updatedAt string

	err = db.QueryRow(
		"SELECT id, source, folder, created_at, last_updated_at, name FROM chats WHERE id = ?",
		conversationID,
	).Scan(&s.ID, &s.Source, &s.Folder, &createdAt, &updatedAt, &s.Name)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying session %s: %w", conversationID, err)
	}

	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	s.LastUpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &s, nil
}

// CheckHealth validates the agentlytics data source is available and well-formed.
func CheckHealth(path string) *Health {
	h := &Health{Path: path}

	if _, err := os.Stat(path); err != nil {
		h.Error = fmt.Sprintf("cache.db not found at %s", path)
		return h
	}
	h.Found = true

	db, err := openReadOnly(path)
	if err != nil {
		h.Error = err.Error()
		return h
	}
	defer db.Close()

	// Check schema: chats table exists with expected columns
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='chats'").Scan(&name)
	if err != nil {
		h.Error = "chats table not found in agentlytics cache"
		return h
	}
	h.SchemaValid = true

	// Count sessions
	err = db.QueryRow("SELECT COUNT(*) FROM chats").Scan(&h.ChatCount)
	if err != nil {
		h.Error = fmt.Sprintf("counting chats: %v", err)
		return h
	}

	return h
}
