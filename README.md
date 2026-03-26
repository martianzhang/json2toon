<p align="center">
  <img src="json2toon.png" alt="JSON2TOON" />
</p>

# JSON2TOON

JSON/JSONC/JSONL to TOON (and back) converter library written in Go.

[![GoDoc](https://pkg.go.dev/badge/github.com/martianzhang/json2toon)](https://pkg.go.dev/github.com/martianzhang/json2toon) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![TOON Format](https://img.shields.io/badge/TOON%20Format-toonformat.dev-blue)](https://toonformat.dev)

## Features

- **Bidirectional conversion** - Convert JSON to TOON and back
- **Streaming support** - Token-based JSON parsing for large files
- **JSONC support** - Handles `//` and `/* */` comments before parsing
- **JSONL support** - Auto-detects and converts JSON Lines format
- **TOON format** - Produces valid TOON output ([Format Overview](https://toonformat.dev/guide/format-overview.html))
- **Configurable** - Indent, delimiter, key folding, strict mode
- **CLI tool** - `j2t` command for convenient file/stdin conversion

## Installation

```bash
go get github.com/martianzhang/json2toon
```

## CLI

```bash
# Install
go install github.com/martianzhang/json2toon/cmd/j2t@latest

# Convert file (auto-detects JSON/JSONC/JSONL)
j2t config.json
j2t config.jsonc
j2t data.jsonl

# With options
j2t -indent 4 -delimiter $'\t' -key-folding safe config.json

# Stdin
echo '{"id": 1}' | j2t
```

### CLI Options

| Flag | Default | Description |
|------|---------|-------------|
| `-o` | stdout | Output file path |
| `-e` | auto | Force encode mode (JSON to TOON) |
| `-d`, `-r` | auto | Force decode mode (TOON to JSON) |
| `-indent` | `2` | Spaces per indentation level |
| `-delimiter` | `,` | Table delimiter (`,`, `\t`, `\|`) |
| `-key-folding` | (none) | Key folding mode (`off`, `safe`) |
| `-strict` | `false` | Enable strict TOON validation |
| `-stream` | `false` | Streaming mode for JSONL (line-by-line, memory-efficient) |
| `--stats` | `false` | Show token count estimates and savings |
| `--expandPaths` | (none) | Path expansion mode on decode (`off`, `safe`) |

### File Extension Detection

The CLI auto-detects operation based on file extension:

- `.json`, `.jsonc`, `.jsonl` → encode (JSON to TOON)
- `.toon` → decode (TOON to JSON)

### Using with jq

```bash
# Convert JSONL by merging lines into array first
cat data.jsonl | jq -s '.' | j2t

# Convert each JSONL object individually (keeps objects separate)
cat data.jsonl | jq -c '.' | j2t -stream

# Reverse conversion: TOON to JSON
echo 'id: 123' | j2t -r
# Output: { "id": 123 }

# Convert TOON file to JSON (auto-detected from .toon extension)
j2t config.toon -o config.json

# Token statistics
j2t data.json --stats

# Output to file
j2t config.json -o config.toon

# Path expansion for lossless round-trip with key folding
j2t compressed.toon --expandPaths safe -o output.json
```

## Library Quick Start

```go
package main

import (
	"fmt"
	"github.com/martianzhang/json2toon"
)

func main() {
	json := `{"id": 123, "name": "Ada", "active": true}`
	result, err := json2toon.Convert([]byte(json))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))
}
```

Output:

```toon
id: 123
name: Ada
active: true
```

## API Reference

All API details (function signatures, options, examples) are on the **[godoc](https://pkg.go.dev/github.com/martianzhang/json2toon)** page. Key highlights:

- **Basic conversion**: `Convert`, `ConvertString`, `ConvertFile`, `ConvertFileToWriter`
- **Format-specific**: `ConvertJSONC`, `ConvertJSONL`, `ConvertJSONLStream` (and `...String` / `...WithOptions` variants)
- **Auto-detection**: `ConvertAuto`, `ConvertAutoFile` (detect JSON/JSONC/JSONL automatically)
- **Reverse conversion**: `ConvertToJSON`, `ConvertToJSONString`, `ConvertToJSONFile`, `ConvertToJSONFromReader`
- **Streaming**: `NewConverter`, `ConvertJSONLStream` on a `Converter` instance
- **Options**: `WithIndent`, `WithDelimiter`, `WithKeyFolding`, `WithStrict` (encoder); `WithDecodeStrict`, `WithExpandPaths` (decoder)

## TOON Format

For detailed format specification, see the [official TOON documentation](https://toonformat.dev/guide/getting-started.html).

## Makefile Targets

```bash
make test      # Run tests
make lint      # Run golangci-lint
make build     # Build CLI
make release   # Cross-compile for Linux/macOS/Windows
```

## Benchmarks

```
BenchmarkConvertSimple-8          8137 ns/op     6064 B/op      60 allocs/op
BenchmarkConvertNested-8         11511 ns/op     6576 B/op      87 allocs/op
BenchmarkConvertArray-8          69058 ns/op    25496 B/op     834 allocs/op
BenchmarkConvertTabular-8       313393 ns/op    97368 B/op    3328 allocs/op
BenchmarkConvertLargeFile-8     919854 ns/op   360137 B/op   10379 allocs/op
BenchmarkConvertJSONC-8          10348 ns/op     6664 B/op      82 allocs/op
BenchmarkConvertJSONL-8        1720938 ns/op    83922 B/op    1866 allocs/op
BenchmarkConvertJSONLStream-8    228059 ns/op    95323 B/op    2657 allocs/op
BenchmarkConvertAuto-8            9513 ns/op     7264 B/op      96 allocs/op
BenchmarkConvertToJSON-8          8054 ns/op     9678 B/op      40 allocs/op
```

## License

MIT
