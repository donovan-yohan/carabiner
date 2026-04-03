package validate

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner/enforce"
	"github.com/google/uuid"
)

type ValidationResult struct {
	Name      string
	Script    string
	Question  string
	RunID     string
	CreatedAt time.Time
	Error     error
}

func ExecuteValidations(validations []enforce.ValidationConfig) ([]ValidationResult, string, error) {
	runID := uuid.New().String()
	results := make([]ValidationResult, 0, len(validations))

	for _, v := range validations {
		cmd := exec.Command("sh", "-c", v.Script)
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
