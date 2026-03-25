package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// --- mayContainComments ---

func TestMayContainComments(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte(`{"key": "value"}`), false},
		{[]byte(`{"url": "http://example.com"}`), false}, // :// inside URL is not a comment
		{[]byte(`{"key": "value"} // comment`), true},
		{[]byte(`{"key": "value"} /* comment */`), true},
		{[]byte(`// full line comment\n{"key": 1}`), true},
		{[]byte(`{"key": 1} /* multi\nline */`), true},
		{[]byte(`{"url": "https://example.com/api"}`), false},
	}
	for _, tt := range tests {
		result := mayContainComments(tt.data)
		if result != tt.expected {
			t.Errorf("mayContainComments(%q) = %v, want %v", string(tt.data), result, tt.expected)
		}
	}
}

// --- isJSONL ---

func TestIsJSONL(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte(`{"id": 1}`), false},                         // single line not JSONL
		{[]byte("{\"id\": 1}\n{\"id\": 2}"), true},           // 2 complete objects
		{[]byte("{\"id\": 1}\n  \n{\"id\": 2}"), true},       // with blank line
		{[]byte("{\"id\": 1}\nnot json\n{\"id\": 2}"), true}, // 2/3 valid > 50%
		{[]byte("1\n2\n3"), false},                           // primitives are not JSON objects/arrays
		{[]byte("[1,2]\n[3,4]"), true},                       // arrays
		{[]byte("invalid\n{\"id\": 1}"), false},              // mostly invalid
	}
	for _, tt := range tests {
		result := isJSONL(tt.data)
		if result != tt.expected {
			t.Errorf("isJSONL(%q) = %v, want %v", string(tt.data), result, tt.expected)
		}
	}
}

// --- formatValue ---

func TestFormatValue(t *testing.T) {
	tests := []struct {
		val      interface{}
		delim    string
		expected string
	}{
		{nil, ",", ""},
		{true, ",", "true"},
		{false, ",", "false"},
		{float64(42), ",", "42"},
		{float64(3.14), ",", "3"}, // %.0f truncates decimals
		{"hello", ",", "hello"},
		{"hello world", ",", `"hello world"`},
		{"a,b", ",", `"a,b"`},
		{`key: value`, ",", `"key: value"`}, // quoted: contains ":"
		{"say \"hi\"", ",", `"say "hi""`},   // quoted: contains `"`, ends with `"`
	}
	for _, tt := range tests {
		result := formatValue(tt.val, tt.delim)
		if result != tt.expected {
			t.Errorf("formatValue(%v, %q) = %q, want %q", tt.val, tt.delim, result, tt.expected)
		}
	}
}

// --- sameKeys ---

func TestSameKeys(t *testing.T) {
	if !sameKeys([]string{"a", "b"}, []string{"a", "b"}) {
		t.Error("same keys should return true")
	}
	if !sameKeys([]string{"a", "b"}, []string{"b", "a"}) {
		t.Error("same keys in different order should return true (sorts both)")
	}
	if sameKeys([]string{"a"}, []string{"a", "b"}) {
		t.Error("different length should return false")
	}
	if !sameKeys([]string{}, []string{}) {
		t.Error("empty slices should return true")
	}
}

// --- convertJSON ---

func TestConvertJSON(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	convertJSON([]byte(`{"id": 1}`), nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// --- convertJSONC ---

func TestConvertJSONC(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	convertJSONC([]byte(`{"id": 1} // comment`), nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

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

			expected, err := os.ReadFile(filepath.Join(projectRoot, tt.expected))
			if err != nil {
				t.Fatalf("Failed to read expected output %s: %v", tt.expected, err)
			}

			if !bytes.Equal(output, expected) {
				t.Errorf("Output mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s", tt.input, expected, output)
			}
		})
	}
}

// --- convertJSONL ---

func TestConvertJSONL(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	jsonl := `{"id": 1}
{"id": 2}`
	convertJSONL([]byte(jsonl), nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// --- mayContainComments ---

func TestMayContainCommentsTrue(t *testing.T) {
	tests := []string{
		`{"id": 1} // comment`,
		`{"id": 1} /* comment */`,
		`// full line comment`,
		`/* block comment */`,
	}
	for _, s := range tests {
		if !mayContainComments([]byte(s)) {
			t.Errorf("expected true for: %s", s)
		}
	}
}

func TestMayContainCommentsFalse(t *testing.T) {
	tests := []string{
		`{"id": 1}`,
		`[1, 2, 3]`,
		`"string with // inside"`,
		`{"url": "http://example.com"}`,
		`"a/*b"`,
	}
	for _, s := range tests {
		if mayContainComments([]byte(s)) {
			t.Errorf("expected false for: %s", s)
		}
	}
}

// --- sameKeys additional cases ---

func TestSameKeysAdditional(t *testing.T) {
	// Empty slices
	if !sameKeys([]string{}, []string{}) {
		t.Error("empty slices should return true")
	}
	// Different lengths
	if sameKeys([]string{"a"}, []string{"a", "b"}) {
		t.Error("different lengths should return false")
	}
	// Same content
	if !sameKeys([]string{"a", "b"}, []string{"a", "b"}) {
		t.Error("same content should return true")
	}
	// Same keys different order
	if !sameKeys([]string{"b", "a"}, []string{"a", "b"}) {
		t.Error("same keys different order should return true (sorts)")
	}
}
