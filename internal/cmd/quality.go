package cmd

import "github.com/spf13/cobra"

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Quality pattern management",
	Long:  "Commands for managing quality patterns: checking relevant patterns, recording new ones from gate failures.",
}

func init() {
	rootCmd.AddCommand(qualityCmd)
}
