package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/spf13/cobra"
)

var (
	eventsListCommand string
	eventsListBranch  string
	eventsListRunID   string
	eventsListLimit   int
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Event log management",
	Long:  "Commands for initializing and querying the carabiner events database.",
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded events",
	Long:  "List recent events from the events database with optional filters.",
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

		filter := &events.EventFilter{
			Command: eventsListCommand,
			Branch:  eventsListBranch,
			RunID:   eventsListRunID,
			Limit:   eventsListLimit,
		}

		eventsList, err := events.ListEvents(db, filter)
		if err != nil {
			return err
		}

		if len(eventsList) == 0 {
			fmt.Fprintln(os.Stdout, "No events found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tTIMESTAMP\tEXIT_CODE\tDURATION")
		for _, event := range eventsList {
			duration := time.Duration(event.DurationMs) * time.Millisecond
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", event.Command, event.Timestamp.Format(time.RFC3339), event.ExitCode, duration)
		}
		return w.Flush()
	},
}

var eventsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize events database",
	Long:  "Initialize the events database in the active .carabiner config directory.",
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

		fmt.Fprintf(os.Stdout, "Initialized events database at %s\n", dbPath)
		return nil
	},
}

func init() {
	eventsListCmd.Flags().StringVar(&eventsListCommand, "command", "", "filter by command")
	eventsListCmd.Flags().StringVar(&eventsListBranch, "branch", "", "filter by branch")
	eventsListCmd.Flags().StringVar(&eventsListRunID, "run-id", "", "filter by run ID")
	eventsListCmd.Flags().IntVar(&eventsListLimit, "limit", 10, "maximum number of results")

	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsInitCmd)
	rootCmd.AddCommand(eventsCmd)
}
