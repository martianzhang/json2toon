// j2t converts JSON/JSONC/JSONL to TOON format (and vice versa with -r or --decode).
//
// Usage:
//
//	j2t [options] [file]
//	j2t -o output.toon input.json
//
// Automatically detects format based on file extension:
//   - .toon files → decode (TOON to JSON)
//   - .json, .jsonc, .jsonl files → encode (JSON to TOON)
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/martianzhang/json2toon"
)

func main() {
	var (
		indent      int
		delimiter   string
		keyFolding  string
		strict      bool
		stream      bool
		reverse     bool
		forceEncode bool
		forceDecode bool
		outputFile  string
		stats       bool
		expandPaths string
	)

	flag.IntVar(&indent, "indent", 2, "indentation spaces")
	flag.StringVar(&delimiter, "delimiter", ",", "table delimiter character (, \\t |)")
	flag.StringVar(&keyFolding, "key-folding", "", "key folding mode (off, safe)")
	flag.BoolVar(&strict, "strict", false, "enable strict validation")
	flag.BoolVar(&stream, "stream", false, "enable streaming mode for JSONL (line-by-line processing)")
	flag.BoolVar(&reverse, "r", false, "reverse: convert TOON to JSON (alias for --decode)")
	flag.BoolVar(&forceEncode, "e", false, "force encode mode (JSON to TOON)")
	flag.BoolVar(&forceDecode, "d", false, "force decode mode (TOON to JSON)")
	flag.BoolVar(&stats, "stats", false, "show token count estimates and savings (encode only)")
	flag.StringVar(&expandPaths, "expandPaths", "", "path expansion mode (off, safe)")
	flag.StringVar(&outputFile, "o", "", "output file path (prints to stdout if omitted)")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `j2t - JSON/TOON converter

Usage:
  j2t [options] [file]          # Auto-detect based on file extension
  j2t -o output.toon input.json  # Encode JSON to TOON
  j2t -o output.json input.toon  # Decode TOON to JSON
  j2t -r [file]                 # Force decode TOON to JSON

Options:
`)
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `
Examples:
  # Encode JSON to TOON
  echo '{"id": 1}' | j2t
  j2t config.json -o config.toon

  # Decode TOON to JSON
  echo "id: 1" | j2t -r
  j2t config.toon -o config.json

  # Token statistics
  j2t data.json --stats

  # Key folding for smaller output
  j2t config.json --key-folding safe -o config.toon

  # Path expansion for lossless round-trip
  j2t compressed.toon --expandPaths safe -o output.json
`)
	}
	flag.Parse()

	var input io.Reader
	var data []byte
	var err error

	// Determine mode based on flags and file extension
	args := flag.Args()
	isDecode := forceDecode || reverse

	if len(args) > 0 && !forceEncode && !forceDecode && !reverse {
		// Check file extension for auto-detection
		ext := strings.ToLower(filepath.Ext(args[0]))
		switch ext {
		case ".toon":
			isDecode = true
		case ".json", ".jsonc", ".jsonl":
			isDecode = false
		}
	}

	// Open input file or read from stdin
	if len(args) > 0 {
		f, openErr := os.Open(args[0])
		if openErr != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", args[0], openErr)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		input = f
	} else {
		input = os.Stdin
	}

	data, err = io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var result []byte
	var output io.Writer = os.Stdout

	// Open output file if specified
	if outputFile != "" {
		f, openErr := os.Create(outputFile)
		if openErr != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", outputFile, openErr)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		output = f
	}

	if isDecode {
		// Decode mode: TOON to JSON
		opts := []json2toon.DecodeOption{}
		if strict {
			opts = append(opts, json2toon.WithDecodeStrict(strict))
		}
		if expandPaths == "safe" {
			opts = append(opts, json2toon.WithExpandPaths(true))
		}

		result, err = json2toon.ConvertToJSONWithOptions(data, opts...)
	} else {
		// Encode mode: JSON/JSONC/JSONL to TOON
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

		if stream {
			result, err = json2toon.ConvertJSONLStreamWithOptions(data, opts...)
		} else {
			result, err = json2toon.ConvertAutoWithOptions(data, opts...)
		}

		// Show stats if requested
		if stats && err == nil {
			jsonTokens := countJSONTokens(data)
			toonTokens := countTOONTokens(result)
			saved := jsonTokens - toonTokens
			pct := 0.0
			if jsonTokens > 0 {
				pct = float64(saved) / float64(jsonTokens) * 100
			}
			fmt.Fprintf(os.Stderr, "Token estimates: ~%d (JSON) -> ~%d (TOON)\n", jsonTokens, toonTokens)
			if saved > 0 {
				fmt.Fprintf(os.Stderr, "Saved ~%d tokens (-%.1f%%)\n", saved, pct)
			} else if saved < 0 {
				fmt.Fprintf(os.Stderr, "Increased by ~%d tokens (+%.1f%%)\n", -saved, -pct)
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}

	_, _ = output.Write(result)
}

// countJSONTokens estimates token count for JSON input
// Uses json.Decoder to count actual tokens for accuracy
func countJSONTokens(data []byte) int {
	decoder := json.NewDecoder(bytes.NewReader(data))
	count := 0
	for {
		_, err := decoder.Token()
		if err != nil {
			break
		}
		count++
	}
	// Adjust for Delim tokens (brackets, braces) which appear as tokens
	// but we want to count semantic tokens (keys, values, array elements)
	// Rough estimate: divide by ~1.5 to get closer to semantic tokens
	return count * 2 / 3
}

// countTOONTokens estimates token count for TOON output.
// It uses a state machine to correctly count:
// - Inline arrays: items:[3]: 1,2,3 (1 key + 3 values = 4 tokens)
// - Tabular arrays: [N]{key1,key2}: header + N rows of comma-separated values
// - Regular key:value lines as 1 token each
func countTOONTokens(data []byte) int {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	count := 0
	inTabular := false // true when inside a tabular array

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Tabular array header: [N]{key1,key2}:
		// Matches lines like "[3]{active,id,name}:" or "items:[3]{key1,key2}:"
		if strings.Contains(line, "[") && strings.Contains(line, "]{") && strings.HasSuffix(line, ":") {
			// Count the header line as 1 token
			count++
			inTabular = true
			continue
		}

		// Inline array on single line: items:[3]: 1,2,3
		if strings.Contains(line, ":[") && strings.Contains(line, "]:") {
			count++ // The key
			afterBracket := strings.Index(line, "]:")
			if afterBracket > 0 {
				values := strings.TrimSpace(line[afterBracket+2:])
				if values != "" {
					count += strings.Count(values, ",") + 1
				}
			}
			continue
		}

		// Tabular array row: comma-separated values (when inside tabular array)
		if inTabular && strings.Contains(line, ",") {
			count += strings.Count(line, ",") + 1
			continue
		}

		// End of tabular array (indent resets, no longer comma-separated)
		if inTabular {
			inTabular = false
		}

		// Regular line: key: value or list item (- prefix)
		count++
	}

	// Tab-delimited arrays are more token-efficient
	tabCount := bytes.Count(data, []byte("\t"))
	count += tabCount / 2

	return count
}
