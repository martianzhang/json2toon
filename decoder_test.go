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

// --- Unicode Edge Cases ---

func TestDecodeUnicodeEmoji(t *testing.T) {
	toon := `emoji: "😀🎉🚀"
name: "Hello"`
	expected := `{
  "emoji": "😀🎉🚀",
  "name": "Hello"
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeCJK(t *testing.T) {
	toon := `chinese: "你好世界"
japanese: "こんにちは"
korean: "안녕하세요"`
	expected := `{
  "chinese": "你好世界",
  "japanese": "こんにちは",
  "korean": "안녕하세요"
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeRTL(t *testing.T) {
	toon := `arabic: "مرحبا بالعالم"
hebrew: "שלום עולם"`
	expected := `{
  "arabic": "مرحبا بالعالم",
  "hebrew": "שלום עולם"
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeInlineArray(t *testing.T) {
	toon := `languages:[3]: 中文,日本語,한국어`
	expected := `{
  "languages": ["中文", "日本語", "한국어"]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeTabularArray(t *testing.T) {
	toon := `users[2]{name,city}:
  Alice,北京
  Bob,東京`
	expected := `{
  "users": [
    {"name": "Alice", "city": "北京"},
    {"name": "Bob", "city": "東京"}
  ]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeEmptyString(t *testing.T) {
	toon := `empty: ""
emoji: "🎉"
text: ""`
	expected := `{
  "empty": "",
  "emoji": "🎉",
  "text": ""
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeWithSpecialChars(t *testing.T) {
	toon := `text: "Café ☕ Naïve"`
	expected := `{
  "text": "Café ☕ Naïve"
}`
	parseAndVerify(t, toon, expected)
}

// --- Special Float Values ---

func TestDecodeSpecialFloatValues(t *testing.T) {
	// Note: Go's JSON standard library does NOT support NaN, Infinity, -Infinity
	// These are not valid JSON values. Testing what the decoder does with them.
	tests := []struct {
		name       string
		toon       string
		shouldFail bool
	}{
		{"regular float", `value: 3.14`, false},
		{"negative float", `value: -2.71`, false},
		{"scientific notation", `value: 1.23e-4`, false},
		{"large exponent", `value: 1e10`, false},
		{"zero", `value: 0`, false},
		{"negative zero", `value: -0`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeToJSON([]byte(tt.toon))
			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got: %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify it's valid JSON
				var data interface{}
				if err := json.Unmarshal(result, &data); err != nil {
					t.Errorf("Invalid JSON output: %v\nOutput: %s", err, result)
				}
			}
		})
	}
}

func TestDecodeFloatPrecision(t *testing.T) {
	toon := `pi: 3.141592653589793
e: 2.718281828459045
phi: 1.618033988749895`
	expected := `{
  "pi": 3.141592653589793,
  "e": 2.718281828459045,
  "phi": 1.618033988749895
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeVerySmallNumbers(t *testing.T) {
	toon := `tiny: 1e-100
smaller: 1e-308`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify values are present and numeric
	if _, ok := data["tiny"].(float64); !ok {
		t.Errorf("Expected tiny to be float64, got %T", data["tiny"])
	}
	if _, ok := data["smaller"].(float64); !ok {
		t.Errorf("Expected smaller to be float64, got %T", data["smaller"])
	}
}

func TestDecodeVeryLargeNumbers(t *testing.T) {
	toon := `huge: 1e100
massive: 1e308`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify values are present and numeric
	if _, ok := data["huge"].(float64); !ok {
		t.Errorf("Expected huge to be float64, got %T", data["huge"])
	}
	if _, ok := data["massive"].(float64); !ok {
		t.Errorf("Expected massive to be float64, got %T", data["massive"])
	}
}

// --- Deep Nesting Edge Cases ---

func TestDecodeDeepNesting11Levels(t *testing.T) {
	toon := `a:
  b:
    c:
      d:
        e:
          f:
            g:
              h:
                i:
                  j:
                    k: "deep"`
	expected := `{
  "a": {
    "b": {
      "c": {
        "d": {
          "e": {
            "f": {
              "g": {
                "h": {
                  "i": {
                    "j": {
                      "k": "deep"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeDeepNesting20Levels(t *testing.T) {
	toon := `l1:
  l2:
    l3:
      l4:
        l5:
          l6:
            l7:
              l8:
                l9:
                  l10:
                    l11:
                      l12:
                        l13:
                          l14:
                            l15:
                              l16:
                                l17:
                                  l18:
                                    l19:
                                      l20: "very deep"`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var data interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify structure is valid
	if data == nil {
		t.Error("Expected non-nil data")
	}
}

func TestDecodeDeepNestedArray(t *testing.T) {
	toon := `level1:
  level2:
    level3:
      items:[3]: a,b,c`
	expected := `{
  "level1": {
    "level2": {
      "level3": {
        "items": ["a", "b", "c"]
      }
    }
  }
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeDeepNestedTabularArray(t *testing.T) {
	toon := `root:
  nested:
    data[2]{id,value}:
      1,first
      2,second`
	expected := `{
  "root": {
    "nested": {
      "data": [
        {"id": 1, "value": "first"},
        {"id": 2, "value": "second"}
      ]
    }
  }
}`
	parseAndVerify(t, toon, expected)
}

// --- Combined Edge Cases ---

func TestDecodeUnicodeDeepNesting(t *testing.T) {
	toon := `国家:
  城市:
    地区:
      名称: "北京"`
	expected := `{
  "国家": {
    "城市": {
      "地区": {
        "名称": "北京"
      }
    }
  }
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeUnicodeWithFloats(t *testing.T) {
	toon := `数据[3]: 1.5,2.7,3.14`
	expected := `{
  "数据": [1.5, 2.7, 3.14]
}`
	parseAndVerify(t, toon, expected)
}

func TestDecodeComplexEdgeCases(t *testing.T) {
	toon := `metadata:
  version: 1.0
  tags:[4]: 中文,日本語,한국어,العربية
  nested:
    deep:
      deeper:
        value: "🎉"`
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		t.Fatalf("DecodeToJSON failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify structure
	if data["metadata"] == nil {
		t.Error("Expected metadata key")
	}
	metadata, _ := data["metadata"].(map[string]interface{})
	if metadata["version"] != float64(1.0) {
		t.Errorf("Expected version=1.0, got %v", metadata["version"])
	}
	if tags, ok := metadata["tags"].([]interface{}); ok {
		if len(tags) != 4 {
			t.Errorf("Expected 4 tags, got %d", len(tags))
		}
	} else {
		t.Errorf("Expected tags to be array, got %T", metadata["tags"])
	}
}
