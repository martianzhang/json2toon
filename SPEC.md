# TOON Format Specification (v3.0)

**Reference**: https://github.com/toon-format/spec  
**Version**: 3.0 (2025-11-24)

## Overview

TOON (Token-Oriented Object Notation) is a line-oriented, indentation-based text format that encodes the JSON data model with explicit structure and minimal quoting. Designed for 30-60% token reduction vs JSON.

## Data Types

| Type | JSON | TOON |
|------|------|------|
| String | `"hello"` | `hello` |
| Number | `42`, `3.14` | `42`, `3.14` |
| Boolean | `true`, `false` | `true`, `false` |
| Null | `null` | `null` |
| Object | `{"id": 1}` | `id: 1` |
| Array | `[1, 2, 3]` | `nums[3]: 1,2,3` |

## Syntax Rules

### Key-Value Pairs
```
key: value
```
- One space after colon (required)
- Keys: `^[A-Za-z_][A-Za-z0-9_.]*$`

### Indentation
- 2 spaces per level (configurable)
- Tabs NOT allowed
- Consistent indentation required

### Comments
```
# Full-line comment
key: value  # Inline comment
```

## Object Syntax

### Simple
```json
{"id": 123, "name": "Ada", "active": true}
```
```toon
id: 123
name: Ada
active: true
```

### Nested
```json
{"user": {"id": 123, "name": "Ada"}}
```
```toon
user:
  id: 123
  name: Ada
```

## Array Syntax

### Primitive Arrays (Inline)
```json
{"tags": ["admin", "ops", "dev"]}
```
```toon
tags[3]: admin,ops,dev
```

### Tabular Arrays (Uniform Objects)
```json
{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}
```
```toon
users[2]{id,name}:
  1,Alice
  2,Bob
```

**Tabular Requirements:**
- Every element is an object
- All objects have same keys
- All values are primitives

### Arrays of Arrays
```json
{"matrix": [[1, 2], [3, 4]]}
```
```toon
matrix[2]:
  - [2]: 1,2
  - [2]: 3,4
```

### Mixed Arrays
```json
{"items": [{"id": 1}, "plain", {"id": 2}]}
```
```toon
items[3]:
  - id: 1
  - plain
  - id: 2
```

## Delimiters

| Delimiter | Header | Use Case |
|-----------|--------|----------|
| Comma | `key[N]:` | Default |
| Tab | `key[N\t]:` | Data with commas |
| Pipe | `key[N\|]:` | Data with commas AND tabs |

## Quoting Rules

### MUST Quote If:
1. Empty string: `""`
2. Leading/trailing whitespace
3. Equals `true`, `false`, or `null`
4. Numeric-like: `"42"`, `"-3.14"`, `"05"`
5. Contains: `:`, `"`, `\`
6. Contains: `[`, `]`, `{`, `}`
7. Contains control characters
8. Contains active delimiter
9. Equals `-` or starts with `-`

### Valid Escapes
```
\\  → backslash
\"  → double quote
\n  → newline
\r  → carriage return
\t  → tab
```

## Number Formatting (Encoding)

- No exponent notation: `1e6` → `1000000`
- No leading zeros: `05` → invalid
- No trailing zeros: `1.500` → `1.5`
- Fractional zero → integer: `1.0` → `1`
- Normalize `-0` → `0`

**Decoders MUST accept**: `42`, `-3.14`, `1e-6`, `-1E+9`

## Key Folding (Optional)

```toon
# Without folding
data:
  meta:
    items[2]: a,b

# With folding
data.meta.items[2]: a,b
```

## Encoder Options

| Option | Default | Description |
|--------|---------|-------------|
| `indent` | `2` | Spaces per indentation level |
| `delimiter` | `,` | Document delimiter (`,`, `\t`, `\|`) |
| `keyFolding` | `"off"` | `"off"` or `"safe"` |
| `flattenDepth` | `Infinity` | Max segments to fold |

## File Format

| Property | Value |
|----------|-------|
| Extension | `.toon` |
| Media Type | `text/toon` |
| Encoding | UTF-8 |
| Line endings | LF (`\n`) |
| No trailing newline | ✓ |

## Strict Mode Errors

| Error | Condition |
|-------|-----------|
| Count mismatch | Inline array values ≠ N |
| Width mismatch | Row values ≠ field count |
| Missing colon | Key without `:` |
| Invalid escape | Unknown escape sequence |
| Unterminated string | Missing closing quote |
| Bad indentation | Not multiple of indentSize |
| Tab in indent | Tabs used for indentation |
| Blank in array | Empty line inside array |
