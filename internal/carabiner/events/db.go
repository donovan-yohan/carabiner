package events

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		command TEXT NOT NULL,
		args TEXT,
		exit_code INTEGER NOT NULL,
		duration_ms INTEGER,
		files_touched TEXT,
		run_id TEXT,
		branch TEXT,
		"commit" TEXT,
		metadata TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_events_run_id ON events(run_id);
	CREATE INDEX IF NOT EXISTS idx_events_command ON events(command);

	CREATE TABLE IF NOT EXISTS validation_events (
		id TEXT PRIMARY KEY,
		run_id TEXT NOT NULL,
		name TEXT NOT NULL,
		script TEXT NOT NULL,
		status TEXT NOT NULL,
		result TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		responded_at DATETIME,
		orphaned_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_validation_run ON validation_events(run_id);
	CREATE INDEX IF NOT EXISTS idx_validation_name ON validation_events(name);
	CREATE INDEX IF NOT EXISTS idx_validation_status ON validation_events(status);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
