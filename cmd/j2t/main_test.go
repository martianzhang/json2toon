package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/martianzhang/json2toon"
)

// --- CLI integration tests ---

func TestCLI(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

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
			cmd := exec.Command(exe, tt.input)
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("j2t %s failed: %v\nOutput: %s", tt.input, err, output)
			}

			expectedPath := filepath.Join(projectRoot, tt.expected)
			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("Failed to read expected output %s: %v", tt.expected, err)
			}

			if !bytes.Equal(output, expected) {
				t.Errorf("Output mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s", tt.input, expected, output)
			}
		})
	}
}

// --- Library integration tests (using j2t internals) ---

func TestDetectFormatJSON(t *testing.T) {
	result := json2toon.DetectFormat([]byte(`{"id": 1}`))
	if result != "json" {
		t.Errorf("expected json, got %s", result)
	}
}

func TestDetectFormatJSONC(t *testing.T) {
	result := json2toon.DetectFormat([]byte(`{"id": 1} // comment`))
	if result != "jsonc" {
		t.Errorf("expected jsonc, got %s", result)
	}
}

func TestDetectFormatJSONL(t *testing.T) {
	result := json2toon.DetectFormat([]byte(`{"id": 1}
{"id": 2}`))
	if result != "jsonl" {
		t.Errorf("expected jsonl, got %s", result)
	}
}

func TestConvertAutoJSON(t *testing.T) {
	result, err := json2toon.ConvertAuto([]byte(`{"id": 1}`))
	if err != nil {
		t.Fatalf("ConvertAuto failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id in output: %s", result)
	}
}

func TestConvertAutoJSONC(t *testing.T) {
	result, err := json2toon.ConvertAuto([]byte(`{"id": 1} // comment`))
	if err != nil {
		t.Fatalf("ConvertAuto failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id in output: %s", result)
	}
}

func TestConvertAutoJSONL(t *testing.T) {
	jsonl := `{"id": 1}
{"id": 2}`
	result, err := json2toon.ConvertAuto([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertAuto failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id in output: %s", result)
	}
}

func TestConvertAutoWithIndent(t *testing.T) {
	result, err := json2toon.ConvertAutoWithOptions([]byte(`{"id": 1}`), json2toon.WithIndent(4))
	if err != nil {
		t.Fatalf("ConvertAutoWithOptions failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id in output: %s", result)
	}
}

func TestConvertAutoWithDelimiter(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := json2toon.ConvertAutoWithOptions([]byte(jsonl), json2toon.WithDelimiter('|'))
	if err != nil {
		t.Fatalf("ConvertAutoWithOptions failed: %v", err)
	}
	if !bytes.Contains(result, []byte("|")) {
		t.Errorf("expected pipe delimiter: %s", result)
	}
}

// --- Test stdin behavior by testing the library directly ---

func TestJSONLToon(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := json2toon.ConvertJSONL([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	// Should output tabular format
	if !bytes.Contains(result, []byte("[2]")) {
		t.Errorf("expected tabular format: %s", result)
	}
}

func TestJSONLNonTabular(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "other": "Bob"}`
	result, err := json2toon.ConvertJSONL([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	// Different keys → non-tabular with --- separators
	if !bytes.Contains(result, []byte("---")) {
		t.Errorf("expected separator for non-tabular: %s", result)
	}
}

func TestJSONLEmpty(t *testing.T) {
	result, err := json2toon.ConvertJSONL([]byte(""))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got: %s", result)
	}
}

func TestJSONCSimple(t *testing.T) {
	jsonc := `{"id": 1} // comment`
	result, err := json2toon.ConvertJSONC([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id in output: %s", result)
	}
}

func TestConvertFile(t *testing.T) {
	// Create a temp file
	tmp, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.WriteString(`{"test": 123}`)
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	result, err := json2toon.ConvertFile(tmp.Name())
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}
	if !bytes.Contains(result, []byte("test")) {
		t.Errorf("missing test in output: %s", result)
	}
}

func TestConvertAutoFile(t *testing.T) {
	// Create a temp file
	tmp, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.WriteString(`{"id": 1} // comment`)
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	result, err := json2toon.ConvertAutoFile(tmp.Name())
	if err != nil {
		t.Fatalf("ConvertAutoFile failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id in output: %s", result)
	}
}
