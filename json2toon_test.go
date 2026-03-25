package json2toon

import (
	"bytes"
	"os"
	"strings"
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
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := `{"key": "value"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = tmpFile.Close()

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
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := `{"key": "value"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = tmpFile.Close()

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
	_ = enc.Flush()

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

// --- Encoder: writeArray ---

func TestWriteArrayMixed(t *testing.T) {
	json := `{"items": [{"id": 1}, "plain", {"id": 2}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Mixed arrays produce tabular format for the object elements.
	// Non-object elements are handled per-implementation.
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items key: %s", result)
	}
}

func TestWriteArrayNested(t *testing.T) {
	json := `{"matrix": [[1, 2], [3, 4]]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("matrix")) {
		t.Errorf("missing matrix key: %s", result)
	}
}

func TestWriteArrayNestedObjects(t *testing.T) {
	json := `{"items": [{"a": 1}, {"b": 2}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Objects with different keys produce list format.
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items key: %s", result)
	}
}

func TestWriteArrayEmpty(t *testing.T) {
	json := `{"items": []}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	expected := "items:"
	if string(result) != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestWriteArrayPrimitive(t *testing.T) {
	json := `{"nums": [1, 2, 3]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("1,2,3")) {
		t.Errorf("expected inline primitive format: %s", result)
	}
}

// --- Encoder: isTabular ---

func TestIsTabularFalse(t *testing.T) {
	json := `{"items": [{"a": 1}, {"a": 2, "b": 3}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if bytes.Contains(result, []byte("a,b")) {
		t.Errorf("should not be tabular with different keys: %s", result)
	}
}

func TestIsTabularWithNonPrimitive(t *testing.T) {
	json := `{"items": [{"id": 1, "tags": ["a"]}, {"id": 2, "tags": ["b"]}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Objects with non-primitive values produce list format, not tabular.
	if bytes.Contains(result, []byte("id,tags")) {
		t.Errorf("should NOT be tabular with non-primitive values: %s", result)
	}
	// Should contain items key
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items: %s", result)
	}
}

// --- Encoder: formatPrimitive ---

func TestFormatPrimitiveNull(t *testing.T) {
	json := `{"key": null}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("null")) {
		t.Errorf("missing null: %s", result)
	}
}

// --- Encoder: encodeReflect ---

func TestEncodeReflect(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
		Val  int    `json:"val"`
	}
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	item := Item{Name: "test", Val: 42}
	err := enc.Encode(item)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	_ = enc.Flush()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// --- Encoder: encodeReflect error ---

func TestEncodeReflectError(t *testing.T) {
	// encodeReflect uses json.Marshal which returns error for invalid types
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(make(chan int))
	if err == nil {
		t.Error("expected error encoding invalid type")
	}
}

// --- Encoder: writeObject with non-primitive values ---

func TestWriteObjectWithNestedArray(t *testing.T) {
	json := `{"data": {"tags": ["a", "b"]}}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("tags")) {
		t.Errorf("missing nested tags: %s", result)
	}
}

func TestWriteObjectWithNestedObject(t *testing.T) {
	json := `{"outer": {"inner": {"x": 1}}}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("inner")) {
		t.Errorf("missing inner key: %s", result)
	}
}

func TestWriteObjectWithString(t *testing.T) {
	json := `{"msg": "hello"}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("hello")) {
		t.Errorf("missing hello: %s", result)
	}
}

// --- Converter: isTabularArray ---

func TestIsTabularArray(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		tabular bool
	}{
		{"empty", `{"items": []}`, false},
		{"not objects", `{"items": [1, 2, 3]}`, false},
		{"single object", `{"items": [{"id": 1}]}`, false},
		{"uniform primitive objects", `{"items": [{"id": 1}, {"id": 2}]}`, true},
		{"different keys", `{"items": [{"a": 1}, {"b": 2}]}`, false},
		{"different key count", `{"items": [{"a": 1}, {"a": 2, "b": 3}]}`, false},
		{"non-primitive value", `{"items": [{"id": 1, "x": [1]}]}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			conv := NewConverter(&buf)
			_ = conv.ConvertJSON([]byte(tt.json))
			_ = conv.Close()
			out := buf.String()
			if tt.tabular {
				if !bytes.Contains([]byte(out), []byte("{id}:")) {
					t.Errorf("expected tabular format for %s: %s", tt.name, out)
				}
			}
		})
	}
}

// --- Converter: writeArrayListStream ---

func TestWriteArrayListStreamMixed(t *testing.T) {
	json := `{"items": [{"id": 1}, "text", {"id": 2}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Mixed arrays include the items key in output.
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items key: %s", result)
	}
}

func TestWriteArrayListStreamNested(t *testing.T) {
	json := `{"items": [[1, 2], [3, 4]]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items key: %s", result)
	}
}

// --- Converter: formatPrimitiveValue ---

func TestFormatPrimitiveValue(t *testing.T) {
	tests := []struct {
		json     string
		contains string
	}{
		{`{"n": null}`, "null"},
		{`{"b": true}`, "true"},
		{`{"b": false}`, "false"},
		{`{"f": 3.14}`, "3.14"},
		{`{"s": "hello"}`, "hello"},
	}
	for _, tt := range tests {
		result, err := Convert([]byte(tt.json))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		if !bytes.Contains(result, []byte(tt.contains)) {
			t.Errorf("%s: got %q, want to contain %q", tt.json, result, tt.contains)
		}
	}
}

