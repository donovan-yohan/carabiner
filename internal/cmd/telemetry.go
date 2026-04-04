package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/donovan-yohan/carabiner/internal/carabiner/telemetry"
	"github.com/spf13/cobra"
)

var (
	telemetryImportSource string
	telemetryImportLimit  int
)

var telemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Telemetry data management (experimental)",
	Long:  "Commands for importing and managing telemetry data from external sources. This is an experimental feature.",
}

var telemetryImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import telemetry from external sources",
	Long:  "Import telemetry data from external sources into the carabiner event log.",
}

var telemetryImportAgentlyticsCmd = &cobra.Command{
	Use:   "agentlytics",
	Short: "Import agentlytics sessions into local workflow history",
	Long: `Import agentlytics sessions from the agentlytics cache database into carabiner's workflow_events table.

This command reads session data from the agentlytics cache and normalizes it into carabiner's
workflow event format. Sessions are imported with idempotent watermark behavior - duplicate
imports are automatically skipped.

The import is conservative and only reads stable, known columns. Unknown columns are
stored in the metadata JSON field for future analysis.

This is an experimental/internal feature.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := telemetryImportSource
		if sourcePath == "" {
			sourcePath = telemetry.DefaultAgentlyticsCachePath()
			if sourcePath == "" {
				return fmt.Errorf("could not determine default agentlytics cache path")
			}
		}

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return fmt.Errorf("agentlytics cache not found at %s", sourcePath)
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

		opts := telemetry.AgentlyticsImportOptions{
			SourcePath: sourcePath,
			Limit:      telemetryImportLimit,
		}

		imported, err := telemetry.ImportAgentlytics(db, opts)
		if err != nil {
			return fmt.Errorf("importing agentlytics: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Imported %d sessions from agentlytics\n", imported)
		return nil
	},
}

func init() {
	telemetryImportAgentlyticsCmd.Flags().StringVar(&telemetryImportSource, "source", "", "path to agentlytics cache.db (default: ~/.agentlytics/cache.db)")
	telemetryImportAgentlyticsCmd.Flags().IntVar(&telemetryImportLimit, "limit", 0, "maximum number of sessions to import (0 = no limit)")

	telemetryImportCmd.AddCommand(telemetryImportAgentlyticsCmd)
	telemetryCmd.AddCommand(telemetryImportCmd)
	rootCmd.AddCommand(telemetryCmd)
}
