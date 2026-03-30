package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configDir string

var rootCmd = &cobra.Command{
	Use:   "carabiner",
	Short: "Agent-agnostic harness for coding agents",
	Long:  "Carabiner is a repo's institutional memory for coding agents. It persists quality patterns across sessions and surfaces them when agents touch affected files.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "path to .carabiner directory (default: walk up from CWD)")
}
