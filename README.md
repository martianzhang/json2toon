# json2toon

JSON/JSONC/JSONL to TOON (and back) converter library written in Go.

**[godoc](https://pkg.go.dev/github.com/martianzhang/json2toon) · [TOON Format](https://toonformat.dev) · [Format Overview](https://toonformat.dev/guide/format-overview.html)**

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

- `.json`, `.jsonc`, `.jsonl`, `.jsonl.gz` → encode (JSON to TOON)
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

For complete documentation, see the **[godoc](https://pkg.go.dev/github.com/martianzhang/json2toon)**.

### Basic Conversion (JSON → TOON)

```go
// JSON bytes to TOON bytes
result, err := json2toon.Convert(jsonBytes)

// JSON string to TOON string
result, err := json2toon.ConvertString(jsonStr)

// Convert JSON file to TOON bytes
result, err := json2toon.ConvertFile("config.json")

// Convert JSON file and write directly to an io.Writer
err := json2toon.ConvertFileToWriter("config.json", writer)
```

### Format-Specific Conversion

```go
// JSONC (with comments) to TOON
result, err := json2toon.ConvertJSONC(jsoncBytes)
result, err := json2toon.ConvertJSONCString(jsoncStr)

// JSON Lines to TOON
result, err := json2toon.ConvertJSONL(jsonlBytes)
result, err := json2toon.ConvertJSONLString(jsonlStr)

// Streaming JSONL (line-by-line, more memory-efficient for large files)
result, err := json2toon.ConvertJSONLStream(jsonlBytes)
result, err := json2toon.ConvertJSONLStreamString(jsonlStr)
```

### Auto-Detection Conversion

```go
// Auto-detect format (JSON/JSONC/JSONL) and convert to TOON
result, err := json2toon.ConvertAuto(data)

// Auto-detect and convert a file
result, err := json2toon.ConvertAutoFile("config.json")
```

### Reverse Conversion (TOON → JSON)

```go
// TOON bytes to JSON bytes
result, err := json2toon.ConvertToJSON(toonBytes)

// TOON string to JSON string
result, err := json2toon.ConvertToJSONString(toonStr)

// TOON from reader to JSON bytes
result, err := json2toon.ConvertToJSONFromReader(reader)

// TOON file to JSON bytes
result, err := json2toon.ConvertToJSONFile("config.toon")
```

### Streaming

```go
converter := json2toon.NewConverter(writer)
defer func() { _ = converter.Close() }()

err := converter.ConvertJSON(jsonBytes)
// or
err := converter.ConvertJSONC(jsoncBytes)

// Streaming JSONL (line-by-line)
err := converter.ConvertJSONLStream(reader)
```

### Options

```go
// Custom indentation
result, err := json2toon.ConvertWithOptions(json,
	json2toon.WithIndent(4))

// Custom delimiter
result, err := json2toon.ConvertWithOptions(json,
	json2toon.WithDelimiter('|'))

// Key folding
result, err := json2toon.ConvertWithOptions(json,
	json2toon.WithKeyFolding("safe"))

// Strict mode (encode)
result, err := json2toon.ConvertWithOptions(json,
	json2toon.WithStrict(true))

// Strict mode (decode) - validates TOON syntax strictly
result, err := json2toon.ConvertToJSONWithOptions(toon,
	json2toon.WithDecodeStrict(true))

// Path expansion (decode) - expands folded keys back to nested objects
result, err := json2toon.ConvertToJSONWithOptions(toon,
	json2toon.WithExpandPaths(true))
```

Each conversion function has a `WithOptions` variant (e.g., `ConvertJSONCWithOptions`, `ConvertJSONLWithOptions`, `ConvertAutoWithOptions`) that accepts the same encoder options.

## TOON Format

For detailed format specification, see the [official TOON documentation](https://toonformat.dev/guide/getting-started.html).

### Round-trip Conversion

You can convert between JSON and TOON bidirectionally:

```go
// JSON → TOON
toon, _ := json2toon.Convert(jsonBytes)

// TOON → JSON
json, _ := json2toon.ConvertToJSON(toon)
```

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
