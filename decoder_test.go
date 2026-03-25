package json2toon

import (
	"encoding/json"
	"strings"
	"testing"
)

// parseAndVerify parses TOON and verifies the JSON output
func parseAndVerify(t *testing.T, toon string, expectedJSON string) {
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, result)
	}
	if err := json.Unmarshal([]byte(expectedJSON), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expectedJSON)
	}
}

func jsonEq(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// --- Basic Tests ---

func TestDecodeSimpleObject(t *testing.T) {
	toon := `id: 123
name: Ada
active: true`
	expected := `{
  "id": 123,
  "name": "Ada",
  "active": true
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeNestedObject(t *testing.T) {
	toon := `user:
  id: 1
  name: Alice`
	expected := `{
  "user": {
    "id": 1,
    "name": "Alice"
  }
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodePrimitives(t *testing.T) {
	tests := []struct {
		toon     string
		expected string
	}{
		{`"hello"`, `"hello"`},
		{`123`, `123`},
		{`-3.14`, `-3.14`},
		{`true`, `true`},
		{`false`, `false`},
		{`null`, `null`},
	}

	for _, tt := range tests {
		result, err := DecodeToJSON([]byte(tt.toon))
		if err != nil {
			t.Errorf("Failed to decode %q: %v", tt.toon, err)
			continue
		}

		// Parse both as interface{} for comparison
		var got, want interface{}
		if err := json.Unmarshal(result, &got); err != nil {
			t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, result)
		}
		if err := json.Unmarshal([]byte(tt.expected), &want); err != nil {
			t.Fatalf("Invalid expected JSON: %v", err)
		}

		if !jsonEq(got, want) {
			t.Errorf("Mismatch for %q:\nGot: %s\nWant: %s", tt.toon, result, tt.expected)
		}
	}
}

// --- Array Tests ---

func TestDecodeInlineArray(t *testing.T) {
	toon := `tags:[3]: admin,ops,dev`
	expected := `{
  "tags": ["admin", "ops", "dev"]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeInlineArrayWithNumbers(t *testing.T) {
	toon := `nums:[3]: 1,2,3`
	expected := `{
  "nums": [1, 2, 3]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeTabularArray(t *testing.T) {
	toon := `users[2]{id,name}:
  1,Alice
  2,Bob`
	expected := `{
  "users": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeListArray(t *testing.T) {
	toon := `items:
  - id: 1
  - id: 2
  - id: 3`
	expected := `{
  "items": [1, 2, 3]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeMixedListArray(t *testing.T) {
	toon := `items:
  - id: 1
  - name: Alice
  - id: 3`
	expected := `{
  "items": [{"id": 1}, {"name": "Alice"}, {"id": 3}]
}`
	parseAndVerify(t, toon, expected)
}

// --- Edge Cases ---

func TestDecodeEmptyObject(t *testing.T) {
	toon := `data: {}`
	expected := `{
  "data": {}
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeEmptyArray(t *testing.T) {
	toon := `items:`
	expected := `{
  "items": []
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeQuotedStrings(t *testing.T) {
	toon := `msg: "hello world"
empty: ""
with spaces: "  leading and trailing  "`
	expected := `{
  "msg": "hello world",
  "empty": "",
  "with spaces": "  leading and trailing  "
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeEscapedStrings(t *testing.T) {
	toon := `text: "line1\nline2"
tabbed: "col1\tcol2"`
	expected := `{
  "text": "line1\nline2",
  "tabbed": "col1\tcol2"
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeWithComments(t *testing.T) {
	toon := `id: 1  # this is a comment
name: Alice`
	expected := `{
  "id": 1,
  "name": "Alice"
}`
	parseAndVerify(t, toon, expected)
}

// --- Error Handling ---

func TestDecodeInvalidInlineArray(t *testing.T) {
	_, err := DecodeToJSON([]byte(`key:[2]: 1,2,3,4`))
	if err == nil {
		t.Error("Expected error for mismatched array count")
	}
}

// --- API Tests ---

func TestConvertToJSONBytes(t *testing.T) {
	toon := `id: 1
name: Alice`
	result, err := ConvertToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("ConvertToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var data interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Not valid JSON: %v\nOutput: %s", err, result)
	}
}

func TestConvertToJSONString(t *testing.T) {
	toon := `id: 1`
	result, err := ConvertToJSONString(toon)
	if err != nil {
		t.Fatalf("ConvertToJSONString failed: %v", err)
	}

	// Verify it contains JSON
	if !strings.Contains(result, "\"id\"") {
		t.Errorf("Expected JSON output, got: %s", result)
	}
}

func TestConvertToJSONFromReader(t *testing.T) {
	toon := `id: 1
name: Bob`
	result, err := ConvertToJSONFromReader(strings.NewReader(toon))
	if err != nil {
		t.Fatalf("ConvertToJSONFromReader failed: %v", err)
	}

	var data interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Not valid JSON: %v", err)
	}
}

// --- Round-trip Tests ---

func TestRoundTripSimple(t *testing.T) {
	original := `{"id": 1, "name": "Alice", "active": true}`

	// Convert to TOON
	toon, err := Convert([]byte(original))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Convert back to JSON
	result, err := DecodeToJSON(toon)
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	// Parse both and compare
	var origData, resultData interface{}
	if err := json.Unmarshal([]byte(original), &origData); err != nil {
		t.Fatalf("Invalid original JSON: %v", err)
	}
	if err := json.Unmarshal(result, &resultData); err != nil {
		t.Fatalf("Invalid result JSON: %v", err)
	}

	if !jsonEq(origData, resultData) {
		t.Errorf("Round-trip failed.\nOriginal: %s\nTOON: %s\nResult: %s", original, toon, result)
	}
}

func TestRoundTripNested(t *testing.T) {
	original := `{"user": {"profile": {"name": "Alice"}}}`

	toon, _ := Convert([]byte(original))
	result, _ := DecodeToJSON(toon)

	var origData, resultData interface{}
	if err := json.Unmarshal([]byte(original), &origData); err != nil {
		t.Fatalf("Invalid original JSON: %v", err)
	}
	if err := json.Unmarshal(result, &resultData); err != nil {
		t.Fatalf("Invalid result JSON: %v", err)
	}

	if !jsonEq(origData, resultData) {
		t.Errorf("Round-trip failed for nested object")
	}
}

func TestRoundTripArray(t *testing.T) {
	original := `{"items": [1, 2, 3]}`

	toon, _ := Convert([]byte(original))
	result, _ := DecodeToJSON(toon)

	var origData, resultData interface{}
	if err := json.Unmarshal([]byte(original), &origData); err != nil {
		t.Fatalf("Invalid original JSON: %v", err)
	}
	if err := json.Unmarshal(result, &resultData); err != nil {
		t.Fatalf("Invalid result JSON: %v", err)
	}

	if !jsonEq(origData, resultData) {
		t.Errorf("Round-trip failed for array")
	}
}

// --- Complex Structures ---

func TestDecodeComplexObject(t *testing.T) {
	toon := `id: 12345
name: Test Project
version: "1.0.0"
active: true
config:
  debug: false
  logLevel: info
  maxRetries: 3
  timeout: 30.5
tags:[3]: json,toon,converter`

	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Not valid JSON: %v", err)
	}

	// Verify structure
	if data["id"] != float64(12345) {
		t.Errorf("Expected id=12345, got %v", data["id"])
	}
	if data["name"] != "Test Project" {
		t.Errorf("Expected name=Test Project, got %v", data["name"])
	}
	if data["active"] != true {
		t.Errorf("Expected active=true, got %v", data["active"])
	}
	if tags, ok := data["tags"].([]interface{}); ok {
		if len(tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(tags))
		}
	} else {
		t.Errorf("Expected tags to be array, got %T", data["tags"])
	}
}

// --- Expand Paths Tests ---

func TestDecodeExpandPaths(t *testing.T) {
	toon := `data.metadata.items[2]: a,b`
	expected := `{
  "data": {
    "metadata": {
      "items": ["a", "b"]
    }
  }
}`
	result, err := ConvertToJSONWithOptions([]byte(toon), WithExpandPaths(true))
	if err != nil {
		t.Fatalf("ConvertToJSONWithOptions failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

func TestDecodeExpandPathsMultiple(t *testing.T) {
	toon := `a.b.c: value1
a.b.d: value2`
	expected := `{
  "a": {
    "b": {
      "c": "value1",
      "d": "value2"
    }
  }
}`
	result, err := ConvertToJSONWithOptions([]byte(toon), WithExpandPaths(true))
	if err != nil {
		t.Fatalf("ConvertToJSONWithOptions failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

// --- Root Array Tests ---

func TestDecodeRootArray(t *testing.T) {
	toon := `[3]: x,y,z`
	expected := `["x", "y", "z"]`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

func TestDecodeRootPrimitive(t *testing.T) {
	toon := `"hello"`
	expected := `"hello"`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

// --- Empty Container Tests ---

func TestDecodeEmptyArrayInline(t *testing.T) {
	toon := `items:[0]:`
	expected := `{"items": []}`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

// --- Tabular Header Format Tests ---

func TestDecodeTabularWithBrackets(t *testing.T) {
	toon := `users[2]{id,name}:
  1,Alice
  2,Bob`
	expected := `{
  "users": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ]
}`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

// --- Inline Array Count Validation ---

func TestDecodeInlineArrayCountMismatch(t *testing.T) {
	_, err := DecodeToJSON([]byte(`key:[2]: 1,2,3,4`))
	if err == nil {
		t.Error("Expected error for count mismatch")
	}
}

func TestDecodeInlineArrayCountMatch(t *testing.T) {
	toon := `key:[2]: 1,2`
	expected := `{"key": [1, 2]}`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

// --- Alternative Format Tests ---

func TestDecodeInlineArrayWithSpaceBeforeBracket(t *testing.T) {
	toon := `key: [3]: 1,2,3`
	expected := `{"key": [1, 2, 3]}`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

func TestDecodeKeyWithoutArray(t *testing.T) {
	toon := `key: value`
	expected := `{"key": "value"}`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

func TestDecodeStandaloneListItems(t *testing.T) {
	// Test standalone list items (parseListItems)
	toon := `- item1
- item2
- item3`
	expected := `["item1","item2","item3"]`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}

	if !jsonEq(got, want) {
		t.Errorf("JSON mismatch.\nGot: %s\nWant: %s", result, expected)
	}
}

func TestDecodeStandaloneListItemsWithObjects(t *testing.T) {
	// Test list items with key:value format (parseListItems)
	// Note: When all items have the same key and no nested content,
	// the decoder returns a simple array of values
	toon := `- name: Alice
- name: Bob
- name: Carol`
	// The decoder returns values when parseListArrayItems detects same key
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var got interface{}
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Should return a valid array
	arr, ok := got.([]interface{})
	if !ok {
		t.Fatalf("Expected array, got: %T", got)
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 items, got %d", len(arr))
	}
}
