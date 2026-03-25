package json2toon

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// DecoderError represents a decoding error with context information.
type DecoderError struct {
	Type    string // Error type (e.g., "InvalidSyntax", "CountMismatch")
	Message string // Human-readable error message
	Line    int    // Line number (1-indexed, 0 if unknown)
	Column  int    // Column position (1-indexed, 0 if unknown)
	Context string // Input snippet for context
}

// Error implements the error interface.
func (e *DecoderError) Error() string {
	if e.Line > 0 {
		if e.Column > 0 {
			return fmt.Sprintf("%s at line %d, column %d: %s", e.Type, e.Line, e.Column, e.Message)
		}
		return fmt.Sprintf("%s at line %d: %s", e.Type, e.Line, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error (for error wrapping compatibility).
func (e *DecoderError) Unwrap() error {
	return nil
}

// NewDecoderError creates a new DecoderError with context.
func NewDecoderError(errType, message string) *DecoderError {
	return &DecoderError{
		Type:    errType,
		Message: message,
	}
}

// NewDecoderErrorWithContext creates a DecoderError with line/column context.
func NewDecoderErrorWithContext(errType, message string, line, column int, context string) *DecoderError {
	return &DecoderError{
		Type:    errType,
		Message: message,
		Line:    line,
		Column:  column,
		Context: context,
	}
}

// Decoder reads TOON format and converts to JSON.
type Decoder struct {
	r           *bufio.Reader
	indent      int
	strict      bool
	expandPaths bool
}

// DecoderOptions holds decoder configuration.
type DecoderOptions struct {
	Strict      bool
	ExpandPaths bool
}

// NewDecoder creates a new Decoder reading from r.
// The decoder will use a default indent of 2 spaces and strict validation.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      bufio.NewReader(r),
		indent: 2,
		strict: true,
	}
}

// NewDecoderWithOptions creates a new Decoder with custom options.
// Use DefaultDecodeOptions() as a starting point and modify as needed.
func NewDecoderWithOptions(r io.Reader, opts DecodeOptions) *Decoder {
	return &Decoder{
		r:           bufio.NewReader(r),
		indent:      2,
		strict:      opts.Strict,
		expandPaths: opts.ExpandPaths,
	}
}

// Decode reads TOON from the reader and returns a JSON-compatible value.
// The returned value will be one of:
//   - map[string]interface{} for objects
//   - []interface{} for arrays
//   - string, float64, bool, or nil for primitives
//
// Returns (nil, nil) if the input is empty.
//
// On error, the returned error will be a *DecoderError with context information
// including line numbers when available.
func (d *Decoder) Decode() (interface{}, error) {
	scanner := bufio.NewScanner(d.r)
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(allLines) == 0 {
		return nil, nil
	}

	// Detect indent from first line with meaningful content
	for _, line := range allLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		leading := len(line) - len(strings.TrimLeft(line, " "))
		if leading > 0 {
			d.indent = leading
			break
		}
	}

	// Parse the document
	return d.parseLines(allLines)
}

// parseLines parses lines starting at any depth.
func (d *Decoder) parseLines(lines []string) (interface{}, error) {
	// Check if this is a list (all non-empty/comment lines start with "- ")
	hasListItems := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			hasListItems = true
			// Check if this is "- key: value" format
			afterDash := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if strings.Contains(afterDash, ":") {
				// This is list array format - use parseListArrayItems to extract values if same key
				return d.parseListArrayItems(lines), nil
			}
		}
		break
	}
	if hasListItems {
		// Standalone list items (no key:value format) - parse as objects
		return d.parseListItems(lines), nil
	}

	result := make(map[string]interface{})
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			i++
			continue
		}

		if trimmed == "---" {
			i++
			continue
		}

		// Check for tabular array header
		if strings.Contains(trimmed, "]{") {
			key, arr, consumed := d.parseTabular(lines[i:])
			result[key] = arr
			i += consumed
			continue
		}

		// Check for inline array: key[N]: val1,val2
		if strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]:") {
			key, values, err := d.parseInlineArray(trimmed)
			if err != nil {
				return nil, err
			}
			if key != "" {
				result[key] = values
				i++
				continue
			}
			// If key is empty but values exist, this is a root-level inline array
			if values != nil {
				return values, nil
			}
			// If key is empty and no values, this is a list array - fall through
		}

		// Check for key: value
		if strings.Contains(trimmed, ":") {
			colonIdx := strings.Index(trimmed, ":")
			key := trimmed[:colonIdx]
			valueStr := strings.TrimSpace(trimmed[colonIdx+1:])

			// Strip [N] from key if present (for arrays)
			if idx := strings.Index(key, "["); idx >= 0 {
				key = strings.TrimSpace(key[:idx])
			}

			if valueStr == "" {
				// Look for nested content
				nestedLines, consumed := d.collectNested(lines, i)
				if len(nestedLines) > 0 {
					// Parse nested content
					nested, err := d.parseLines(nestedLines)
					if err != nil {
						return nil, err
					}
					result[key] = nested
					i += consumed + 1 // +1 to skip the key line itself
					continue
				}
				// Empty content after key: means empty array (encoder format)
				result[key] = []interface{}{}
				i++
				continue
			}

			value := d.parseValue(valueStr)
			result[key] = value
			i++
			continue
		}

		// Standalone value
		return d.parseValue(trimmed), nil
	}

	return result, nil
}

