package carabiner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ExtractionResult holds the structured output from model extraction.
type ExtractionResult struct {
	Pattern        string   `json:"pattern"`
	Recommendation string   `json:"recommendation"`
	Paths          []string `json:"paths"`
	Tags           []string `json:"tags"`
}

const extractionPrompt = `You are extracting a structured quality pattern from a gate failure report.

INPUT (gate failure rationale):
%s

INPUT (files changed):
%s

Extract a structured quality pattern. Output ONLY valid JSON with these fields:
{
  "pattern": "one-line description of what went wrong",
  "recommendation": "what to do instead",
  "paths": ["directory/prefixes/that/this/applies/to/"],
  "tags": ["relevant", "categories"]
}

Rules:
- paths must be directory prefixes ending in "/"
- pattern should be actionable, not vague
- recommendation should be specific enough to follow
- tags should be 1-3 relevant categories`

// Extract shells out to the configured model CLI and parses the response.
func Extract(command string, rationale string, files []string, timeout time.Duration) (*ExtractionResult, error) {
	filesStr := strings.Join(files, "\n")
	prompt := fmt.Sprintf(extractionPrompt, rationale, filesStr)

	// Write prompt to temp file to avoid pipe deadlock
	tmpFile, err := os.CreateTemp("", "carabiner-extract-*.txt")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(prompt); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("writing prompt: %w", err)
	}
	tmpFile.Close()

	// Try extraction with one retry
	for attempt := 0; attempt < 2; attempt++ {
		result, err := runExtraction(command, tmpFile.Name(), timeout)
		if err == nil {
			return result, nil
		}
		if attempt == 0 {
			fmt.Fprintf(os.Stderr, "warning: extraction attempt 1 failed: %v, retrying...\n", err)
			time.Sleep(2 * time.Second)
		} else {
			return nil, fmt.Errorf("extraction failed after 2 attempts: %w", err)
		}
	}
	return nil, fmt.Errorf("extraction failed")
}

func runExtraction(command string, promptFile string, timeout time.Duration) (*ExtractionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty extraction command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	promptData, err := os.Open(promptFile)
	if err != nil {
		return nil, fmt.Errorf("opening prompt file: %w", err)
	}
	defer promptData.Close()
	cmd.Stdin = promptData

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running extraction command: %w", err)
	}

	var result ExtractionResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing extraction output: %w", err)
	}

	if result.Pattern == "" {
		return nil, fmt.Errorf("extraction returned empty pattern")
	}
	return &result, nil
}
