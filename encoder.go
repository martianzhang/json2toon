package json2toon

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Encoder writes TOON format to an io.Writer.
type Encoder struct {
	w          *bufio.Writer
	opts       EncoderOptions
	depth      int
	firstField bool
}

// NewEncoder creates a new Encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	return NewEncoderWithOptions(w, DefaultEncoderOptions())
}

// NewEncoderWithOptions creates a new Encoder with custom options.
func NewEncoderWithOptions(w io.Writer, opts EncoderOptions) *Encoder {
	if opts.Indent <= 0 {
		opts.Indent = 2
	}
	return &Encoder{
		w:    bufio.NewWriter(w),
		opts: opts,
	}
}

// Encode encodes a JSON-compatible value to TOON.
func (e *Encoder) Encode(v interface{}) error {
	return e.encodeValue(v)
}

// Flush flushes any buffered data.
func (e *Encoder) Flush() error {
	return e.w.Flush()
}

func (e *Encoder) encodeValue(v interface{}) error {
	switch val := v.(type) {
	case nil:
		_, err := e.w.WriteString("null")
		return err
	case bool:
		if val {
			_, err := e.w.WriteString("true")
			return err
		}
		_, err := e.w.WriteString("false")
		return err
	case float64:
		return e.writeNumber(val)
	case string:
		return e.writeString(val)
	case []interface{}:
		return e.writeArray(val)
	case map[string]interface{}:
		return e.writeObject(val)
	default:
		// Try reflection for other types
		return e.encodeReflect(v)
	}
}

func (e *Encoder) encodeReflect(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	return e.encodeValue(decoded)
}

func (e *Encoder) indent() {
	indent := strings.Repeat(" ", e.depth*e.opts.Indent)
	_, err := e.w.WriteString(indent)
	if err != nil {
		return
	}
}

func (e *Encoder) writeNumber(n float64) error {
	// Normalize number per TOON spec
	s := normalizeNumber(n)
	_, err := e.w.WriteString(s)
	return err
}

func normalizeNumber(n float64) string {
	// Handle -0 as 0
	if n == 0 {
		return "0"
	}

	// Check if it's an integer
	if n == float64(int64(n)) {
		return strconv.FormatFloat(n, 'f', 0, 64)
	}

	// Format with up to 15 significant digits, remove trailing zeros
	s := strconv.FormatFloat(n, 'f', 15, 64)

	// Remove trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	return s
}

func (e *Encoder) writeString(s string) error {
	if needsQuoting(s, e.opts.Delimiter) {
		return e.writeQuotedString(s)
	}
	_, err := e.w.WriteString(s)
	return err
}

func (e *Encoder) writeQuotedString(s string) error {
	_, err := e.w.WriteString(`"`)
	if err != nil {
		return err
	}
	// Escape special characters
	for _, r := range s {
		switch r {
		case '\\':
			_, err = e.w.WriteString(`\\`)
		case '"':
			_, err = e.w.WriteString(`\"`)
		case '\n':
			_, err = e.w.WriteString(`\n`)
		case '\r':
			_, err = e.w.WriteString(`\r`)
		case '\t':
			_, err = e.w.WriteString(`\t`)
		default:
			_, err = e.w.WriteRune(r)
		}
		if err != nil {
			return err
		}
	}
	_, err = e.w.WriteString(`"`)
	return err
}

func needsQuoting(s string, delimiter rune) bool {
	if s == "" {
		return true
	}

	// Check for leading/trailing whitespace
	trimmed := strings.TrimSpace(s)
	if trimmed != s {
		return true
	}

	// Check for true/false/null
	lower := strings.ToLower(s)
	if lower == "true" || lower == "false" || lower == "null" {
		return true
	}

	// Check if numeric-like
	if isNumericLike(s) {
		return true
	}

	// Check for forbidden characters
	for _, r := range s {
		if r == ':' || r == '"' || r == '\\' ||
			r == '[' || r == ']' || r == '{' || r == '}' ||
			r == delimiter {
			return true
		}
		if r < 0x20 || r == 0x7F {
			return true
		}
	}

	// Check for equals true/false/null
	if s == "-" || strings.HasPrefix(s, "-") {
		return true
	}

	return false
}

func isNumericLike(s string) bool {
	if len(s) == 0 {
		return false
	}

	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
	}

	if start >= len(s) {
		return false
	}

	hasDigit := false
	hasDot := false
	hasExp := false

	for i := start; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
			// Check for leading zero (not allowed unless it's just "0")
			if i == start && c == '0' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
				return true // leading zero
			}
		} else if c == '.' && !hasDot && !hasExp {
			hasDot = true
		} else if (c == 'e' || c == 'E') && !hasExp {
			hasExp = true
		} else {
			return false
		}
	}

	return hasDigit
}

