package json2toon

import (
	"encoding/json"
	"os"
	"testing"
)

// Integration tests use actual files and test end-to-end functionality

func TestIntegrationConvertFile(t *testing.T) {
	result, err := ConvertFile("testdata/example.json")
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertFile returned empty result")
	}
}

func TestIntegrationConvertFileNotFound(t *testing.T) {
	_, err := ConvertFile("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestIntegrationFileToWriter(t *testing.T) {
	err := ConvertFileToWriter("testdata/example.json", &nopWriter{})
	if err != nil {
		t.Fatalf("ConvertFileToWriter failed: %v", err)
	}
}

func TestIntegrationFileToWriterNotFound(t *testing.T) {
	err := ConvertFileToWriter("nonexistent.json", &nopWriter{})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestIntegrationToJSONFile(t *testing.T) {
	result, err := ConvertToJSONFile("testdata/example.toon")
	if err != nil {
		t.Fatalf("ConvertToJSONFile failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertToJSONFile returned empty result")
	}
}

func TestIntegrationToJSONFileString(t *testing.T) {
	result, err := ConvertToJSONFileString("testdata/example.toon")
	if err != nil {
		t.Fatalf("ConvertToJSONFileString failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertToJSONFileString returned empty result")
	}
}

func TestIntegrationToJSONFileNotFound(t *testing.T) {
	_, err := ConvertToJSONFile("nonexistent.toon")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// nopWriter implements io.Writer for testing
type nopWriter struct{}

func (w *nopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestIntegrationRoundTrip(t *testing.T) {
	// Test that both directions produce valid JSON/TOON output
	jsonData, err := os.ReadFile("testdata/example.json")
	if err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	toon, err := Convert(jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	backToJSON, err := ConvertToJSON(toon)
	if err != nil {
		t.Fatalf("ConvertToJSON failed: %v", err)
	}

	// Verify round-tripped output is valid JSON
	var roundTripped interface{}
	if err := json.Unmarshal(backToJSON, &roundTripped); err != nil {
		t.Fatalf("Round-trip produced invalid JSON: %v\nOutput: %s", err, backToJSON)
	}
}

func TestIntegrationLargeFile(t *testing.T) {
	result, err := ConvertFile("testdata/example.json")
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertFile returned empty result")
	}
}
