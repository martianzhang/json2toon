package json2toon

import (
	"bytes"
	"encoding/json"
)

// mayContainComments checks if data may contain JSONC comments (// or /*).
func mayContainComments(data []byte) bool {
	inString := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' && (i == 0 || data[i-1] != '\\') {
			inString = !inString
			continue
		}
		if !inString && i+1 < len(data) {
			if data[i] == '/' && (data[i+1] == '/' || data[i+1] == '*') {
				return true
			}
		}
	}
	return false
}

// isJSONL checks if data is JSON Lines format.
// Returns true if:
//  1. Multiple lines exist
//  2. Most non-empty lines are valid JSON values (objects, arrays, or primitives)
func isJSONL(data []byte) bool {
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) < 2 {
		return false
	}

	validLines := 0
	totalLines := 0

	for _, line := range lines {
		if isJSONLine(line) {
			validLines++
			totalLines++
		} else {
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				totalLines++
			}
		}
	}

	// If we have 2+ lines and most are valid JSON, it's JSONL
	return totalLines >= 2 && validLines >= 2 && float64(validLines)/float64(totalLines) > 0.5
}

// isJSONLine checks if a single line looks like a valid JSON value (object, array, or primitive).
func isJSONLine(line []byte) bool {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}
	// Check for complete JSON values
	if json.Valid(trimmed) {
		return true
	}
	return false
}

// DetectFormat detects the input format: JSON, JSONC, or JSONL.
func DetectFormat(data []byte) string {
	if mayContainComments(data) {
		return "jsonc"
	}
	if isJSONL(data) {
		return "jsonl"
	}
	return "json"
}
