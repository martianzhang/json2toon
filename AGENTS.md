# Agent Guidelines for json2toon

This document provides guidelines for agents working in this repository.

## Project Overview

**json2toon** is a Go library that converts JSON/JSONC to TOON (Token-Oriented Object Notation) format.
- Module: `github.com/martianzhang/json2toon`
- Go version: 1.25.0
- Standard library only (no external dependencies)

## Build, Lint, and Test Commands

### Build
```bash
go build ./...
```

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run a Single Test
```bash
go test -v -run TestFunctionName
```

### Run Benchmarks
```bash
go test -bench=. -benchmem ./...
```

### Run Specific Benchmark
```bash
go test -bench=BenchmarkName -benchmem ./...
```

### Lint (requires golangci-lint)
```bash
golangci-lint run
```

### Format Code
```bash
go fmt ./...
```

### Tidy Dependencies
```bash
go mod tidy
```

## Code Style Guidelines

### General

- **Package name**: `json2toon` (no underscores, lowercase)
- **Internal functions**: Can be unexported (lowercase)
- **API functions**: Must be exported (uppercase)
- **File naming**: Use descriptive lowercase names, e.g., `encoder.go`, `converter.go`

### Imports

- Use standard library exclusively
- Group imports: stdlib first, then blank line
- Use `go fmt` for automatic formatting

```go
import (
    "bufio"
    "encoding/json"
    "errors"
    "io"
    "strconv"
    "strings"
)
```

### Types

- Use specific types over `interface{}` when possible
- For JSON compatibility, use `map[string]interface{}` and `[]interface{}`
- Use `float64` for JSON numbers (Go's JSON decoder behavior)
- Use `json.Delim` for JSON delimiters (`{`, `}`, `[`, `]`)

### Naming Conventions

- **Variables/Functions**: `camelCase`
- **Types/Interfaces**: `PascalCase`
- **Constants**: `PascalCase` or `ALL_CAPS` for true constants
- **Acronyms**: Keep original casing (e.g., `ID`, `URL`, not `Id`, `Url`)
- **Receiver names**: 1-2 chars, e.g., `e` for Encoder, `c` for Converter

### Error Handling

- Return errors from functions when failure is possible
- Use `fmt.Errorf` with context for wrapped errors
- Never silently ignore errors with `_`
- Check errors immediately after operations

```go
// Good
if err != nil {
    return fmt.Errorf("failed to decode: %w", err)
}

// Bad - ignoring error
data, _ := json.Marshal(v) // Don't do this
```

### Documentation

- Document all exported types and functions with comments
- Use complete sentences with proper punctuation
- Start with the name of the thing being documented

```go
// Encoder writes TOON format to an io.Writer.
type Encoder struct {
    ...
}

// NewEncoder creates a new Encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
    ...
}
```

### Functions

- Keep functions focused (single responsibility)
- Use functional options pattern for configuration
- Prefer early returns over deeply nested conditionals

### Testing

- Test file naming: `<package>_test.go`
- Test function naming: `TestFunctionName`
- Benchmark naming: `BenchmarkFunctionName`
- Table-driven tests for multiple test cases
- Use `t.Run` for subtests

```go
func TestConvertBasicTypes(t *testing.T) {
    tests := []struct {
        name     string
        json     string
        expected string
    }{
        {"integer", `42`, "42"},
        {"float", `3.14`, "3.14"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Convert([]byte(tt.json))
            if err != nil {
                t.Fatalf("Convert failed: %v", err)
            }
            if string(result) != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### TOON Format Notes

TOON (Token-Oriented Object Notation) key rules:
- Key-value: `key: value` (one space after colon)
- Objects: indentation-based (default 2 spaces)
- Primitive arrays: `[N]: value1,value2,value3`
- Tabular arrays: `[N]{field1,field2}:\n  val1,val2\n  val3,val4`
- Comments: `#` to end of line
- Numbers: normalize exponents, no leading zeros
- Strings: quote if contains delimiter, `:`, `"`, `\`, or is numeric-like

## File Structure

```
json2toon/
├── encoder.go           # Core TOON encoder
├── encoder_options.go   # Encoder configuration
├── converter.go         # JSON/JSONC to TOON converter
├── converter_options.go # Converter configuration
├── jsonc.go            # JSONC comment stripping
├── json2toon.go        # Public API
├── json2toon_test.go  # Unit tests
├── bench_test.go       # Benchmarks
├── go.mod
└── README.md
```

## Common Patterns

### Functional Options Pattern

```go
type ConverterOption func(*ConverterOptions)

func WithIndent(n int) ConverterOption {
    return func(opts *ConverterOptions) {
        opts.Encoder.Indent = n
    }
}
```

### Streaming with json.Decoder

```go
decoder := json.NewDecoder(reader)
for decoder.More() {
    token, err := decoder.Token()
    if err != nil {
        return err
    }
    // Process token
}
```

## Don'ts

- Don't add external dependencies
- Don't use `interface{}` without good reason
- Don't suppress errors with `_`
- Don't leave TODO comments (implement or file an issue)
- Don't commit generated files (coverage.out, etc.)
