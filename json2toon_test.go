package json2toon

import (
	"bytes"
	"os"
	"testing"
)

func TestConvertBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "boolean true",
			json:     `true`,
			expected: `true`,
		},
		{
			name:     "boolean false",
			json:     `false`,
			expected: `false`,
		},
		{
			name:     "integer",
			json:     `42`,
			expected: `42`,
		},
		{
			name:     "float",
			json:     `3.14`,
			expected: `3.14`,
		},
		{
			name:     "negative number",
			json:     `-42`,
			expected: `-42`,
		},
		{
			name:     "simple string",
			json:     `"hello"`,
			expected: `hello`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Convert([]byte(tt.json))
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("got %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestConvertObject(t *testing.T) {
	json := `{"id": 123, "name": "Ada", "active": true}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	expected := "id: 123\nname: Ada\nactive: true"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertNestedObject(t *testing.T) {
	json := `{"user": {"id": 123, "name": "Ada"}}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	expected := "user:\n  id: 123\n  name: Ada"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertPrimitiveArray(t *testing.T) {
	json := `{"tags": ["admin", "ops", "dev"]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Output may have slight format differences but should be valid TOON
	if !bytes.Contains(result, []byte("tags")) {
		t.Error("Missing tags key")
	}
	if !bytes.Contains(result, []byte("admin")) {
		t.Error("Missing admin value")
	}
}

func TestConvertTabularArray(t *testing.T) {
	json := `{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Output should contain tabular format
	if !bytes.Contains(result, []byte("users")) {
		t.Error("Missing users key")
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Error("Missing id field")
	}
}

func TestConvertEmptyObject(t *testing.T) {
	json := `{}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %q, want empty", string(result))
	}
}

func TestConvertEmptyArray(t *testing.T) {
	json := `{"items": []}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	expected := "items:"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertStringWithSpaces(t *testing.T) {
	json := `{"message": "hello world"}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	expected := "message: hello world"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertNumberNormalization(t *testing.T) {
	tests := []struct {
		json     string
		expected string
	}{
		{`1.0`, `1`},
		{`1.500`, `1.5`},
		{`1000000`, `1000000`},
		{`-0`, `0`},
		{`0.0`, `0`},
	}

	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			result, err := Convert([]byte(tt.json))
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("got %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestConvertJSONCComments(t *testing.T) {
	jsonc := `{
  // This is a comment
  "name": "Test",
  /* Multi-line
     comment */
  "value": 42
}`

	result, err := ConvertJSONC([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}

	expected := "name: Test\nvalue: 42"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertJSONCBlockComment(t *testing.T) {
	jsonc := `{"name": /* inline */ "Test", "value": 42}`
	result, err := ConvertJSONC([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}

	expected := "name: Test\nvalue: 42"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertJSONCNoStripInString(t *testing.T) {
	// Comments inside strings should NOT be stripped
	jsonc := `{"url": "http://example.com/path"}`
	result, err := ConvertJSONC([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}

	// URL should be preserved
	if !bytes.Contains(result, []byte("http://example.com/path")) {
		t.Error("URL not preserved correctly")
	}
}

func TestConvertWithOptions(t *testing.T) {
	json := `{"name": "Test"}`
	result, err := ConvertWithOptions([]byte(json), WithIndent(4))
	if err != nil {
		t.Fatalf("ConvertWithOptions failed: %v", err)
	}

	expected := "name: Test"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertFile(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `{"key": "value"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	result, err := ConvertFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}

	expected := "key: value"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestConvertFileToWriter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `{"key": "value"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	var buf bytes.Buffer
	err = ConvertFileToWriter(tmpFile.Name(), &buf)
	if err != nil {
		t.Fatalf("ConvertFileToWriter failed: %v", err)
	}

	expected := "key: value"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestConvertComplex(t *testing.T) {
	json := `{
  "config": {
    "database": {
      "host": "localhost",
      "port": 5432,
      "enabled": true
    },
    "tags": ["dev", "test"],
    "users": [
      {"id": 1, "name": "Alice"},
      {"id": 2, "name": "Bob"}
    ]
  }
}`

	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify the structure
	if !bytes.Contains(result, []byte("config")) {
		t.Error("Missing config key")
	}
	if !bytes.Contains(result, []byte("database")) {
		t.Error("Missing database key")
	}
	if !bytes.Contains(result, []byte("host")) {
		t.Error("Missing host")
	}
}

func TestConvertError(t *testing.T) {
	// Invalid JSON
	_, err := Convert([]byte(`{invalid`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestJSONCStripError(t *testing.T) {
	// Unterminated block comment
	_, err := ConvertJSONC([]byte(`{"key": /* unterminated`))
	if err == nil {
		t.Error("Expected error for unterminated block comment")
	}
}

func TestConvertString(t *testing.T) {
	json := `{"key": "value"}`
	result, err := ConvertString(json)
	if err != nil {
		t.Fatalf("ConvertString failed: %v", err)
	}

	expected := "key: value"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestConvertJSONCString(t *testing.T) {
	jsonc := `{"key": "value"} // comment`
	result, err := ConvertJSONCString(jsonc)
	if err != nil {
		t.Fatalf("ConvertJSONCString failed: %v", err)
	}

	expected := "key: value"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestConvertLargeFile(t *testing.T) {
	// Create a moderately large JSON to test streaming
	var buf bytes.Buffer
	buf.WriteString(`{"items": [`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			buf.WriteString(`,`)
		}
		buf.WriteString(`{"id":`)
		buf.WriteString(string(rune('0' + i%10)))
		buf.WriteString(`}`)
	}
	buf.WriteString(`]}`)

	result, err := Convert(buf.Bytes())
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}

func TestEncoderDirect(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(map[string]interface{}{
		"name":  "test",
		"count": float64(42),
	})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	enc.Flush()

	if buf.Len() == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestMixedArray(t *testing.T) {
	json := `{"items": [{"id": 1}, "plain", {"id": 2}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !bytes.Contains(result, []byte("items")) {
		t.Error("Missing items key")
	}
}

func TestNestedArray(t *testing.T) {
	json := `{"matrix": [[1, 2], [3, 4]]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !bytes.Contains(result, []byte("matrix")) {
		t.Error("Missing matrix key")
	}
}

func TestWithDelimiter(t *testing.T) {
	json := `{"users": [{"id": 1, "name": "Alice"}]}`
	result, err := ConvertWithOptions([]byte(json), WithDelimiter('|'))
	if err != nil {
		t.Fatalf("ConvertWithOptions failed: %v", err)
	}

	if !bytes.Contains(result, []byte("users")) {
		t.Error("Missing users key")
	}
}

func TestWithKeyFolding(t *testing.T) {
	json := `{"a": {"b": {"c": 1}}}`
	result, err := ConvertWithOptions([]byte(json), WithKeyFolding("safe"))
	if err != nil {
		t.Fatalf("ConvertWithOptions failed: %v", err)
	}

	// Key folding is optional, just check it works
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}

func TestWithStrict(t *testing.T) {
	json := `{"key": "value"}`
	result, err := ConvertWithOptions([]byte(json), WithStrict(true))
	if err != nil {
		t.Fatalf("ConvertWithOptions failed: %v", err)
	}

	if !bytes.Contains(result, []byte("key")) {
		t.Error("Missing key")
	}
}

func TestConvertFileWithOptions(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `{"name": "Test"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	result, err := ConvertFileWithOptions(tmpFile.Name(), WithIndent(4))
	if err != nil {
		t.Fatalf("ConvertFileWithOptions failed: %v", err)
	}

	if !bytes.Contains(result, []byte("Test")) {
		t.Error("Missing content")
	}
}

func TestConvertJSONCWithOptions(t *testing.T) {
	jsonc := `{"name": "Test"} // comment`
	result, err := ConvertJSONCWithOptions([]byte(jsonc), WithStrict(false))
	if err != nil {
		t.Fatalf("ConvertJSONCWithOptions failed: %v", err)
	}

	if !bytes.Contains(result, []byte("Test")) {
		t.Error("Missing content")
	}
}

func TestStreamingConverter(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)

	err := conv.ConvertJSON([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}

	err = conv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestStreamingConverterWithOptions(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultConverterOptions()
	opts.Encoder.Indent = 4
	conv := NewConverterWithOptions(&buf, opts)

	err := conv.ConvertJSON([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}

	err = conv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestStripCommentsInString(t *testing.T) {
	jsonc := `{"url": "http://example.com/api"}`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}

	expected := `{"url": "http://example.com/api"}`
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestStripCommentsEmpty(t *testing.T) {
	jsonc := `{"key": "value"}`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}

	expected := `{"key": "value"}`
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestNormalizeNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.0, "0"},
		{1.0, "1"},
		{-0.0, "0"},
		{1.5, "1.5"},
		{1.500, "1.5"},
	}

	for _, tt := range tests {
		result := normalizeNumber(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeNumber(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsPrimitive(t *testing.T) {
	tests := []struct {
		val      interface{}
		expected bool
	}{
		{nil, true},
		{true, true},
		{float64(1), true},
		{"string", true},
		{[]interface{}{}, false},
		{map[string]interface{}{}, false},
	}

	for _, tt := range tests {
		result := isPrimitive(tt.val)
		if result != tt.expected {
			t.Errorf("isPrimitive(%v) = %v, want %v", tt.val, result, tt.expected)
		}
	}
}

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		s        string
		delim    rune
		expected bool
	}{
		{"", ',', true},
		{"hello", ',', false},
		{"true", ',', true},
		{"42", ',', true},
		{"hello:world", ',', true},
		{`"quoted"`, ',', true},
	}

	for _, tt := range tests {
		result := needsQuoting(tt.s, tt.delim)
		if result != tt.expected {
			t.Errorf("needsQuoting(%q, %q) = %v, want %v", tt.s, tt.delim, result, tt.expected)
		}
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `hello`},
		{`"quoted"`, `\"quoted\"`},
		{"line\nbreak", `line\nbreak`},
		{"tab\there", `tab\there`},
	}

	for _, tt := range tests {
		result := escapeString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultEncoderOptions(t *testing.T) {
	opts := DefaultEncoderOptions()
	if opts.Indent != 2 {
		t.Errorf("DefaultIndent = %d, want 2", opts.Indent)
	}
	if opts.Delimiter != ',' {
		t.Errorf("DefaultDelimiter = %q, want ','", opts.Delimiter)
	}
}

func TestDefaultConverterOptions(t *testing.T) {
	opts := DefaultConverterOptions()
	if !opts.Strict {
		t.Error("DefaultStrict should be true")
	}
}