// parseInlineArray parses "key[3]: val1,val2" or "key:[3]: val1,val2"
// Returns (key, values, nil) for valid inline arrays
// Returns ("", nil, nil) if there's no values on the same line (list array instead)
func (d *Decoder) parseInlineArray(line string) (string, []interface{}, error) {
	// Find the first colon
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return "", nil, nil
	}

	key := strings.TrimSpace(line[:colonIdx])
	rest := strings.TrimSpace(line[colonIdx+1:])

	// If no values on the same line, this is a list array, not inline
	if rest == "" {
		return "", nil, nil
	}

	// Extract count from [N]
	var count int
	// Check for "[N]:" pattern (encoder format with or without leading space)
	if strings.HasPrefix(rest, "[") || strings.HasPrefix(strings.TrimSpace(rest), "[") {
		// Trim and find the pattern
		trimmedRest := strings.TrimSpace(rest)
		// "[N]: values" format - find closing ]
		if endIdx := strings.Index(trimmedRest, "]"); endIdx > 0 {
			countStr := trimmedRest[1:endIdx]
			c, err := strconv.Atoi(countStr)
			if err != nil {
				// Invalid count format
				return "", nil, NewDecoderError("InvalidSyntax",
					fmt.Sprintf("invalid array count format: %q", countStr))
			}
			count = c
			rest = strings.TrimSpace(trimmedRest[endIdx+1:])
			if strings.HasPrefix(rest, ":") {
				rest = strings.TrimSpace(rest[1:])
			}
		} else {
			// Unclosed bracket
			return "", nil, NewDecoderError("UnclosedBracket",
				"unclosed bracket in array count specification")
		}
	} else if idx := strings.Index(key, "["); idx >= 0 {
		// "key[N]:" format
		keyName := strings.TrimSpace(key[:idx])
		countStr := strings.TrimSpace(key[idx+1:])
		if countStr != "" {
			countStr = strings.TrimSuffix(countStr, "]")
			c, err := strconv.Atoi(countStr)
			if err != nil {
				// Invalid count format
				return "", nil, NewDecoderError("InvalidSyntax",
					fmt.Sprintf("invalid array count format: %q", countStr))
			}
			count = c
		}
		key = keyName
	}

	values := d.parseCommaList(rest, ',')

	// Validate count if specified and not zero (empty array)
	if count > 0 && len(values) != count {
		err := NewDecoderError("CountMismatch",
			fmt.Sprintf("inline array count mismatch: expected %d, got %d", count, len(values)))
		return "", nil, err
	}

	return key, values, nil
}

// parseTabular parses a tabular array
func (d *Decoder) parseTabular(lines []string) (string, []interface{}, int) {
	if len(lines) == 0 {
		return "", nil, 0
	}

	header := strings.TrimSpace(lines[0])

	// Extract key and fields
	braceStart := strings.Index(header, "{")
	braceEnd := strings.Index(header, "}")
	if braceStart < 0 || braceEnd < 0 || braceEnd <= braceStart {
		return "", nil, 0
	}

	fields := strings.Split(header[braceStart+1:braceEnd], ",")

	beforeBrace := strings.TrimSpace(header[:braceStart])
	bracketIdx := strings.LastIndex(beforeBrace, "[")
	if bracketIdx < 0 {
		return "", nil, 0
	}
	key := strings.TrimSpace(beforeBrace[:bracketIdx])

	// Parse rows
	rows := make([]interface{}, 0)
	consumed := 1

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check indentation - if less than one indent, we've left the tabular
		leading := len(lines[i]) - len(line)
		if leading < d.indent || line == "---" || strings.Contains(line, ":") {
			break
		}

		values := d.parseCommaList(line, ',')
		if len(values) != len(fields) {
			break
		}

		row := make(map[string]interface{})
		for j, f := range fields {
			row[strings.TrimSpace(f)] = values[j]
		}
		rows = append(rows, row)
		consumed++
	}

	return key, rows, consumed
}