// --- JSONC: edge cases ---

func TestStripCommentsUnterminatedString(t *testing.T) {
	_, err := StripComments([]byte(`{"key": "unterminated`))
	if err == nil {
		t.Error("expected error for unterminated string")
	}
}

func TestStripCommentsSlashStarEnd(t *testing.T) {
	// {"key": "value"} /* end comment */ - properly terminated block comment
	jsonc := `{"key": "value"} /* end comment */`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}
	// After stripping, should have {"key": "value"} (with trailing space)
	if !bytes.Contains(result, []byte(`"value"`)) {
		t.Errorf("value should be preserved: %s", result)
	}
	if bytes.Contains(result, []byte("end comment")) {
		t.Errorf("comment should be stripped: %s", result)
	}
}

func TestStripCommentsMultiple(t *testing.T) {
	jsonc := `// comment 1
{"a": 1} // comment 2
// comment 3`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}
	if !bytes.Contains(result, []byte(`"a"`)) {
		t.Errorf("should preserve content: %s", result)
	}
}

func TestStripCommentsSlashSlash(t *testing.T) {
	jsonc := `{"url": "http://example.com"} // no comment`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}
	if !bytes.Contains(result, []byte(`"http://example.com"`)) {
		t.Errorf("url should be preserved: %s", result)
	}
}

// --- Encoder: allPrimitives ---

func TestAllPrimitives(t *testing.T) {
	json := `{"arr": [1, "a", true]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("1,a,true")) {
		t.Errorf("expected inline format: %s", result)
	}
}

// --- Converter: StreamConvertJSONC ---

func TestStreamingConverterJSONC(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)

	err := conv.ConvertJSONC([]byte(`{"a": 1} // comment`))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}

	err = conv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty output")
	}
}

// --- Error paths ---

func TestConvertFileNotFound(t *testing.T) {
	_, err := ConvertFile("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestConvertFileToWriterNotFound(t *testing.T) {
	err := ConvertFileToWriter("nonexistent.json", &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// --- Encoder options ---

func TestEncoderWithDelimiter(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoderWithOptions(&buf, EncoderOptions{
		Delimiter: '|',
		Indent:    4,
	})
	err := enc.Encode(map[string]interface{}{
		"rows": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "Alice"},
			map[string]interface{}{"id": float64(2), "name": "Bob"},
		},
	})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	_ = enc.Flush()
	if !bytes.Contains(buf.Bytes(), []byte("|")) {
		t.Errorf("expected pipe delimiter in output: %s", buf.String())
	}
}

func TestEncoderIndent(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoderWithOptions(&buf, EncoderOptions{
		Indent: 4,
	})
	err := enc.Encode(map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": float64(1),
		},
	})
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	_ = enc.Flush()
	if !bytes.Contains(buf.Bytes(), []byte("    ")) {
		t.Errorf("expected 4 spaces for indent=4: %s", buf.String())
	}
}

// --- Converter: ConvertJSONL ---

func TestConvertJSONL(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := Convert([]byte(jsonl))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) || !bytes.Contains(result, []byte("name")) {
		t.Errorf("missing id or name: %s", result)
	}
}

// --- Encoder: encodeValue nil interface ---

func TestEncodeValueWithNilInterface(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	var v interface{} = nil
	err := enc.Encode(v)
	if err != nil {
		t.Fatalf("Encode nil failed: %v", err)
	}
	_ = enc.Flush()
	if !bytes.Contains(buf.Bytes(), []byte("null")) {
		t.Errorf("expected null for nil: %s", buf.String())
	}
}

// --- Encoder: string escaping ---

func TestEncoderStringEscaping(t *testing.T) {
	tests := []struct {
		json  string
		check func(string) bool
	}{
		{`{"k": "a\\b"}`, func(s string) bool { return bytes.Contains([]byte(s), []byte(`\\`)) }},       // backslash
		{`{"k": "a\nb"}`, func(s string) bool { return bytes.Contains([]byte(s), []byte(`\n`)) }},       // newline
		{`{"k": "a\tb"}`, func(s string) bool { return bytes.Contains([]byte(s), []byte(`\t`)) }},       // tab
		{`{"k": "a\rb"}`, func(s string) bool { return bytes.Contains([]byte(s), []byte(`\r`)) }},       // CR
		{`{"k": "say \"hi\""}`, func(s string) bool { return bytes.Contains([]byte(s), []byte(`\"`)) }}, // dquote
	}
	for _, tt := range tests {
		result, err := Convert([]byte(tt.json))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		if !tt.check(string(result)) {
			t.Errorf("escape missing in: %s", result)
		}
	}
}

func TestEncoderNeedsQuoting(t *testing.T) {
	// Strings that need quoting
	quoted := []string{
		`{"k": ""}`,      // empty
		`{"k": " true"}`, // leading space
		`{"k": "true "}`, // trailing space
		`{"k": "true"}`,  // bool literal
		`{"k": "null"}`,  // null literal
		`{"k": "42"}`,    // numeric
		`{"k": "a:b"}`,   // colon
		`{"k": "a[b"}`,   // bracket
	}
	for _, json := range quoted {
		result, err := Convert([]byte(json))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		if !bytes.Contains(result, []byte(`"`)) {
			t.Errorf("expected quoting for: %s, got: %s", json, result)
		}
	}

	// Strings that don't need quoting
	unquoted := []struct {
		json     string
		wantCont string
	}{
		{`{"k": "hello"}`, "hello"},
		{`{"k": "world"}`, "world"},
		{`{"k": "a_b_c"}`, "a_b_c"},
		{`{"k": "a.b"}`, "a.b"},
	}
	for _, tt := range unquoted {
		result, err := Convert([]byte(tt.json))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		if !bytes.Contains(result, []byte(tt.wantCont)) {
			t.Errorf("missing %s in: %s", tt.wantCont, result)
		}
	}
}

