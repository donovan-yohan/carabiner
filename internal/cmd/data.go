package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/spf13/cobra"
)

var (
	dataFormat string
	querySQL   string
)

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Data export and query",
	Long:  "Commands for exporting and querying carabiner data.",
}

var dataExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all data",
	Long:  "Export all data from the events database as JSON.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return err
		}

		dbPath := filepath.Join(cfgDir, "carabiner.db")
		db, err := events.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("initializing events database: %w", err)
		}
		defer db.Close()

		export := struct {
			WorkflowEvents    []events.WorkflowEvent    `json:"workflow_events"`
			WorkContextEvents []events.WorkContextEvent `json:"work_context_events"`
			GitAttributions   []events.GitAttribution   `json:"git_attributions"`
			ExportedAt        string                    `json:"exported_at"`
		}{
			ExportedAt: time.Now().UTC().Format(time.RFC3339),
		}

		export.WorkflowEvents, err = events.ListWorkflowEvents(db, "", 0)
		if err != nil {
			return fmt.Errorf("querying workflow events: %w", err)
		}

		export.WorkContextEvents, err = events.ListWorkContextEvents(db, 0)
		if err != nil {
			return fmt.Errorf("querying work context events: %w", err)
		}

		export.GitAttributions, err = events.ListRecentAttributions(db, 0)
		if err != nil {
			return fmt.Errorf("querying git attributions: %w", err)
		}

		output, err := json.MarshalIndent(export, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling export: %w", err)
		}

		fmt.Fprintln(os.Stdout, string(output))
		return nil
	},
}

var dataQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Execute read-only SQL query (experimental/internal)",
	Long: `Execute a read-only SQL query against the events database.

This command is experimental and intended for internal use only.
Only SELECT queries are allowed for safety.

Example:
  carabiner data query --sql "SELECT workflow, COUNT(*) FROM workflow_events GROUP BY workflow"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if querySQL == "" {
			return cmd.Help()
		}

		if !isReadOnlySQL(querySQL) {
			return fmt.Errorf("carabiner data query only supports SELECT queries")
		}

		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return err
		}

		dbPath := filepath.Join(cfgDir, "carabiner.db")
		db, err := events.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("initializing events database: %w", err)
		}
		defer db.Close()

		rows, err := db.Query(querySQL)
		if err != nil {
			return fmt.Errorf("executing query: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("getting columns: %w", err)
		}

		switch dataFormat {
		case "json":
			return outputJSON(rows, columns)
		case "table":
			return outputTable(rows, columns)
		default:
			return fmt.Errorf("unsupported format: %s (use 'json' or 'table')", dataFormat)
		}
	},
}

func isReadOnlySQL(sqlQuery string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(sqlQuery))

	if !strings.HasPrefix(trimmed, "select") {
		return false
	}

	dangerous := []string{"insert", "update", "delete", "drop", "create", "alter", "truncate", "replace"}
	for _, keyword := range dangerous {
		if strings.Contains(trimmed, keyword) {
			return false
		}
	}

	return true
}

func outputJSON(rows *sql.Rows, columns []string) error {
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating rows: %w", err)
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling results: %w", err)
	}

	fmt.Fprintln(os.Stdout, string(output))
	return nil
}

func outputTable(rows *sql.Rows, columns []string) error {
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(os.Stdout, "\t")
		}
		fmt.Fprint(os.Stdout, col)
	}
	fmt.Fprintln(os.Stdout)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}

		for i, val := range values {
			if i > 0 {
				fmt.Fprint(os.Stdout, "\t")
			}
			switch v := val.(type) {
			case []byte:
				fmt.Fprint(os.Stdout, string(v))
			case string:
				fmt.Fprint(os.Stdout, v)
			case int64:
				fmt.Fprint(os.Stdout, v)
			case float64:
				fmt.Fprint(os.Stdout, v)
			case nil:
				fmt.Fprint(os.Stdout, "NULL")
			default:
				fmt.Fprint(os.Stdout, v)
			}
		}
		fmt.Fprintln(os.Stdout)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating rows: %w", err)
	}

	return nil
}

func init() {
	dataExportCmd.Flags().StringVar(&dataFormat, "format", "json", "output format: json or table")
	dataQueryCmd.Flags().StringVar(&querySQL, "sql", "", "SQL query to execute (required)")
	dataQueryCmd.Flags().StringVar(&dataFormat, "format", "table", "output format: json or table")

	dataCmd.AddCommand(dataExportCmd)
	dataCmd.AddCommand(dataQueryCmd)
	rootCmd.AddCommand(dataCmd)
}
