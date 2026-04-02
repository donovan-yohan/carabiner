package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/spf13/cobra"
)

var initMode string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .carabiner directory",
	Long:  "Scaffold the .carabiner/ directory with config.yaml and quality subdirectories.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := carabiner.Init(initMode)
		if err != nil {
			return err
		}

		db, err := events.InitDB(filepath.Join(dir, "carabiner.db"))
		if err != nil {
			return fmt.Errorf("initializing events database: %w", err)
		}
		defer db.Close()

		fmt.Printf("Initialized carabiner at %s\n", dir)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initMode, "mode", "repo", "storage mode: 'repo' (committed) or 'local' (~/.carabiner/)")
	rootCmd.AddCommand(initCmd)
}
