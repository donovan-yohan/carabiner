package cmd

import (
	"fmt"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
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
		fmt.Printf("Initialized carabiner at %s\n", dir)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initMode, "mode", "repo", "storage mode: 'repo' (committed) or 'local' (~/.carabiner/)")
	rootCmd.AddCommand(initCmd)
}
