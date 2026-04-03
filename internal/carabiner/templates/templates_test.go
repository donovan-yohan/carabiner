package templates

import (
	"strings"
	"testing"
)

func TestGetTemplate_ReactTypescript(t *testing.T) {
	tmpl, err := GetTemplate("react-typescript")
	if err != nil {
		t.Fatalf("GetTemplate(\"react-typescript\") failed: %v", err)
	}

	if tmpl.Name != "react-typescript" {
		t.Errorf("Expected Name \"react-typescript\", got %q", tmpl.Name)
	}

	if tmpl.EnforceYAML == "" {
		t.Error("EnforceYAML should not be empty")
	}

	if len(tmpl.ConfigFiles) == 0 {
		t.Error("ConfigFiles should not be empty")
	}
}

func TestGetTemplate_Go(t *testing.T) {
	tmpl, err := GetTemplate("go")
	if err != nil {
		t.Fatalf("GetTemplate(\"go\") failed: %v", err)
	}

	if tmpl.Name != "go" {
		t.Errorf("Expected Name \"go\", got %q", tmpl.Name)
	}

	if tmpl.EnforceYAML == "" {
		t.Error("EnforceYAML should not be empty")
	}

	if len(tmpl.ConfigFiles) == 0 {
		t.Error("ConfigFiles should not be empty")
	}
}

func TestGetTemplate_Unknown(t *testing.T) {
	_, err := GetTemplate("unknown-template")
	if err == nil {
		t.Error("GetTemplate(\"unknown-template\") should return error")
	}

	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("Expected error to contain \"unknown template\", got %v", err)
	}
}

func TestListTemplates(t *testing.T) {
	templates := ListTemplates()

	hasReact := false
	hasGo := false
	for _, name := range templates {
		if name == "react-typescript" {
			hasReact = true
		}
		if name == "go" {
			hasGo = true
		}
	}

	if !hasReact {
		t.Error("ListTemplates should include \"react-typescript\"")
	}
	if !hasGo {
		t.Error("ListTemplates should include \"go\"")
	}
}

func TestReactTypescriptTemplate_HasRequiredFiles(t *testing.T) {
	tmpl, err := GetTemplate("react-typescript")
	if err != nil {
		t.Fatalf("GetTemplate(\"react-typescript\") failed: %v", err)
	}

	requiredFiles := []string{".eslintrc.json", "tsconfig.json", ".prettierrc"}
	for _, filename := range requiredFiles {
		if _, exists := tmpl.ConfigFiles[filename]; !exists {
			t.Errorf("React-TypeScript template missing required file: %s", filename)
		}
	}

	// Verify enforce.yaml has required tools
	if !strings.Contains(tmpl.EnforceYAML, "eslint:") {
		t.Error("enforce.yaml should contain eslint tool")
	}
	if !strings.Contains(tmpl.EnforceYAML, "typescript:") {
		t.Error("enforce.yaml should contain typescript tool")
	}
	if !strings.Contains(tmpl.EnforceYAML, "prettier:") {
		t.Error("enforce.yaml should contain prettier tool")
	}
	if !strings.Contains(tmpl.EnforceYAML, "vitest:") {
		t.Error("enforce.yaml should contain vitest tool")
	}

	// Verify behavior section
	if !strings.Contains(tmpl.EnforceYAML, "fail_on_warning: true") {
		t.Error("enforce.yaml should have fail_on_warning: true")
	}
}

func TestGoTemplate_HasRequiredFiles(t *testing.T) {
	tmpl, err := GetTemplate("go")
	if err != nil {
		t.Fatalf("GetTemplate(\"go\") failed: %v", err)
	}

	requiredFiles := []string{".golangci.yml", "Makefile"}
	for _, filename := range requiredFiles {
		if _, exists := tmpl.ConfigFiles[filename]; !exists {
			t.Errorf("Go template missing required file: %s", filename)
		}
	}

	// Verify enforce.yaml has required tools
	if !strings.Contains(tmpl.EnforceYAML, "golangci-lint:") {
		t.Error("enforce.yaml should contain golangci-lint tool")
	}
	if !strings.Contains(tmpl.EnforceYAML, "gofmt:") {
		t.Error("enforce.yaml should contain gofmt tool")
	}
	if !strings.Contains(tmpl.EnforceYAML, "staticcheck:") {
		t.Error("enforce.yaml should contain staticcheck tool")
	}
	if !strings.Contains(tmpl.EnforceYAML, "gotestsum:") {
		t.Error("enforce.yaml should contain gotestsum tool")
	}

	// Verify behavior section
	if !strings.Contains(tmpl.EnforceYAML, "fail_on_warning: true") {
		t.Error("enforce.yaml should have fail_on_warning: true")
	}
}

