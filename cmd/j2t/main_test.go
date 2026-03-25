package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLI(t *testing.T) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Project root is two levels up from cmd/j2t
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	exe := filepath.Join(projectRoot, "bin", "j2t.exe")
	if _, err := os.Stat(exe); os.IsNotExist(err) {
		exe = filepath.Join(projectRoot, "bin", "j2t")
	}
	if _, err := os.Stat(exe); os.IsNotExist(err) {
		t.Skip("j2t binary not found, run 'make build' first")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON",
			input:    "testdata/example.json",
			expected: "testdata/example.json.out",
		},
		{
			name:     "JSONC",
			input:    "testdata/example.jsonc",
			expected: "testdata/example.jsonc.out",
		},
		{
			name:     "JSONL",
			input:    "testdata/example.jsonl",
			expected: "testdata/example.jsonl.out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run j2t
			cmd := exec.Command(exe, tt.input)
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("j2t %s failed: %v\nOutput: %s", tt.input, err, output)
			}

			// Read expected output
			expected, err := os.ReadFile(filepath.Join(projectRoot, tt.expected))
			if err != nil {
				t.Fatalf("Failed to read expected output %s: %v", tt.expected, err)
			}

			// Compare
			if !bytes.Equal(output, expected) {
				t.Errorf("Output mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s", tt.input, expected, output)
			}
		})
	}
}
