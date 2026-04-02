package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/enforce"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/donovan-yohan/carabiner/internal/carabiner/validate"
	"github.com/spf13/cobra"
)

var (
	validateResult string
	validateName   string
	validateRunID  string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run or respond to agent validations",
	Long:  "Non-blocking reflection questions for agents. Run without args to execute validations, or with --result to record a response.",
}

var validateExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute all configured validations and print questions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return fmt.Errorf("finding config dir: %w", err)
		}

		enforceCfg, err := enforce.LoadEnforceConfig(cfgDir)
		if err != nil {
			return fmt.Errorf("loading enforce config: %w", err)
		}

		_ = enforceCfg
		return fmt.Errorf("execute not yet wired to enforce config - use respond subcommand")
	},
}

var validateRespondCmd = &cobra.Command{
	Use:   "respond",
	Short: "Record a validation response",
	RunE: func(cmd *cobra.Command, args []string) error {
		if validateName == "" {
			return fmt.Errorf("--name is required")
		}
		if validateRunID == "" {
			return fmt.Errorf("--run-id is required")
		}
		if validateResult != "pass" && validateResult != "fail" && validateResult != "irrelevant" {
			return fmt.Errorf("--result must be pass, fail, or irrelevant")
		}

		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return fmt.Errorf("finding config dir: %w", err)
		}

		dbPath := filepath.Join(cfgDir, "carabiner.db")
		db, err := events.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("initializing database: %w", err)
		}
		defer db.Close()

		if err := validate.RecordResult(db, validateName, validateRunID, validateResult); err != nil {
			return fmt.Errorf("recording result: %w", err)
		}

		fmt.Printf("Recorded %s for %s (run %s)\n", validateResult, validateName, validateRunID)
		return nil
	},
}

var validateStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show validation statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return fmt.Errorf("finding config dir: %w", err)
		}

		dbPath := filepath.Join(cfgDir, "carabiner.db")
		db, err := events.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("initializing database: %w", err)
		}
		defer db.Close()

		stats, err := validate.ValidationStats(db)
		if err != nil {
			return fmt.Errorf("querying stats: %w", err)
		}

		if len(stats) == 0 {
			fmt.Println("No validation data yet.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VALIDATION\tPENDING\tRESPONDED\tORPHANED\tLAST_RUN")
		for _, s := range stats {
			lastRun := "—"
			if s.LastRun != nil {
				lastRun = time.Since(*s.LastRun).Round(time.Second).String() + " ago"
			}
			fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\n", s.Name, s.Pending, s.Responded, s.Orphaned, lastRun)
		}
		return w.Flush()
	},
}

func init() {
	validateCmd.AddCommand(validateExecuteCmd)
	validateCmd.AddCommand(validateRespondCmd)
	validateCmd.AddCommand(validateStatsCmd)

	validateRespondCmd.Flags().StringVar(&validateResult, "result", "", "Response: pass, fail, or irrelevant")
	validateRespondCmd.Flags().StringVar(&validateName, "name", "", "Validation name")
	validateRespondCmd.Flags().StringVar(&validateRunID, "run-id", "", "Run ID from validate execute")

	if err := validateRespondCmd.MarkFlagRequired("result"); err != nil {
		panic(err)
	}
	if err := validateRespondCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := validateRespondCmd.MarkFlagRequired("run-id"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(validateCmd)
}