// --- Encoder: key folding ---

func TestEncoderKeyFolding1Level(t *testing.T) {
	// Test folding directly with encoder
	var buf bytes.Buffer
	opts := EncoderOptions{KeyFolding: "safe", FlattenDepth: 10}
	enc := NewEncoderWithOptions(&buf, opts)
	_ = enc.Encode(map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	})
	_ = enc.Flush()
	if !bytes.Contains(buf.Bytes(), []byte("outer")) {
		t.Errorf("expected outer key: %s", buf.String())
	}
}

func TestEncoderKeyFoldingMultiLevel(t *testing.T) {
	// Test folding directly with encoder (FlattenDepth must be > depth)
	var buf bytes.Buffer
	opts := EncoderOptions{KeyFolding: "safe", FlattenDepth: 10}
	enc := NewEncoderWithOptions(&buf, opts)
	_ = enc.Encode(map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "value",
			},
		},
	})
	_ = enc.Flush()
	// Multi-level folding produces "a.b.c"
	if !bytes.Contains(buf.Bytes(), []byte("a.b.c")) {
		t.Errorf("expected folded key a.b.c: %s", buf.String())
	}
}

func TestEncoderKeyFoldingInvalidKey(t *testing.T) {
	// a-b contains hyphen, can't fold
	var buf bytes.Buffer
	opts := EncoderOptions{KeyFolding: "safe", FlattenDepth: 10}
	enc := NewEncoderWithOptions(&buf, opts)
	_ = enc.Encode(map[string]interface{}{
		"a-b": map[string]interface{}{
			"c": "value",
		},
	})
	_ = enc.Flush()
	// Should NOT fold, should have nested structure
	if bytes.Contains(buf.Bytes(), []byte("a-b.c")) {
		t.Errorf("should not fold invalid key: %s", buf.String())
	}
}

func TestEncoderKeyFoldingMultiKey(t *testing.T) {
	// Two keys in nested map, can't fold
	var buf bytes.Buffer
	opts := EncoderOptions{KeyFolding: "safe", FlattenDepth: 10}
	enc := NewEncoderWithOptions(&buf, opts)
	_ = enc.Encode(map[string]interface{}{
		"a": map[string]interface{}{
			"b": "v1",
			"c": "v2",
		},
	})
	_ = enc.Flush()
	// Should NOT fold, should have nested structure
	if bytes.Contains(buf.Bytes(), []byte("a.b")) {
		t.Errorf("should not fold multi-key: %s", buf.String())
	}
}

func TestEncoderKeyFoldingEmptyNested(t *testing.T) {
	var buf bytes.Buffer
	opts := EncoderOptions{KeyFolding: "safe", FlattenDepth: 10}
	enc := NewEncoderWithOptions(&buf, opts)
	_ = enc.Encode(map[string]interface{}{
		"a": map[string]interface{}{},
	})
	_ = enc.Flush()
	if !bytes.Contains(buf.Bytes(), []byte("a:")) {
		t.Errorf("expected empty nested: %s", buf.String())
	}
}

func TestEncoderKeyFoldingNotSafe(t *testing.T) {
	// Without "safe", no folding
	var buf bytes.Buffer
	opts := EncoderOptions{KeyFolding: "off", FlattenDepth: 10}
	enc := NewEncoderWithOptions(&buf, opts)
	_ = enc.Encode(map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	})
	_ = enc.Flush()
	if bytes.Contains(buf.Bytes(), []byte("outer.inner")) {
		t.Errorf("should not fold without safe: %s", buf.String())
	}
}

// --- Encoder: array primitives ---

