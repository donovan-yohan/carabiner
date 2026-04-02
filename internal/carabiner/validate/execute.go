package validate

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

type ValidationConfig struct {
	Name   string
	Script string
	Files  []string
}

type ValidationResult struct {
	Name      string
	Script    string
	Question  string
	RunID     string
	CreatedAt time.Time
	Error     error
}

// ExecuteValidations runs all configured validations and returns their results.
// It generates a new run ID and marks any pending records from previous runs as orphaned.
func ExecuteValidations(validations []ValidationConfig) ([]ValidationResult, string, error) {
	runID := uuid.New().String()
	results := make([]ValidationResult, 0, len(validations))

	for _, v := range validations {
		cmd := exec.Command("bash", "-c", v.Script)
		output, err := cmd.CombinedOutput()

		result := ValidationResult{
			Name:      v.Name,
			Script:    v.Script,
			Question:  string(output),
			RunID:     runID,
			CreatedAt: time.Now(),
		}
		if err != nil {
			result.Error = fmt.Errorf("script error: %w", err)
		}
		results = append(results, result)
	}

	return results, runID, nil
}
