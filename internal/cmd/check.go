package cmd

import (
	"fmt"
	"os"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/spf13/cobra"
)

var checkFiles []string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for relevant quality patterns",
	Long:  "Retrieve quality patterns relevant to the given file paths. Output is prompt-ready markdown.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(checkFiles) == 0 {
			return fmt.Errorf("--files is required")
		}

		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return err
		}

		cfg, err := carabiner.LoadConfig(cfgDir)
		if err != nil {
			return err
		}

		learnings, err := carabiner.Check(cfgDir, checkFiles, cfg.Quality.Check.MaxResults)
		if err != nil {
			return err
		}

		if len(learnings) == 0 {
			os.Exit(1)
		}

		fmt.Print(carabiner.FormatMarkdown(learnings))
		return nil
	},
}

func init() {
	checkCmd.Flags().StringSliceVar(&checkFiles, "files", nil, "file paths to check for relevant patterns")
	checkCmd.MarkFlagRequired("files")
	qualityCmd.AddCommand(checkCmd)
}