// parseListItems parses list items
func (d *Decoder) parseListItems(lines []string) []interface{} {
	var items []interface{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "- ") {
			content := strings.TrimPrefix(trimmed, "- ")
			// Check if it's a key: value
			if strings.Contains(content, ":") {
				obj := make(map[string]interface{})
				colonIdx := strings.Index(content, ":")
				k := content[:colonIdx]
				v := strings.TrimSpace(content[colonIdx+1:])
				obj[k] = d.parseValue(v)
				items = append(items, obj)
			} else {
				items = append(items, d.parseValue(content))
			}
		}
	}

	return items
}

// parseListArrayItems parses "- key: value" items for list arrays
// Returns values only if all items have the same key and no nested content, otherwise returns objects
func (d *Decoder) parseListArrayItems(lines []string) []interface{} {
	type listItem struct {
		key       string
		value     interface{}
		nested    map[string]interface{}
		hasNested bool
	}

	var items []listItem
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}

		if strings.HasPrefix(line, "- ") {
			content := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			item := listItem{nested: make(map[string]interface{})}

			if strings.Contains(content, ":") {
				colonIdx := strings.Index(content, ":")
				item.key = content[:colonIdx]
				valueStr := strings.TrimSpace(content[colonIdx+1:])
				item.value = d.parseValue(valueStr)

				// Collect nested content for this list item
				j := i + 1
				for j < len(lines) {
					nextLine := lines[j]
					trimmed := strings.TrimSpace(nextLine)
					if trimmed == "" || strings.HasPrefix(trimmed, "#") {
						j++
						continue
					}

					// Check if this is a new list item (next - at same or lesser indentation)
					if strings.HasPrefix(trimmed, "- ") {
						break
					}

					leading := len(nextLine) - len(trimmed)

					// If back to base level (no indentation), stop
					if leading == 0 {
						break
					}

					// Nested content (has indentation)
					if strings.Contains(trimmed, ":") {
						nc := strings.Index(trimmed, ":")
						nk := trimmed[:nc]
						nv := strings.TrimSpace(trimmed[nc+1:])

						// Check if value is an inline array
						if strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]:") {
							_, values, err := d.parseInlineArray(trimmed)
							if err == nil && values != nil {
								item.nested[nk] = values
								item.hasNested = true
								j++
								continue
							}
						}

						item.nested[nk] = d.parseValue(nv)
						item.hasNested = true
					}
					j++
				}
			} else {
				item.value = d.parseValue(content)
			}
			items = append(items, item)
		}
		i++
	}

	// Determine result format
	// If any item has nested content, ALL items become objects with full content
	anyHasNested := false
	for _, item := range items {
		if item.hasNested {
			anyHasNested = true
			break
		}
	}

	// If items have nested content, check if all have same key and should be extracted
	allSameKey := len(items) > 0
	firstKey := ""
	if len(items) > 0 {
		firstKey = items[0].key
	}
	for _, item := range items {
		if item.key != firstKey {
			allSameKey = false
			break
		}
	}

	var result []interface{}
	for _, item := range items {
		if anyHasNested {
			// Build object with key:value and nested content
			obj := make(map[string]interface{})
			if item.key != "" {
				obj[item.key] = item.value
			}
			for k, v := range item.nested {
				obj[k] = v
			}
			result = append(result, obj)
		} else if allSameKey && item.key != "" {
			// Extract just the values when all items have the same key
			result = append(result, item.value)
		} else {
			// Build objects for mixed or no-key items
			if item.key == "" {
				result = append(result, item.value)
			} else {
				obj := make(map[string]interface{})
				obj[item.key] = item.value
				result = append(result, obj)
			}
		}
	}

	return result
}