func (e *Encoder) writeObject(m map[string]interface{}) error {
	// Check if we can flatten
	// FlattenDepth of 0 means unlimited
	if e.opts.KeyFolding == "safe" && (e.opts.FlattenDepth == 0 || e.depth < e.opts.FlattenDepth) {
		if folded, ok := e.tryFold(m); ok {
			_, err := e.w.WriteString(folded)
			return err
		}
	}

	first := true
	e.firstField = true
	for k, v := range m {
		if !first {
			if err := e.newLine(); err != nil {
				return err
			}
		}
		first = false

		e.indent()
		_, err := e.w.WriteString(k)
		if err != nil {
			return err
		}

		switch val := v.(type) {
		case nil:
			_, err = e.w.WriteString(": null")
			if err != nil {
				return err
			}
		case bool:
			if val {
				_, err = e.w.WriteString(": true")
			} else {
				_, err = e.w.WriteString(": false")
			}
			if err != nil {
				return err
			}
		case float64:
			_, err = e.w.WriteString(": ")
			if err != nil {
				return err
			}
			if err := e.writeNumber(val); err != nil {
				return err
			}
		case string:
			_, err = e.w.WriteString(": ")
			if err != nil {
				return err
			}
			if err := e.writeString(val); err != nil {
				return err
			}
		case []interface{}:
			if len(val) == 0 {
				_, err = e.w.WriteString(":")
				if err != nil {
					return err
				}
				continue
			}
			// Check if tabular
			if e.isTabular(val) {
				_, err = e.w.WriteString(e.formatTabularHeader(val))
				if err != nil {
					return err
				}
				if err := e.writeTabularRows(val); err != nil {
					return err
				}
				continue
			}
			_, err = e.w.WriteString(e.formatArrayHeader(val))
			if err != nil {
				return err
			}
			if err := e.writeArrayItems(val); err != nil {
				return err
			}
		case map[string]interface{}:
			_, err = e.w.WriteString(":")
			if err != nil {
				return err
			}
			if err := e.newLine(); err != nil {
				return err
			}
			e.depth++
			if err := e.writeObject(val); err != nil {
				return err
			}
			e.depth--
		default:
			_, err = e.w.WriteString(": ")
			if err != nil {
				return err
			}
			if err := e.writeString(formatValue(val)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Encoder) tryFold(m map[string]interface{}) (string, bool) {
	if len(m) != 1 {
		return "", false
	}

	for k, v := range m {
		if !isValidFoldKey(k) {
			return "", false
		}

		nested, ok := v.(map[string]interface{})
		if !ok {
			return "", false
		}

		if len(nested) == 0 {
			return k + ":", true
		}

		inner := nested
		parts := []string{k}

		for len(inner) == 1 {
			for key, val := range inner {
				if !isValidFoldKey(key) {
					return "", false
				}
				parts = append(parts, key)
				inner, ok = val.(map[string]interface{})
				if !ok {
					// Can't fold further, but we have parts
					if len(parts) > 1 {
						return strings.Join(parts, "."), true
					}
					return "", false
				}
				if len(inner) == 0 {
					return strings.Join(parts, "."), true
				}
			}
		}

		if len(parts) > 1 {
			return strings.Join(parts, "."), true
		}
	}

	return "", false
}

func isValidFoldKey(k string) bool {
	for _, r := range k {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return normalizeNumber(val)
	case string:
		return val
	default:
		return ""
	}
}

func (e *Encoder) isTabular(arr []interface{}) bool {
	if len(arr) == 0 {
		return false
	}

	first, ok := arr[0].(map[string]interface{})
	if !ok || len(first) == 0 {
		return false
	}

	firstKeys := make([]string, 0, len(first))
	for k := range first {
		firstKeys = append(firstKeys, k)
	}

	// Check if all elements have the same keys and are all primitives
	for i := 1; i < len(arr); i++ {
		elem, ok := arr[i].(map[string]interface{})
		if !ok || len(elem) != len(first) {
			return false
		}
		for _, k := range firstKeys {
			if _, exists := elem[k]; !exists {
				return false
			}
			// Check if value is primitive
			if !isPrimitive(elem[k]) {
				return false
			}
		}
	}

	return true
}

func isPrimitive(v interface{}) bool {
	switch v.(type) {
	case nil, bool, float64, string:
		return true
	}
	return false
}

func (e *Encoder) formatTabularHeader(arr []interface{}) string {
	if len(arr) == 0 {
		return ":"
	}

	first, _ := arr[0].(map[string]interface{})
	keys := make([]string, 0, len(first))
	for k := range first {
		keys = append(keys, k)
	}

	header := "{" + strings.Join(keys, ",") + "}:"
	return header
}

func (e *Encoder) writeTabularRows(arr []interface{}) error {
	if len(arr) == 0 {
		return nil
	}

	first, _ := arr[0].(map[string]interface{})
	keys := make([]string, 0, len(first))
	for k := range first {
		keys = append(keys, k)
	}

	delim := string(e.opts.Delimiter)

	for _, elem := range arr {
		m, _ := elem.(map[string]interface{})
		if err := e.newLine(); err != nil {
			return err
		}
		e.indent()

		values := make([]string, len(keys))
		for i, k := range keys {
			v := m[k]
			values[i] = e.formatPrimitive(v, delim)
		}
		_, err := e.w.WriteString(strings.Join(values, delim))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) formatArrayHeader(arr []interface{}) string {
	return "[" + strconv.Itoa(len(arr)) + "]:"
}

func (e *Encoder) writeArray(arr []interface{}) error {
	if len(arr) == 0 {
		_, err := e.w.WriteString(":")
		return err
	}

	// Check if tabular
	if e.isTabular(arr) {
		_, err := e.w.WriteString(e.formatTabularHeader(arr))
		if err != nil {
			return err
		}
		return e.writeTabularRows(arr)
	}

	// Check if all primitives
	if e.allPrimitives(arr) && len(arr) > 0 {
		_, err := e.w.WriteString(e.formatArrayHeader(arr))
		if err != nil {
			return err
		}
		e.indent()
		_, err = e.w.WriteString(" ")
		if err != nil {
			return err
		}
		for i, v := range arr {
			if i > 0 {
				_, err = e.w.WriteString(",")
				if err != nil {
					return err
				}
			}
			if err := e.writePrimitive(v); err != nil {
				return err
			}
		}
		return nil
	}

	// Mixed or nested array
	_, err := e.w.WriteString(":")
	if err != nil {
		return err
	}
	for _, v := range arr {
		if err := e.newLine(); err != nil {
			return err
		}
		e.indent()
		_, err := e.w.WriteString("- ")
		if err != nil {
			return err
		}

		switch val := v.(type) {
		case []interface{}:
			if len(val) == 0 {
				_, err = e.w.WriteString(":")
				if err != nil {
					return err
				}
			} else if e.isTabular(val) {
				_, err = e.w.WriteString(val[0].(map[string]interface{})[getFirstKey(val)].(string) + e.formatArrayHeader(val))
				if err != nil {
					return err
				}
				if err := e.writeTabularRows(val); err != nil {
					return err
				}
			} else {
				e.depth++
				if err := e.writeArrayItems(val); err != nil {
					return err
				}
				e.depth--
			}
		case map[string]interface{}:
			e.depth++
			if err := e.writeObject(val); err != nil {
				return err
			}
			e.depth--
		default:
			if err := e.writePrimitive(val); err != nil {
				return err
			}
		}
	}

	return nil
}

func getFirstKey(arr []interface{}) string {
	if len(arr) == 0 {
		return ""
	}
	m, ok := arr[0].(map[string]interface{})
	if !ok || len(m) == 0 {
		return ""
	}
	for k := range m {
		return k
	}
	return ""
}

func (e *Encoder) allPrimitives(arr []interface{}) bool {
	for _, v := range arr {
		if !isPrimitive(v) {
			return false
		}
	}
	return true
}

func (e *Encoder) formatPrimitive(v interface{}, delim string) string {
	switch val := v.(type) {
	case nil:
		return ""
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return normalizeNumber(val)
	case string:
		if needsQuoting(val, rune(delim[0])) {
			return `"` + escapeString(val) + `"`
		}
		return val
	default:
		return ""
	}
}

func (e *Encoder) writePrimitive(v interface{}) error {
	switch val := v.(type) {
	case nil:
		_, err := e.w.WriteString("null")
		return err
	case bool:
		if val {
			_, err := e.w.WriteString("true")
			return err
		}
		_, err := e.w.WriteString("false")
		return err
	case float64:
		return e.writeNumber(val)
	case string:
		return e.writeString(val)
	default:
		return errors.New("not a primitive value")
	}
}

func escapeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			result.WriteString(`\\`)
		case '"':
			result.WriteString(`\"`)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (e *Encoder) writeArrayItems(arr []interface{}) error {
	for i, v := range arr {
		if i > 0 {
			if err := e.newLine(); err != nil {
				return err
			}
		}
		if err := e.writePrimitive(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) newLine() error {
	_, err := e.w.WriteString("\n")
	return err
}
