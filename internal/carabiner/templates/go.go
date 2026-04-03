package templates

func getGoTemplate() *Template {
	enforceYAML := `version: 1
tools:
  golangci-lint:
    enabled: true
    command: "golangci-lint"
    args: ["run", "--tests"]
    files: ["**/*.go"]

  gofmt:
    enabled: true
    command: "gofmt"
    args: ["-l", "-d", "."]
    files: ["**/*.go"]

  staticcheck:
    enabled: true
    command: "staticcheck"
    args: ["."]
    files: ["**/*.go"]

  gotestsum:
    enabled: true
    command: "gotestsum"
    args: ["--", "-race", "./..."]
    files: ["**/*_test.go"]

behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`

	golangci := `linters:
  enable:
    - errcheck
    - gocognit
    - gocyclo
    - gosimple
    - govet
    - ineffassign
    - nestif
    - staticcheck
    - unused
linters-settings:
  gocognit:
    min-complexity: 15
  gocyclo:
    min-complexity: 10
  nestif:
    min-complexity: 4
run:
  timeout: 5m
`

	makefile := `.PHONY: lint test coverage

lint:
	golangci-lint run --tests --max-warnings=0

test:
	gotestsum -- -race -coverprofile=coverage.out

coverage:
	go tool cover -func=coverage.out

fmt:
	gofmt -l -d .
`

	return &Template{
		Name:        "go",
		EnforceYAML: enforceYAML,
		ConfigFiles: map[string]string{
			".golangci.yml": golangci,
			"Makefile":      makefile,
		},
	}
}