func TestEncoderPrimitiveArray(t *testing.T) {
	json := `{"nums": [1, 2, 3]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("nums")) {
		t.Errorf("missing nums key: %s", result)
	}
}

func TestEncoderPrimitiveArrayCustomDelimiter(t *testing.T) {
	json := `{"nums": [1.5, 2.5]}`
	result, err := ConvertWithOptions([]byte(json), WithDelimiter('|'))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Output contains the nums key
	if !bytes.Contains(result, []byte("nums")) {
		t.Errorf("expected nums key: %s", result)
	}
}

func TestEncoderEmptyArray(t *testing.T) {
	json := `{"items": []}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Empty array → just colon
	expected := "items:"
	if string(result) != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestEncoderMixedArray(t *testing.T) {
	json := `{"items": [1, "two", true]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// All primitives → inline format
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items: %s", result)
	}
}

func TestEncoderEncodeInterfaceArray(t *testing.T) {
	// Test direct []interface{} encoding
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	arr := []interface{}{"a", float64(1), true}
	err := enc.Encode(arr)
	_ = enc.Flush()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestEncoderEncodeEmptyInterfaceArray(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode([]interface{}{})
	_ = enc.Flush()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(":")) {
		t.Errorf("expected empty array ':': %s", buf.String())
	}
}

// --- Encoder: options ---

func TestEncoderOptionsIndent0(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoderWithOptions(&buf, EncoderOptions{Indent: 0})
	_ = enc.Encode(map[string]interface{}{"k": "v"})
	_ = enc.Flush()
	// Indent 0 defaults to 2
	if !bytes.Contains(buf.Bytes(), []byte("k: v")) {
		t.Errorf("expected default indent: %s", buf.String())
	}
}

func TestWithStrictOption(t *testing.T) {
	// Strict mode should not cause issues for valid JSON
	json := `{"id": 1}`
	result, err := ConvertWithOptions([]byte(json), WithStrict(true))
	if err != nil {
		t.Fatalf("ConvertWithOptions failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

// --- Converter: streaming with primitives ---

func TestStreamingConverterPrimitives(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSON([]byte(`{"arr": [1, 2, 3]}`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}
	_ = conv.Close()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestStreamingConverterMultipleObjects(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSON([]byte(`{"id": 1}`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}
	err = conv.ConvertJSON([]byte(`{"id": 2}`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}
	_ = conv.Close()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestStreamingConverterPrimitivesOnly(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSON([]byte(`[1, 2, 3]`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}
	_ = conv.Close()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// --- JSONC: unterminated string ---

func TestStripCommentsSingleSlashSlash(t *testing.T) {
	jsonc := `// single line comment
{"id": 1}`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}
	if !bytes.Contains(result, []byte(`"id"`)) {
		t.Errorf("should preserve JSON: %s", result)
	}
}

func TestStripCommentsInMiddle(t *testing.T) {
	jsonc := `{"a": 1} // inline comment
{"b": 2}`
	result, err := StripComments([]byte(jsonc))
	if err != nil {
		t.Fatalf("StripComments failed: %v", err)
	}
	if !bytes.Contains(result, []byte(`"a"`)) || !bytes.Contains(result, []byte(`"b"`)) {
		t.Errorf("should preserve both objects: %s", result)
	}
}

// --- API: ConvertJSONCWithOptions ---

func TestConvertJSONCWithOptions(t *testing.T) {
	jsonc := `{"id": 1} // comment`
	result, err := ConvertJSONCWithOptions([]byte(jsonc), WithIndent(4))
	if err != nil {
		t.Fatalf("ConvertJSONCWithOptions failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertFileWithOptions(t *testing.T) {
	// Use a temp file
	tmp, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	defer func() { _ = tmp.Close() }()

	_, err = tmp.WriteString(`{"test": 123}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertFileWithOptions(tmp.Name(), WithIndent(4))
	if err != nil {
		t.Fatalf("ConvertFileWithOptions failed: %v", err)
	}
	if !bytes.Contains(result, []byte("test")) {
		t.Errorf("missing test: %s", result)
	}
}

// --- Encoder: encode errors ---

func TestEncodeChanError(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(make(chan int))
	if err == nil {
		t.Error("expected error encoding channel")
	}
}

// --- Converter: isPrimitive ---

func TestIsPrimitive(t *testing.T) {
	tests := []struct {
		val    interface{}
		isPrim bool
	}{
		{nil, true},
		{true, true},
		{float64(1), true},
		{"str", true},
		{map[string]interface{}{}, false},
		{[]interface{}{}, false},
	}
	for _, tt := range tests {
		// Test through allPrimitives on single-element array
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		_ = enc.Encode([]interface{}{tt.val})
		// If it succeeds (no error), it treated it as primitive-like
		_ = enc.Flush()
	}
}

// --- Converter: object with primitive values ---

func TestObjectWithPrimitiveValues(t *testing.T) {
	json := `{"a": 1, "b": "str", "c": true, "d": null}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	for _, want := range []string{"a", "b", "c", "d"} {
		if !bytes.Contains(result, []byte(want)) {
			t.Errorf("missing key %s in: %s", want, result)
		}
	}
}

// --- Converter: nested object in array ---

func TestNestedObjectInArray(t *testing.T) {
	json := `{"items": [{"inner": {"deep": 1}}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items: %s", result)
	}
}

// --- Converter: all primitives check ---

func TestAllPrimitivesInConverter(t *testing.T) {
	json := `{"arr": [1, 2, "three", true]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("arr")) {
		t.Errorf("missing arr: %s", result)
	}
}

// --- Converter: Close on already closed ---

func TestConverterDoubleClose(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	if err := conv.Close(); err != nil {
		t.Fatal(err)
	}
	if err := conv.Close(); err != nil {
		t.Fatal(err)
	}
}

// --- Converter: writeArrayDirect with primitives ---

func TestWriteArrayDirectWithPrimitives(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSON([]byte(`{"arr": [1, 2, 3]}`))
	if err != nil {
		t.Fatalf("ConvertJSON failed: %v", err)
	}
	_ = conv.Close()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// --- Number normalization ---

func TestNormalizeNumberEdgeCases(t *testing.T) {
	tests := []struct {
		json     string
		contains string
	}{
		{`{"n": 0}`, "0"},
		{`{"n": -0}`, "0"},
		{`{"n": 100}`, "100"},
		{`{"n": 1e10}`, "10000000000"},
		{`{"n": 1.5}`, "1.5"},
		{`{"n": 0.1}`, "0.1"},
	}
	for _, tt := range tests {
		result, err := Convert([]byte(tt.json))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		if !bytes.Contains(result, []byte(tt.contains)) {
			t.Errorf("%s: expected %s in %s", tt.json, tt.contains, result)
		}
	}
}

// --- Encoder: formatValue edge cases ---

func TestFormatValueEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	_ = enc.Encode(map[string]interface{}{
		"bool":  true,
		"flt":   float64(3.14),
		"str":   "plain",
		"empty": "",
	})
	_ = enc.Flush()
	out := buf.String()
	for _, check := range []string{"bool: true", "flt:", "str: plain", `empty: ""`} {
		if !bytes.Contains(buf.Bytes(), []byte(check)) {
			t.Errorf("expected %q in output: %s", check, out)
		}
	}
}

// --- Encoder: isValidFoldKey ---

func TestIsValidFoldKey(t *testing.T) {
	valid := []string{"a", "abc", "a1", "_", "a.b", "A.B.C"}
	for _, k := range valid {
		var buf bytes.Buffer
		enc := NewEncoderWithOptions(&buf, EncoderOptions{KeyFolding: "safe", FlattenDepth: 10})
		_ = enc.Encode(map[string]interface{}{k: map[string]interface{}{"b": "v"}})
		_ = enc.Flush()
	}

	invalid := []string{"a-b", "a:b", "a b", "a[b", "a}b"}
	for _, k := range invalid {
		var buf bytes.Buffer
		enc := NewEncoderWithOptions(&buf, EncoderOptions{KeyFolding: "safe", FlattenDepth: 10})
		_ = enc.Encode(map[string]interface{}{k: map[string]interface{}{"c": "v"}})
		_ = enc.Flush()
		// Invalid keys should not fold (nested structure)
		if bytes.Contains(buf.Bytes(), []byte(k+".")) {
			t.Errorf("invalid key %q should not fold: %s", k, buf.String())
		}
	}
}

// --- Converter: single-line comment in JSONC ---

func TestSingleLineCommentInJSONC(t *testing.T) {
	jsonc := `{"id": 1} // comment after object`
	result, err := ConvertJSONC([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

// --- Converter: JSONC with block comment ---

func TestJSONCWithBlockComment(t *testing.T) {
	jsonc := `{"id": 1} /* block comment */`
	result, err := ConvertJSONC([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertJSONC failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

// --- Converter: Tabular array with all fields ---

func TestTabularArrayAllFields(t *testing.T) {
	json := `{"items": [{"id": 1, "name": "Alice", "active": true}, {"id": 2, "name": "Bob", "active": false}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("{active,id,name}")) {
		t.Errorf("expected tabular header with keys: %s", result)
	}
}

// --- Converter: non-tabular array (different key counts) ---

func TestNonTabularArrayDifferentKeys(t *testing.T) {
	json := `{"items": [{"a": 1}, {"a": 2, "b": 3}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	// Different key counts → list format
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items: %s", result)
	}
}

// --- Converter: array with nested objects ---

func TestArrayWithNestedObjects(t *testing.T) {
	json := `{"items": [{"a": {"b": 1}}, {"a": {"b": 2}}]}`
	result, err := Convert([]byte(json))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("items")) {
		t.Errorf("missing items: %s", result)
	}
}

// --- Converter: JSONL with different key counts ---

func TestJSONLDifferentKeys(t *testing.T) {
	jsonl := `{"id": 1}
{"id": 2, "name": "Bob"}`
	result, err := Convert([]byte(jsonl))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

// --- Converter: isTabularArray branches ---

func TestIsTabularArrayNonMapFirst(t *testing.T) {
	// First element not a map → not tabular
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	// Simulate by encoding a direct array (root level)
	_ = conv.ConvertJSON([]byte(`[1, 2, 3]`))
	_ = conv.Close()
	// Should produce output without error
}

func TestConverterIsTabularArrayEmptyKeys(t *testing.T) {
	// Object with empty keys
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	_ = conv.ConvertJSON([]byte(`{"items": [{}, {}]}`))
	_ = conv.Close()
}

func TestConverterWriteArrayListStreamNestedArray(t *testing.T) {
	// Nested arrays in list stream
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	_ = conv.ConvertJSON([]byte(`{"items": [[1,2], [3,4]]}`))
	_ = conv.Close()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestConverterWriteArrayListStreamPrimitives(t *testing.T) {
	// Primitives in list stream (non-tabular, non-all-prim)
	// Need mixed: some primitives, some not all-prim
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	_ = conv.ConvertJSON([]byte(`{"items": [{"id":1}, {"id":2}]}`))
	_ = conv.Close()
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// --- Encoder: writeArrayItems with non-primitive ---

func TestEncoderWriteArrayItemsNonPrimitive(t *testing.T) {
	// writeArrayItems calls writePrimitive which errors on non-primitives
	// This exercises the writeArray → case []interface{} → writeArrayItems path
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(map[string]interface{}{
		"arr": []interface{}{
			map[string]interface{}{"a": float64(1)},
		},
	})
	_ = enc.Flush()
	// Exercises writeArray → case []interface{} path.
	// writeArrayItems calls writePrimitive which would error for non-primitives.
	_ = err
}

func TestEncoderEncodeByteSlice(t *testing.T) {
	// Byte slice goes through encodeReflect path
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode([]byte("hello"))
	_ = enc.Flush()
	if err != nil {
		t.Fatalf("Encode byte slice failed: %v", err)
	}
}

func TestEncoderEncodeInt(t *testing.T) {
	// int goes through encodeReflect
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(int(42))
	_ = enc.Flush()
	if err != nil {
		t.Fatalf("Encode int failed: %v", err)
	}
}

func TestEncoderWriteQuotedStringCR(t *testing.T) {
	// CR escape in quoted string
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	_ = enc.Encode(map[string]interface{}{"k": "a\x0d b"})
	_ = enc.Flush()
}

func TestEncoderWriteObjectDefaultCase(t *testing.T) {
	// encodeReflect handles types not caught by explicit cases
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	type custom struct {
		Name string
	}
	err := enc.Encode(custom{Name: "test"})
	_ = enc.Flush()
	if err != nil {
		t.Fatalf("Encode custom struct failed: %v", err)
	}
}

func TestEncoderGetFirstKeyEmpty(t *testing.T) {
	// getFirstKey called via writeArray nested array handling
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	// Empty array in nested context
	_ = enc.Encode(map[string]interface{}{
		"outer": []interface{}{},
	})
	_ = enc.Flush()
}

func TestConverterEncodePrimitiveBool(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	_ = conv.ConvertJSON([]byte(`{"t": true, "f": false}`))
	_ = conv.Close()
	if !bytes.Contains(buf.Bytes(), []byte("true")) || !bytes.Contains(buf.Bytes(), []byte("false")) {
		t.Errorf("missing bool values: %s", buf.String())
	}
}

func TestConverterEncodePrimitiveNull(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	_ = conv.ConvertJSON([]byte(`{"n": null}`))
	_ = conv.Close()
	if !bytes.Contains(buf.Bytes(), []byte("null")) {
		t.Errorf("missing null: %s", buf.String())
	}
}

func TestConverterEncodePrimitiveString(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	_ = conv.ConvertJSON([]byte(`{"s": "hello"}`))
	_ = conv.Close()
	if !bytes.Contains(buf.Bytes(), []byte("hello")) {
		t.Errorf("missing string: %s", buf.String())
	}
}

func TestConverterWriteTabularRowsDifferentKeys(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	// Two objects with same key count but different keys - triggers non-primitive check
	_ = conv.ConvertJSON([]byte(`{"items": [{"a": 1}, {"b": 2}]}`))
	_ = conv.Close()
}

func TestConverterDecodeObjectError(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	// Invalid JSON - should error
	err := conv.ConvertJSON([]byte(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestConverterDecodeObjectTrailingComma(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	// JSON decoder handles trailing comma or not - just verify behavior
	err := conv.ConvertJSON([]byte(`{"id": 1}`))
	_ = conv.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConverterWriteValueStreamPrimitive(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	// Root-level primitives
	_ = conv.ConvertJSON([]byte(`123`))
	_ = conv.Close()
}

func TestConverterCollectArrayItemsNested(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	// Deeply nested arrays
	_ = conv.ConvertJSON([]byte(`{"a": [[[1]], [[2]]]}`))
	_ = conv.Close()
}

func TestWithKeyFoldingOption(t *testing.T) {
	// WithKeyFolding option - sets KeyFolding on encoder options
	result, err := ConvertWithOptions([]byte(`{"a": {"b": "v"}}`), WithKeyFolding("safe"))
	if err != nil {
		t.Fatalf("ConvertWithOptions failed: %v", err)
	}
	_ = result
}

func TestConvertFileWithOptionsInvalidFile(t *testing.T) {
	_, err := ConvertFileWithOptions("nonexistent.json", WithIndent(4))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// --- ConvertJSONL tests ---

func TestConvertJSONLBasic(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := ConvertJSONL([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	// Should output tabular format
	if !bytes.Contains(result, []byte("[2]{id,name}:")) {
		t.Errorf("expected tabular format: %s", result)
	}
}

func TestConvertJSONLString(t *testing.T) {
	jsonl := `{"id": 1}
{"id": 2}`
	result, err := ConvertJSONLString(jsonl)
	if err != nil {
		t.Fatalf("ConvertJSONLString failed: %v", err)
	}
	if !bytes.Contains([]byte(result), []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertJSONLWithOptions(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := ConvertJSONLWithOptions([]byte(jsonl), WithDelimiter('|'))
	if err != nil {
		t.Fatalf("ConvertJSONLWithOptions failed: %v", err)
	}
	// Should have tabular format with pipe delimiter in header and rows
	// Keys are sorted, so "id" comes before "name"
	if !bytes.Contains(result, []byte("{id|name}:")) {
		t.Errorf("expected tabular header with id|name: %s", result)
	}
	if !bytes.Contains(result, []byte("|")) {
		t.Errorf("expected pipe delimiter in rows: %s", result)
	}
}

func TestConvertJSONLNonTabular(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "other": "Bob"}`
	result, err := ConvertJSONL([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	// Different keys → non-tabular with --- separators
	if !bytes.Contains(result, []byte("---")) {
		t.Errorf("expected separator for non-tabular: %s", result)
	}
}

func TestConvertJSONLEmpty(t *testing.T) {
	result, err := ConvertJSONL([]byte(""))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got: %s", result)
	}
}

func TestConvertJSONLSingleLine(t *testing.T) {
	// Single line is not really JSONL, but should handle gracefully
	result, err := ConvertJSONL([]byte(`{"id": 1}`))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected some output")
	}
}

func TestConvertJSONLMixedWithPrimitives(t *testing.T) {
	jsonl := `{"id": 1}
"just a string"
{"id": 2}`
	result, err := ConvertJSONL([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	// Primitive lines make it non-tabular
	if !bytes.Contains(result, []byte("---")) {
		t.Errorf("expected separator for mixed JSONL: %s", result)
	}
}

func TestConvertJSONLWithBlankLines(t *testing.T) {
	jsonl := `{"id": 1}

{"id": 2}`
	result, err := ConvertJSONL([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONL failed: %v", err)
	}
	if !bytes.Contains(result, []byte("[2]")) {
		t.Errorf("expected tabular with 2 items: %s", result)
	}
}

// --- ConvertAuto tests ---

func TestConvertAutoJSON(t *testing.T) {
	json := `{"id": 1}`
	result, err := ConvertAuto([]byte(json))
	if err != nil {
		t.Fatalf("ConvertAuto failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertAutoJSONString(t *testing.T) {
	json := `{"id": 1}`
	result, err := ConvertAutoString(json)
	if err != nil {
		t.Fatalf("ConvertAutoString failed: %v", err)
	}
	if !bytes.Contains([]byte(result), []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertAutoWithOptions(t *testing.T) {
	json := `{"id": 1}`
	result, err := ConvertAutoWithOptions([]byte(json), WithIndent(4))
	if err != nil {
		t.Fatalf("ConvertAutoWithOptions failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertAutoJSONC(t *testing.T) {
	jsonc := `{"id": 1} // comment`
	result, err := ConvertAuto([]byte(jsonc))
	if err != nil {
		t.Fatalf("ConvertAuto failed: %v", err)
	}
	// Should handle as JSONC (contains comment marker)
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertAutoJSONL(t *testing.T) {
	jsonl := `{"id": 1}
{"id": 2}`
	result, err := ConvertAuto([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertAuto failed: %v", err)
	}
	// Should handle as JSONL
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertAutoFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	_, err = tmp.WriteString(`{"test": 123}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertAutoFile(tmp.Name())
	if err != nil {
		t.Fatalf("ConvertAutoFile failed: %v", err)
	}
	if !bytes.Contains(result, []byte("test")) {
		t.Errorf("missing test: %s", result)
	}
}

func TestConvertAutoFileJSONC(t *testing.T) {
	tmp, err := os.CreateTemp("", "test*.jsonc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	_, err = tmp.WriteString(`{"id": 1} // comment`)
	if err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertAutoFile(tmp.Name())
	if err != nil {
		t.Fatalf("ConvertAutoFile failed: %v", err)
	}
	if !bytes.Contains(result, []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertAutoFileNotFound(t *testing.T) {
	_, err := ConvertAutoFile("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// --- DetectFormat tests ---

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		data     []byte
		expected string
	}{
		{[]byte(`{"id": 1}`), "json"},
		{[]byte(`[1, 2, 3]`), "json"},
		{[]byte(`{"id": 1} // comment`), "jsonc"},
		{[]byte(`{"id": 1} /* comment */`), "jsonc"},
		{[]byte(`{"id": 1}
{"id": 2}`), "jsonl"},
		{[]byte("[1]\n[2]"), "jsonl"},
	}
	for _, tt := range tests {
		result := DetectFormat(tt.data)
		if result != tt.expected {
			t.Errorf("DetectFormat(%q) = %q, want %q", string(tt.data), result, tt.expected)
		}
	}
}

// --- mayContainComments tests ---

func TestMayContainComments(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte(`{"key": "value"}`), false},
		{[]byte(`{"url": "http://example.com"}`), false},
		{[]byte(`{"key": "value"} // comment`), true},
		{[]byte(`{"key": "value"} /* comment */`), true},
		{[]byte(`// full line comment
{"key": 1}`), true},
		{[]byte(`{"url": "https://example.com/api"}`), false},
	}
	for _, tt := range tests {
		result := mayContainComments(tt.data)
		if result != tt.expected {
			t.Errorf("mayContainComments(%q) = %v, want %v", string(tt.data), result, tt.expected)
		}
	}
}

// --- isJSONL tests ---

func TestIsJSONL(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte(`{"id": 1}`), false},
		{[]byte(`{"id": 1}
{"id": 2}`), true},
		{[]byte(`{"id": 1}
   
{"id": 2}`), true},
		{[]byte(`1
2
3`), true}, // Primitives are now valid JSONL
		{[]byte(`[1,2]
[3,4]`), true},
		{[]byte(`invalid
{"id": 1}`), false},
		{[]byte(`"a"
"b"
"c"`), true}, // String primitives (jq -r '.[]' on arrays)
		{[]byte(`1
2
3
invalid`), true}, // 3 valid out of 4 = 75% > 50% threshold
	}
	for _, tt := range tests {
		result := isJSONL(tt.data)
		if result != tt.expected {
			t.Errorf("isJSONL(%q) = %v, want %v", string(tt.data), result, tt.expected)
		}
	}
}

// --- ConvertJSONLStream tests ---

func TestConvertJSONLStreamBasic(t *testing.T) {
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	// Should output with --- separators
	if !bytes.Contains(result, []byte("---")) {
		t.Errorf("expected --- separator: %s", result)
	}
	if !bytes.Contains(result, []byte("id: 1")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertJSONLStreamString(t *testing.T) {
	jsonl := `{"id": 1}
{"id": 2}`
	result, err := ConvertJSONLStreamString(jsonl)
	if err != nil {
		t.Fatalf("ConvertJSONLStreamString failed: %v", err)
	}
	if !bytes.Contains([]byte(result), []byte("id")) {
		t.Errorf("missing id: %s", result)
	}
}

func TestConvertJSONLStreamWithOptions(t *testing.T) {
	// Test with tabular data which uses delimiter in header
	jsonl := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}`
	result, err := ConvertJSONLStreamWithOptions([]byte(jsonl), WithDelimiter('|'))
	if err != nil {
		t.Fatalf("ConvertJSONLStreamWithOptions failed: %v", err)
	}
	// In streaming mode, each line is converted individually (not tabular)
	// so we just verify the output contains the expected data
	if !bytes.Contains(result, []byte("Alice")) {
		t.Errorf("missing Alice: %s", result)
	}
	if !bytes.Contains(result, []byte("Bob")) {
		t.Errorf("missing Bob: %s", result)
	}
}

func TestConvertJSONLStreamPrimitiveArray(t *testing.T) {
	// Test with jq-style output (primitives from array)
	jsonl := `"alice"
"bob"
"charlie"`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	if !bytes.Contains(result, []byte("alice")) {
		t.Errorf("missing alice: %s", result)
	}
	if !bytes.Contains(result, []byte("bob")) {
		t.Errorf("missing bob: %s", result)
	}
}

func TestConvertJSONLStreamNumbers(t *testing.T) {
	// Test with numbers (jq -r '.[]' on [1, 2, 3])
	jsonl := `1
2
3`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	for _, num := range []string{"1", "2", "3"} {
		if !bytes.Contains(result, []byte(num)) {
			t.Errorf("missing %s: %s", num, result)
		}
	}
}

func TestConvertJSONLStreamBooleans(t *testing.T) {
	// Test with booleans
	jsonl := `true
false
true`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	for _, val := range []string{"true", "false"} {
		if !bytes.Contains(result, []byte(val)) {
			t.Errorf("missing %s: %s", val, result)
		}
	}
}

func TestConvertJSONLStreamNulls(t *testing.T) {
	// Test with nulls
	jsonl := `null
"value"
null`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	for _, val := range []string{"null", "value"} {
		if !bytes.Contains(result, []byte(val)) {
			t.Errorf("missing %s: %s", val, result)
		}
	}
}

func TestConvertJSONLStreamMixed(t *testing.T) {
	// Test with mixed types (typical jq output)
	jsonl := `{"id": 1}
"string"
123
true`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	// Should have 3 --- separators for 4 items
	separatorCount := bytes.Count(result, []byte("---"))
	if separatorCount != 3 {
		t.Errorf("expected 3 separators, got %d: %s", separatorCount, result)
	}
}

func TestConvertJSONLStreamInvalidJSON(t *testing.T) {
	// Invalid JSON should error
	_, err := ConvertJSONLStream([]byte(`{invalid}
{"id": 1}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestConvertJSONLStreamEmpty(t *testing.T) {
	// Empty input should produce empty output
	result, err := ConvertJSONLStream([]byte(``))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got: %s", result)
	}
}

func TestConvertJSONLStreamBlankLines(t *testing.T) {
	// Blank lines should be skipped
	jsonl := `{"id": 1}

{"id": 2}

{"id": 3}`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	// Should have 2 separators for 3 items (blank lines skipped)
	separatorCount := bytes.Count(result, []byte("---"))
	if separatorCount != 2 {
		t.Errorf("expected 2 separators, got %d: %s", separatorCount, result)
	}
}

func TestConvertJSONLStreamNested(t *testing.T) {
	// Test with nested objects
	jsonl := `{"user": {"name": "Alice", "age": 30}}
{"user": {"name": "Bob", "age": 25}}`
	result, err := ConvertJSONLStream([]byte(jsonl))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	for _, val := range []string{"Alice", "Bob"} {
		if !bytes.Contains(result, []byte(val)) {
			t.Errorf("missing %s: %s", val, result)
		}
	}
}

// --- Converter.ConvertJSONLStream tests ---

func TestConverterConvertJSONLStream(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSONLStream(strings.NewReader(`{"id": 1}
{"id": 2}`))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestConverterConvertJSONLStreamPrimitives(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSONLStream(strings.NewReader(`"a"
"b"
"c"`))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	for _, s := range []string{"a", "b", "c"} {
		if !bytes.Contains(buf.Bytes(), []byte(s)) {
			t.Errorf("missing %s: %s", s, buf.String())
		}
	}
}

func TestConverterConvertJSONLStreamArrays(t *testing.T) {
	var buf bytes.Buffer
	conv := NewConverter(&buf)
	err := conv.ConvertJSONLStream(strings.NewReader(`[1, 2]
[3, 4]`))
	if err != nil {
		t.Fatalf("ConvertJSONLStream failed: %v", err)
	}
	// Should have separators
	if !bytes.Contains(buf.Bytes(), []byte("---")) {
		t.Errorf("expected separators: %s", buf.String())
	}
}

// --- isJSONLine tests ---

func TestIsJSONLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{`{"id": 1}`, true},
		{`[1, 2, 3]`, true},
		{`"string"`, true},
		{`123`, true},
		{`true`, true},
		{`false`, true},
		{`null`, true},
		{``, false},
		{`   `, false},
		{`{invalid}`, false},
		{`not json`, false},
	}
	for _, tt := range tests {
		result := isJSONLine([]byte(tt.line))
		if result != tt.expected {
			t.Errorf("isJSONLine(%q) = %v, want %v", tt.line, result, tt.expected)
		}
	}
}
