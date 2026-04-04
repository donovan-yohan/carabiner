package cmd

import "testing"

func TestParseFileLine_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantFile string
		wantLine int
	}{
		{"src/main.go:47", "src/main.go", 47},
		{"handler.go:1", "handler.go", 1},
		{"path/to/file.go:100", "path/to/file.go", 100},
	}
	for _, tt := range tests {
		file, line, err := parseFileLine(tt.input)
		if err != nil {
			t.Errorf("parseFileLine(%q) error: %v", tt.input, err)
			continue
		}
		if file != tt.wantFile {
			t.Errorf("parseFileLine(%q) file = %q, want %q", tt.input, file, tt.wantFile)
		}
		if line != tt.wantLine {
			t.Errorf("parseFileLine(%q) line = %d, want %d", tt.input, line, tt.wantLine)
		}
	}
}

func TestParseFileLine_NoColon(t *testing.T) {
	_, _, err := parseFileLine("src/main.go")
	if err == nil {
		t.Error("expected error for missing colon")
	}
}

func TestParseFileLine_NoLine(t *testing.T) {
	_, _, err := parseFileLine("src/main.go:")
	if err == nil {
		t.Error("expected error for missing line number")
	}
}

func TestParseFileLine_BadLine(t *testing.T) {
	_, _, err := parseFileLine("src/main.go:abc")
	if err == nil {
		t.Error("expected error for non-numeric line")
	}
}

func TestParseFileLine_ZeroLine(t *testing.T) {
	_, _, err := parseFileLine("src/main.go:0")
	if err == nil {
		t.Error("expected error for line 0")
	}
}

func TestParseFileLine_NegativeLine(t *testing.T) {
	_, _, err := parseFileLine("src/main.go:-1")
	if err == nil {
		t.Error("expected error for negative line")
	}
}
