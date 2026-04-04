package telemetry

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	_ "modernc.org/sqlite"
)

type AgentlyticsImportOptions struct {
	SourcePath string
	Limit      int
}

func DefaultAgentlyticsCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".agentlytics", "cache.db")
}

func ImportAgentlytics(db *sql.DB, opts AgentlyticsImportOptions) (int, error) {
	sourceDB, err := sql.Open("sqlite", opts.SourcePath)
	if err != nil {
		return 0, fmt.Errorf("opening agentlytics database: %w", err)
	}
	defer sourceDB.Close()

	tables, err := discoverTables(sourceDB)
	if err != nil {
		return 0, fmt.Errorf("discovering tables: %w", err)
	}

	if !hasTable(tables, "sessions") {
		return 0, fmt.Errorf("agentlytics database missing required 'sessions' table")
	}

	sessionColumns, err := discoverColumns(sourceDB, "sessions")
	if err != nil {
		return 0, fmt.Errorf("discovering sessions columns: %w", err)
	}

	sessions, err := readSessions(sourceDB, sessionColumns, opts.Limit)
	if err != nil {
		return 0, fmt.Errorf("reading sessions: %w", err)
	}

	imported := 0
	for _, session := range sessions {
		exists, err := workflowEventExists(db, session.ID)
		if err != nil {
			return imported, fmt.Errorf("checking existing event: %w", err)
		}
		if exists {
			continue
		}

		event, err := sessionToWorkflowEvent(session)
		if err != nil {
			return imported, fmt.Errorf("converting session to workflow event: %w", err)
		}

		if err := events.AppendWorkflowEvent(db, event); err != nil {
			return imported, fmt.Errorf("inserting workflow event: %w", err)
		}
		imported++
	}

	return imported, nil
}

func discoverTables(db *sql.DB) ([]string, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func hasTable(tables []string, name string) bool {
	return slices.Contains(tables, name)
}

func discoverColumns(db *sql.DB, table string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", table)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name string
		var dtype string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, nil
}

type Session struct {
	ID           string
	CreatedAt    time.Time
	Editor       string
	RepoPath     string
	Model        string
	MessageCount int
	Extra        map[string]interface{}
}

func readSessions(db *sql.DB, columns []string, limit int) ([]Session, error) {
	query := "SELECT "

	knownCols := map[string]bool{
		"id": true, "created_at": true, "editor": true,
		"repo_path": true, "model": true, "message_count": true,
	}

	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col
	}
	query += " FROM sessions"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		colMap := make(map[string]interface{})
		for i, col := range columns {
			colMap[col] = values[i]
		}

		session := Session{
			Extra: make(map[string]interface{}),
		}

		if id, ok := colMap["id"]; ok {
			if id != nil {
				session.ID = fmt.Sprintf("%v", id)
			}
		}

		if createdAt, ok := colMap["created_at"]; ok {
			if createdAt != nil {
				if str, ok := createdAt.(string); ok {
					if t, err := time.Parse(time.RFC3339, str); err == nil {
						session.CreatedAt = t
					} else if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
						session.CreatedAt = t
					}
				}
			}
		}

		if editor, ok := colMap["editor"]; ok {
			if editor != nil {
				session.Editor = fmt.Sprintf("%v", editor)
			}
		}

		if repoPath, ok := colMap["repo_path"]; ok {
			if repoPath != nil {
				session.RepoPath = fmt.Sprintf("%v", repoPath)
			}
		}

		if model, ok := colMap["model"]; ok {
			if model != nil {
				session.Model = fmt.Sprintf("%v", model)
			}
		}

		if messageCount, ok := colMap["message_count"]; ok {
			if messageCount != nil {
				switch v := messageCount.(type) {
				case int64:
					session.MessageCount = int(v)
				case int:
					session.MessageCount = v
				}
			}
		}

		for col, val := range colMap {
			if !knownCols[col] && val != nil {
				session.Extra[col] = val
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func workflowEventExists(db *sql.DB, sessionID string) (bool, error) {
	id := fmt.Sprintf("agentlytics:%s", sessionID)
	query := "SELECT COUNT(*) FROM workflow_events WHERE id = ?"
	var count int
	err := db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func sessionToWorkflowEvent(session Session) (*events.WorkflowEvent, error) {
	metadata := make(map[string]interface{})
	if len(session.Extra) > 0 {
		metadata["agentlytics_fields"] = session.Extra
	}
	if session.MessageCount > 0 {
		metadata["message_count"] = session.MessageCount
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata: %w", err)
	}

	return &events.WorkflowEvent{
		ID:                fmt.Sprintf("agentlytics:%s", session.ID),
		Timestamp:         session.CreatedAt,
		Workflow:          "agentlytics",
		EventType:         "session_imported",
		ExternalSessionID: session.ID,
		RepoPath:          session.RepoPath,
		Agent:             session.Editor,
		Model:             session.Model,
		Metadata:          string(metadataJSON),
	}, nil
}