func TestReactTypescriptTemplate_ValidYAML(t *testing.T) {
	tmpl, err := GetTemplate("react-typescript")
	if err != nil {
		t.Fatalf("GetTemplate(\"react-typescript\") failed: %v", err)
	}

	// Basic YAML validation - check for proper structure
	if !strings.Contains(tmpl.EnforceYAML, "version: 1") {
		t.Error("enforce.yaml should start with version: 1")
	}
	if !strings.Contains(tmpl.EnforceYAML, "tools:") {
		t.Error("enforce.yaml should contain tools section")
	}
	if !strings.Contains(tmpl.EnforceYAML, "behavior:") {
		t.Error("enforce.yaml should contain behavior section")
	}
}

func TestGoTemplate_ValidYAML(t *testing.T) {
	tmpl, err := GetTemplate("go")
	if err != nil {
		t.Fatalf("GetTemplate(\"go\") failed: %v", err)
	}

	// Basic YAML validation - check for proper structure
	if !strings.Contains(tmpl.EnforceYAML, "version: 1") {
		t.Error("enforce.yaml should start with version: 1")
	}
	if !strings.Contains(tmpl.EnforceYAML, "tools:") {
		t.Error("enforce.yaml should contain tools section")
	}
	if !strings.Contains(tmpl.EnforceYAML, "behavior:") {
		t.Error("enforce.yaml should contain behavior section")
	}
}

func TestReactTypescriptTemplate_ConfigFilesValid(t *testing.T) {
	tmpl, err := GetTemplate("react-typescript")
	if err != nil {
		t.Fatalf("GetTemplate(\"react-typescript\") failed: %v", err)
	}

	// Verify .eslintrc.json is valid JSON
	eslintrc := tmpl.ConfigFiles[".eslintrc.json"]
	if !strings.Contains(eslintrc, `"extends"`) {
		t.Error(".eslintrc.json should contain extends field")
	}
	if !strings.Contains(eslintrc, `"rules"`) {
		t.Error(".eslintrc.json should contain rules field")
	}

	// Verify tsconfig.json is valid JSON
	tsconfig := tmpl.ConfigFiles["tsconfig.json"]
	if !strings.Contains(tsconfig, `"compilerOptions"`) {
		t.Error("tsconfig.json should contain compilerOptions field")
	}
	if !strings.Contains(tsconfig, `"strict": true`) {
		t.Error("tsconfig.json should have strict: true")
	}

	// Verify .prettierrc is valid JSON
	prettierrc := tmpl.ConfigFiles[".prettierrc"]
	if !strings.Contains(prettierrc, `"semi"`) {
		t.Error(".prettierrc should contain semi field")
	}
}

func TestGoTemplate_ConfigFilesValid(t *testing.T) {
	tmpl, err := GetTemplate("go")
	if err != nil {
		t.Fatalf("GetTemplate(\"go\") failed: %v", err)
	}

	// Verify .golangci.yml is valid YAML
	golangci := tmpl.ConfigFiles[".golangci.yml"]
	if !strings.Contains(golangci, "linters:") {
		t.Error(".golangci.yml should contain linters section")
	}
	if !strings.Contains(golangci, "enable:") {
		t.Error(".golangci.yml should contain enable field")
	}

	// Verify Makefile has required targets
	makefile := tmpl.ConfigFiles["Makefile"]
	if !strings.Contains(makefile, "lint:") {
		t.Error("Makefile should contain lint target")
	}
	if !strings.Contains(makefile, "test:") {
		t.Error("Makefile should contain test target")
	}
	if !strings.Contains(makefile, "coverage:") {
		t.Error("Makefile should contain coverage target")
	}
}
