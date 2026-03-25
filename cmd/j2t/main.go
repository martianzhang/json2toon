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
	)

	flag.IntVar(&indent, "indent", 2, "indentation spaces")
	flag.StringVar(&delimiter, "delimiter", ",", "table delimiter character")
	flag.StringVar(&keyFolding, "key-folding", "", "key folding mode (empty, \"safe\", \"upper\", \"lower\")")
	flag.BoolVar(&strict, "strict", false, "strict mode")
	flag.BoolVar(&stream, "stream", false, "enable streaming mode for JSONL (line-by-line processing)")
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
  echo '["a","b"]' | jq -r '.[]' | j2t -stream
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

	var result []byte

	if stream {
		// Streaming mode: process input line-by-line
		result, err = json2toon.ConvertJSONLStreamWithOptions(data, opts...)
	} else {
		// Auto-detect format and convert using library function
		result, err = json2toon.ConvertAutoWithOptions(data, opts...)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(result)
}
