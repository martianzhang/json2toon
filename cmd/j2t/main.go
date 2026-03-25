// j2t converts JSON/JSONC/JSONL to TOON format (and vice versa with -r).
//
// Usage:
//
//	j2t [options] [file]
//	j2t -r [options] [file]
//
// If no file is specified, reads from stdin.
// Automatically detects format based on content:
//   - JSONC: contains // or /* comments
//   - JSONL: multiple complete JSON objects on separate lines
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/martianzhang/json2toon"
)

func main() {
	var (
		indent     int
		delimiter  string
		keyFolding string
		strict     bool
		stream     bool
		reverse    bool
	)

	flag.IntVar(&indent, "indent", 2, "indentation spaces")
	flag.StringVar(&delimiter, "delimiter", ",", "table delimiter character")
	flag.StringVar(&keyFolding, "key-folding", "", "key folding mode (empty, \"safe\", \"upper\", \"lower\")")
	flag.BoolVar(&strict, "strict", false, "strict mode")
	flag.BoolVar(&stream, "stream", false, "enable streaming mode for JSONL (line-by-line processing)")
	flag.BoolVar(&reverse, "r", false, "reverse: convert TOON to JSON")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `j2t - JSON/TOON converter

Usage:
  j2t [options] [file]        # JSON/JSONC/JSONL -> TOON
  j2t -r [options] [file]      # TOON -> JSON

Options:
`)
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `
Examples:
  # JSON to TOON
  echo '{"id": 1}' | j2t
  j2t config.json

  # JSONL with jq
  cat data.jsonl | jq -s '.' | j2t
  cat data.jsonl | jq -c '.' | j2t -stream

  # TOON to JSON (reverse)
  echo "id: 1" | j2t -r
  j2t -r config.toon
`)
	}
	flag.Parse()

	var input io.Reader
	var data []byte
	var err error

	args := flag.Args()
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

	if reverse {
		// Reverse mode: TOON to JSON
		result, err = json2toon.ConvertToJSON(data)
	} else {
		// Forward mode: JSON/JSONC/JSONL to TOON
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
			// Streaming mode: process input line-by-line
			result, err = json2toon.ConvertJSONLStreamWithOptions(data, opts...)
		} else {
			// Auto-detect format and convert using library function
			result, err = json2toon.ConvertAutoWithOptions(data, opts...)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(result)
}
