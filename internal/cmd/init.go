package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/spf13/cobra"
)

var initMode string
var initTemplate string
var initAddOns []string
var initContext bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .carabiner directory",
	Long:  "Scaffold the .carabiner/ directory with config.yaml and quality subdirectories.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if initTemplate != "" {
			if err := carabiner.InitWithTemplate(initMode, initTemplate, initAddOns); err != nil {
				return err
			}

			fmt.Printf("Initialized carabiner with template '%s'\n", initTemplate)

			if initContext {
				if err := carabiner.ScaffoldContextSupport(); err != nil {
					return err
				}
				fmt.Println("Scaffolded context support (hooks and agent instructions)")
			}

			return nil
		}

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

		if initContext {
			if err := carabiner.ScaffoldContextSupport(); err != nil {
				return err
			}
			fmt.Println("Scaffolded context support (hooks and agent instructions)")
		}

		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initMode, "mode", "repo", "storage mode: 'repo' (committed) or 'local' (~/.carabiner/)")
	initCmd.Flags().StringVar(&initTemplate, "template", "", "Template to scaffold (go, react-typescript)")
	initCmd.Flags().StringSliceVar(&initAddOns, "add-ons", nil, "Add-ons to install (e.g., vigiles)")
	initCmd.Flags().BoolVar(&initContext, "context", false, "Initialize with work context support (hooks and agent instructions)")
	rootCmd.AddCommand(initCmd)
}
