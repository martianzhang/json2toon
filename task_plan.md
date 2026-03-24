# Task Plan: json2toon - JSON to TOON Converter Library

## Goal
Build a Go library that converts JSON/JSONC to TOON format, with streaming support for large files.

## Key Decisions

1. **Streaming-first design**: Use `json.Decoder` for token-based parsing (not `json.Unmarshal`)
2. **JSONC pre-processing**: Strip comments before feeding to JSON decoder
3. **Tabular detection**: Auto-detect uniform object arrays → convert to TOON table format
4. **Encoder options**: Configurable indent, delimiter, key folding

## Project Structure

```
json2toon/
├── go.mod
├── task_plan.md
├── SPEC.md                          # TOON format spec
├── encoder.go                       # TOON encoder (writes TOON)
├── encoder_options.go               # Encoder configuration
├── converter.go                     # JSON/JSONC → TOON converter
├── converter_options.go             # Converter configuration
├── jsonc.go                         # JSONC comment stripper
├── json2toon.go                     # Public API
├── json2toon_test.go                # Unit tests
├── bench_test.go                    # Benchmarks
└── examples_test.go                 # Usage examples
```

## API Design

```go
// Simple conversion
func Convert(jsonBytes []byte) ([]byte, error)
func ConvertString(jsonStr string) (string, error)

// Streaming conversion
func NewConverter(w io.Writer) *Converter
func (c *Converter) WriteJSON(jsonBytes []byte) error
func (c *Converter) Close() error

// JSONC support
func ConvertJSONC(jsoncBytes []byte) ([]byte, error)

// File conversion
func ConvertFile(path string) ([]byte, error)
func ConvertFileToWriter(path string, w io.Writer) error
```

## Encoder Options

```go
type EncoderOptions struct {
    Indent      int    // Spaces per level (default: 2)
    Delimiter   rune   // Table delimiter: ',', '\t', '|' (default: ',')
    KeyFolding  string // "off" or "safe"
    FlattenDepth int   // Max fold depth (default: unlimited)
}
```

## Converter Options

```go
type ConverterOptions struct {
    Encoder EncoderOptions
    Strict  bool   // Enforce strict TOON spec (default: true)
}
```

## Phases

- [x] Phase 1: Plan and setup (create task_plan.md, SPEC.md, go.mod)
- [ ] Phase 2: Implement TOON encoder (encoder.go, encoder_options.go)
- [ ] Phase 3: Implement JSONC pre-processor (jsonc.go)
- [ ] Phase 4: Implement converter (converter.go, converter_options.go)
- [ ] Phase 5: Implement public API (json2toon.go)
- [ ] Phase 6: Add tests and benchmarks
- [ ] Phase 7: Verify build and test

## TOON Format Key Points

1. **Key-value**: `key: value` (one space after colon)
2. **Objects**: Indentation-based (2 spaces), no braces
3. **Arrays**: `[N]:` header + comma-separated values OR tabular format
4. **Tabular arrays**: `{field1,field2}:` header + CSV rows
5. **Comments**: `#` to end of line (full-line and inline)
6. **Quoting**: Must quote if empty, whitespace, true/false/null, numeric-like, contains `:`, `"`, `\`, `[`, `]`, `{`, `}`
7. **Escapes**: Only `\\`, `\"`, `\n`, `\r`, `\t`
8. **Numbers**: Normalize exponents, no leading zeros, `-0` → `0`

## Status
**Currently in Phase 1** - Creating task_plan.md and SPEC.md