// collectNested collects lines that belong to a nested block
func (d *Decoder) collectNested(lines []string, startIdx int) ([]string, int) {
	if startIdx+1 >= len(lines) {
		return nil, 0
	}

	var nested []string
	consumed := 0
	minDepth := -1      // Will be set to the depth of the first content line
	listItemDepth := -1 // Depth of list items (- prefix)

	for i := startIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		leading := len(line) - len(strings.TrimLeft(line, " "))
		depth := leading / d.indent

		// List items (- prefix) - set listItemDepth on first one
		if strings.HasPrefix(trimmed, "- ") {
			if listItemDepth < 0 {
				listItemDepth = depth
				minDepth = depth
			}
			// Collect list item
			nested = append(nested, line)
			consumed++
			continue
		}

		// If we started with list items, only collect lines at listItemDepth or deeper
		if listItemDepth >= 0 {
			if depth < listItemDepth {
				// Back to parent level - stop
				break
			}
			// At listItemDepth or deeper - collect
			nested = append(nested, line)
			consumed++
			continue
		}

		// Set minDepth on first non-list content
		if minDepth < 0 {
			minDepth = depth
		}

		// If depth is less than minDepth, we've left the current block
		if depth < minDepth {
			break
		}

		// If depth is less than 1 (parent level), we've left
		if depth < 1 {
			break
		}

		nested = append(nested, line)
		consumed++
	}

	return nested, consumed
}

// parseCommaList parses comma-separated values
func (d *Decoder) parseCommaList(s string, delim rune) []interface{} {
	if strings.TrimSpace(s) == "" {
		return []interface{}{}
	}

	var result []interface{}
	var current strings.Builder
	inQuote := false

	for i, r := range s {
		if r == '"' {
			if inQuote && i > 0 && s[i-1] == '\\' {
				current.WriteRune(r)
			} else {
				inQuote = !inQuote
			}
		} else if r == delim && !inQuote {
			val := d.parseValue(strings.TrimSpace(current.String()))
			result = append(result, val)
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}

	if val := strings.TrimSpace(current.String()); val != "" {
		result = append(result, d.parseValue(val))
	}

	return result
}

// parseValue parses a single value
func (d *Decoder) parseValue(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Remove inline comments
	if idx := strings.Index(s, "#"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}

	if s == "" {
		return nil
	}

	// Quoted string
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return d.parseQuoted(s)
	}

	// Empty object literal
	if s == "{}" {
		return map[string]interface{}{}
	}

	// Null
	if s == "null" {
		return nil
	}

	// Boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Number
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return n
	}

	// String
	return s
}

// parseQuoted parses a quoted string with escapes
func (d *Decoder) parseQuoted(s string) string {
	s = s[1 : len(s)-1]

	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			default:
				result.WriteByte(s[i])
			}
			i += 2
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// DecodeToJSON reads TOON and returns JSON bytes.
func DecodeToJSON(toon []byte) ([]byte, error) {
	decoder := NewDecoder(bytes.NewReader(toon))
	value, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(value, "", "  ")
}

// DecodeToJSONString reads TOON and returns JSON string.
func DecodeToJSONString(toon string) (string, error) {
	result, err := DecodeToJSON([]byte(toon))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// DecodeFromReader reads TOON from reader and returns JSON bytes.
func DecodeFromReader(r io.Reader) ([]byte, error) {
	decoder := NewDecoder(r)
	value, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(value, "", "  ")
}

// DecodeToJSONWithOptions converts TOON to JSON with custom options.
func DecodeToJSONWithOptions(toon []byte, opts DecodeOptions) ([]byte, error) {
	decoder := NewDecoderWithOptions(bytes.NewReader(toon), opts)
	value, err := decoder.Decode()
	if err != nil {
		return nil, err
	}

	// Apply path expansion if enabled
	if decoder.expandPaths {
		value = expandPaths(value)
	}

	return json.MarshalIndent(value, "", "  ")
}

// expandPaths expands dotted keys into nested objects.
// For example: {"a.b.c": "value"} -> {"a": {"b": {"c": "value"}}}
func expandPaths(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			if strings.Contains(k, ".") {
				// Expand the key
				parts := strings.Split(k, ".")
				setNestedValue(result, parts, expandPaths(val))
			} else {
				result[k] = expandPaths(val)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = expandPaths(item)
		}
		return result
	default:
		return v
	}
}

// setNestedValue sets a value at a nested path, creating intermediate objects as needed.
// path ["a", "b", "c"] with value "v" -> {"a": {"b": {"c": "v"}}}
func setNestedValue(root map[string]interface{}, path []string, value interface{}) {
	if len(path) == 0 {
		return
	}
	if len(path) == 1 {
		root[path[0]] = value
		return
	}

	// Get or create the nested map
	if _, exists := root[path[0]]; !exists {
		root[path[0]] = make(map[string]interface{})
	}
	nested, ok := root[path[0]].(map[string]interface{})
	if !ok {
		// Can't expand into a non-map value, overwrite it
		root[path[0]] = make(map[string]interface{})
		nested, _ = root[path[0]].(map[string]interface{})
	}
	setNestedValue(nested, path[1:], value)
}
