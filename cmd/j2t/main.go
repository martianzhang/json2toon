// j2t converts JSON/JSONC/JSONL to TOON format.
//
// Usage:
//
//	j2t [options] [file]
//
// If no file is specified, reads from stdin.
// Automatically detects format based on content:
//   - JSONC: contains // or /* comments
//   - JSONL: multiple complete JSON objects on separate lines
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/martianzhang/json2toon"
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
// 1. Multiple lines exist
// 2. Most non-empty lines are complete JSON objects/arrays (start and end properly)
// 3. Each line parses as valid JSON when trimmed
func isJSONL(data []byte) bool {
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) < 2 {
		return false
	}

	validLines := 0
	totalLines := 0

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		totalLines++

		// Check if line looks like a complete JSON object or array
		if (trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}') ||
			(trimmed[0] == '[' && trimmed[len(trimmed)-1] == ']') {
			// Verify it's valid JSON
			if json.Valid(trimmed) {
				validLines++
			}
		}
	}

	// If we have 2+ lines and most are valid complete JSON objects, it's JSONL
	return totalLines >= 2 && validLines >= 2 && float64(validLines)/float64(totalLines) > 0.5
}

func main() {
	var (
		indent     int
		delimiter  string
		keyFolding string
		strict     bool
	)

	flag.IntVar(&indent, "indent", 2, "indentation spaces")
	flag.StringVar(&delimiter, "delimiter", ",", "table delimiter character")
	flag.StringVar(&keyFolding, "key-folding", "", "key folding mode (empty, \"safe\", \"upper\", \"lower\")")
	flag.BoolVar(&strict, "strict", false, "strict mode")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `j2t - JSON to TOON converter

Usage:
  j2t [options] [file]

Options:
`)
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `
Examples:
  echo '{"id": 1}' | j2t
  j2t config.json
  j2t config.jsonc
  j2t data.jsonl
`)
	}
	flag.Parse()

	var input io.Reader

	args := flag.Args()
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", args[0], err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		input = f
	} else {
		input = os.Stdin
	}

	data, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	opts := []json2toon.ConverterOption{
		json2toon.WithIndent(indent),
	}
	if delimiter != "," {
		if len(delimiter) != 1 {
			fmt.Fprintln(os.Stderr, "Error: delimiter must be a single character")
			os.Exit(1)
		}
		opts = append(opts, json2toon.WithDelimiter(rune(delimiter[0])))
	}
	if keyFolding != "" {
		opts = append(opts, json2toon.WithKeyFolding(keyFolding))
	}
	if strict {
		opts = append(opts, json2toon.WithStrict(strict))
	}

	// Detect format and convert
	if mayContainComments(data) {
		convertJSONC(data, opts)
	} else if isJSONL(data) {
		convertJSONL(data, opts)
	} else {
		convertJSON(data, opts)
	}
}

func convertJSON(data []byte, opts []json2toon.ConverterOption) {
	result, err := json2toon.ConvertWithOptions(data, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(result)
}

func convertJSONC(data []byte, opts []json2toon.ConverterOption) {
	result, err := json2toon.ConvertJSONCWithOptions(data, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(result)
}

func convertJSONL(data []byte, opts []json2toon.ConverterOption) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Parse all lines first
	type lineData struct {
		raw    []byte
		keys   []string
		values map[string]interface{}
	}
	var lines []lineData
	isTabular := true
	var expectedKeys []string

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		// Parse to get keys
		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			// Not a JSON object, fall back to line-by-line conversion
			isTabular = false
		} else {
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}

			if expectedKeys == nil {
				expectedKeys = keys
			} else if !sameKeys(expectedKeys, keys) {
				isTabular = false
			}

			lines = append(lines, lineData{
				raw:    line,
				keys:   keys,
				values: obj,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if isTabular && len(lines) > 0 {
		// Output as tabular array
		header := fmt.Sprintf("[%d]{%s}:\n", len(lines), strings.Join(expectedKeys, ","))
		_, _ = os.Stdout.Write([]byte(header))

		for i, line := range lines {
			indent := "  "
			var vals []string
			for _, k := range expectedKeys {
				vals = append(vals, formatValue(line.values[k], ","))
			}
			_, _ = os.Stdout.Write([]byte(indent + strings.Join(vals, ",")))
			if i < len(lines)-1 {
				_, _ = os.Stdout.Write([]byte("\n"))
			}
		}
	} else {
		// Output each line separately with separator
		first := true
		for _, line := range lines {
			result, err := json2toon.ConvertWithOptions(line.raw, opts...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error converting line: %v\n", err)
				os.Exit(1)
			}

			if !first {
				_, _ = os.Stdout.Write([]byte("---\n"))
			}
			_, _ = os.Stdout.Write(bytes.TrimSpace(result))
			first = false
		}
	}
}

// sameKeys checks if two key slices contain the same elements.
func sameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	slices.Sort(a)
	slices.Sort(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// formatValue converts a value to its string representation.
func formatValue(v interface{}, delim string) string {
	switch val := v.(type) {
	case nil:
		return ""
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%.0f", val)
	case string:
		if strings.ContainsAny(val, delim+`"`) || strings.ContainsAny(val, " \n:") {
			return `"` + val + `"`
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
