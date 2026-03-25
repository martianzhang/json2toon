# json2toon

JSON/JSONC/JSONL to TOON converter library written in Go.

## Features

- **Streaming support** - Token-based JSON parsing for large files
- **JSONC support** - Handles `//` and `/* */` comments before parsing
- **JSONL support** - Auto-detects and converts JSON Lines format
- **TOON v3.0 format** - Produces valid TOON output
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
| `-indent` | `2` | Spaces per indentation level |
| `-delimiter` | `,` | Table delimiter (`,`, `\t`, `\|`) |
| `-key-folding` | `""` | Key folding mode (`""`, `"safe"`, `"upper"`, `"lower"`) |
| `-strict` | `false` | Enable strict TOON validation |

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

### Basic Conversion

```go
// JSON bytes to TOON bytes
result, err := json2toon.Convert(jsonBytes)

// JSON string to TOON string
result, err := json2toon.ConvertString(jsonStr)

// JSONC (with comments) to TOON
result, err := json2toon.ConvertJSONC(jsoncBytes)

// JSONC string to TOON string
result, err := json2toon.ConvertJSONCString(jsoncStr)

// Convert file (auto-detects JSON/JSONC/JSONL)
result, err := json2toon.ConvertFile("config.json")

// Convert file to writer
err := json2toon.ConvertFileToWriter("config.json", writer)

// JSON Lines to TOON
result, err := json2toon.ConvertJSONL(jsonLines)
```

### Streaming

```go
converter := json2toon.NewConverter(writer)
defer converter.Close()

err := converter.ConvertJSON(jsonBytes)
// or
err := converter.ConvertJSONC(jsoncBytes)
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

// Strict mode
result, err := json2toon.ConvertWithOptions(json,
    json2toon.WithStrict(true))
```

## TOON Format

TOON (Token-Oriented Object Notation) is a line-oriented, indentation-based format:

| JSON | TOON |
|------|------|
| `{"id": 1}` | `id: 1` |
| `{"items": [1,2,3]}` | `items[3]: 1,2,3` |
| `[{"id":1},{"id":2}]` | `[2]{id}:\n  1\n  2` |

## Makefile Targets

```bash
make test      # Run tests
make lint      # Run golangci-lint
make build     # Build CLI
make release   # Cross-compile for Linux/macOS/Windows
```

## Benchmarks

```
BenchmarkConvertSimple       277158 ns/op    6080 B/op    63 allocs/op
BenchmarkConvertNested       172875 ns/op    6608 B/op    90 allocs/op
BenchmarkConvertArray         22027 ns/op   26984 B/op   934 allocs/op
BenchmarkConvertLargeFile     165196 ns/op    6456 B/op    62 allocs/op
```

## License

MIT
