package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/spf13/cobra"
)

var (
	recordGateID      string
	recordGateResult  string
	recordSkipExtract bool
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record a learning from a gate failure",
	Long:  "Create a new quality learning from a gate failure or review finding. Extracts structured patterns via model CLI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return err
		}

		cfg, err := carabiner.LoadConfig(cfgDir)
		if err != nil {
			return err
		}

		var rawInput string
		var gateResult *carabiner.GateResult

		if recordGateResult != "" {
			data, err := os.ReadFile(recordGateResult)
			if err != nil {
				return fmt.Errorf("reading gate result file: %w", err)
			}
			rawInput = string(data)

			var gr carabiner.GateResult
			if err := json.Unmarshal(data, &gr); err != nil {
				return fmt.Errorf("parsing gate result JSON: %w", err)
			}
			gateResult = &gr
		} else {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			rawInput = string(data)
		}

		if rawInput == "" {
			return fmt.Errorf("no input provided (use --gate-result or pipe to stdin)")
		}

		extractCmd := cfg.Quality.Extraction.Command
		if recordSkipExtract {
			extractCmd = ""
		}

		learning, err := carabiner.Record(cfgDir, recordGateID, rawInput, gateResult, recordSkipExtract, extractCmd)
		if err != nil {
			return err
		}

		fmt.Printf("Recorded learning %s: %s\n", learning.ID, learning.Pattern)
		return nil
	},
}

func init() {
	recordCmd.Flags().StringVar(&recordGateID, "gate-id", "", "identifier for the gate run (required)")
	recordCmd.Flags().StringVar(&recordGateResult, "gate-result", "", "path to gate result JSON file")
	recordCmd.Flags().BoolVar(&recordSkipExtract, "skip-extraction", false, "store raw input without model extraction")
	if err := recordCmd.MarkFlagRequired("gate-id"); err != nil {
		panic(err)
	}
	qualityCmd.AddCommand(recordCmd)
}
