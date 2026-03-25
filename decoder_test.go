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
