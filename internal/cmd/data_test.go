package cmd

import (
	"testing"
)

func TestIsReadOnlySQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		{
			name:     "simple select",
			sql:      "SELECT * FROM events",
			expected: true,
		},
		{
			name:     "select with where",
			sql:      "SELECT workflow, COUNT(*) FROM workflow_events WHERE workflow = 'test' GROUP BY workflow",
			expected: true,
		},
		{
			name:     "select with uppercase",
			sql:      "SELECT * FROM events",
			expected: true,
		},
		{
			name:     "select with whitespace",
			sql:      "  SELECT * FROM events  ",
			expected: true,
		},
		{
			name:     "insert rejected",
			sql:      "INSERT INTO events VALUES (1, 2, 3)",
			expected: false,
		},
		{
			name:     "update rejected",
			sql:      "UPDATE events SET command = 'test'",
			expected: false,
		},
		{
			name:     "delete rejected",
			sql:      "DELETE FROM events",
			expected: false,
		},
		{
			name:     "drop rejected",
			sql:      "DROP TABLE events",
			expected: false,
		},
		{
			name:     "create rejected",
			sql:      "CREATE TABLE test (id INT)",
			expected: false,
		},
		{
			name:     "alter rejected",
			sql:      "ALTER TABLE events ADD COLUMN test TEXT",
			expected: false,
		},
		{
			name:     "truncate rejected",
			sql:      "TRUNCATE TABLE events",
			expected: false,
		},
		{
			name:     "replace rejected",
			sql:      "REPLACE INTO events VALUES (1, 2, 3)",
			expected: false,
		},
		{
			name:     "select with insert in string rejected",
			sql:      "SELECT * FROM events WHERE command = 'insert'",
			expected: false,
		},
		{
			name:     "empty string rejected",
			sql:      "",
			expected: false,
		},
		{
			name:     "non-select rejected",
			sql:      "SHOW TABLES",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReadOnlySQL(tt.sql)
			if result != tt.expected {
				t.Errorf("isReadOnlySQL(%q) = %v, expected %v", tt.sql, result, tt.expected)
			}
		})
	}
}

func TestReadOnlySQLCaseInsensitive(t *testing.T) {
	variations := []string{
		"SELECT * FROM events",
		"select * from events",
		"Select * From events",
		"  SELECT * FROM events  ",
	}

	for _, sql := range variations {
		if !isReadOnlySQL(sql) {
			t.Errorf("isReadOnlySQL(%q) should return true", sql)
		}
	}
}

func TestDangerousKeywordsDetected(t *testing.T) {
	dangerousSQL := []string{
		"SELECT * FROM events; INSERT INTO events VALUES (1)",
		"SELECT * FROM events WHERE command = 'update'",
		"SELECT * FROM events WHERE command LIKE '%delete%'",
	}

	for _, sql := range dangerousSQL {
		if isReadOnlySQL(sql) {
			t.Errorf("isReadOnlySQL(%q) should return false (contains dangerous keyword)", sql)
		}
	}
}
