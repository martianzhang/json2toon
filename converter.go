package json2toon

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// mayContainComments checks if data may contain JSONC comments (// or /*).
func mayContainComments(data []byte) bool {
	inString := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' && (i == 0 || data[i-1] != '\\') {
			inString = !inString
			continue
		}
		if !inString && i+1 < len(data) {
			if data[i] == '/' && (data[i+1] == '/' || data[i+1] == '*') {
				return true
			}
		}
	}
	return false
}

// isJSONL checks if data is JSON Lines format.
// Returns true if:
//  1. Multiple lines exist
//  2. Most non-empty lines are complete JSON objects/arrays (start and end properly)
//  3. Each line parses as valid JSON when trimmed
func isJSONL(data []byte) bool {
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) < 2 {
		return false
	}

	validLines := 0
	totalLines := 0

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		totalLines++

		// Check if line looks like a complete JSON object or array
		if (trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}') ||
			(trimmed[0] == '[' && trimmed[len(trimmed)-1] == ']') {
			// Verify it's valid JSON
			if json.Valid(trimmed) {
				validLines++
			}
		}
	}

	// If we have 2+ lines and most are valid complete JSON objects, it's JSONL
	return totalLines >= 2 && validLines >= 2 && float64(validLines)/float64(totalLines) > 0.5
}

// DetectFormat detects the input format: JSON, JSONC, or JSONL.
func DetectFormat(data []byte) string {
	if mayContainComments(data) {
		return "jsonc"
	}
	if isJSONL(data) {
		return "jsonl"
	}
	return "json"
}

type Converter struct {
	encoder *Encoder
	w       io.Writer
	opts    ConverterOptions
}

func NewConverter(w io.Writer) *Converter {
	return NewConverterWithOptions(w, DefaultConverterOptions())
}

func NewConverterWithOptions(w io.Writer, opts ConverterOptions) *Converter {
	return &Converter{
		encoder: NewEncoderWithOptions(w, opts.Encoder),
		w:       w,
		opts:    opts,
	}
}

func (c *Converter) ConvertJSON(jsonBytes []byte) error {
	reader := bytes.NewReader(jsonBytes)
	decoder := json.NewDecoder(reader)

	token, err := decoder.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	if token == nil {
		return nil
	}

	delim, ok := token.(json.Delim)
	if !ok {
		return c.encodePrimitive(token)
	}

	switch delim {
	case '{':
		if err := c.decodeObject(decoder); err != nil {
			return err
		}
	case '[':
		items, err := c.collectArrayItems(decoder)
		if err != nil {
			return err
		}
		if err := c.writeArrayDirect(items); err != nil {
			return err
		}
	}

	return c.encoder.Flush()
}

func (c *Converter) ConvertJSONC(jsoncBytes []byte) error {
	cleaned, err := StripComments(jsoncBytes)
	if err != nil {
		return err
	}
	return c.ConvertJSON(cleaned)
}

// ConvertJSONL converts JSON Lines (JSONL) to TOON.
// Each line is parsed as a separate JSON object.
// If all objects have the same keys, output is in tabular format.
func (c *Converter) ConvertJSONL(jsonlBytes []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(jsonlBytes))

	type lineData struct {
		raw    []byte
		keys   []string
		values map[string]interface{}
	}
	var lines []lineData
	isTabular := true
	var expectedKeys []string

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			isTabular = false
		} else {
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}

			if expectedKeys == nil {
				expectedKeys = keys
			} else if !sameKeys(expectedKeys, keys) {
				isTabular = false
			}

			lines = append(lines, lineData{
				raw:    line,
				keys:   keys,
				values: obj,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	delim := string(c.encoder.opts.Delimiter)

	if isTabular && len(lines) > 0 {
		header := fmt.Sprintf("[%d]{%s}:\n", len(lines), strings.Join(expectedKeys, delim))
		if _, err := c.w.Write([]byte(header)); err != nil {
			return err
		}

		for _, line := range lines {
			values := make([]string, len(expectedKeys))
			for i, k := range expectedKeys {
				values[i] = c.formatPrimitiveValue(line.values[k], delim)
			}
			if _, err := c.w.Write([]byte("  " + strings.Join(values, delim))); err != nil {
				return err
			}
			if _, err := c.w.Write([]byte("\n")); err != nil {
				return err
			}
		}
	} else {
		for i, line := range lines {
			if i > 0 {
				if _, err := c.w.Write([]byte("---\n")); err != nil {
					return err
				}
			}
			conv := NewConverterWithOptions(c.w, c.opts)
			if err := conv.ConvertJSON(line.raw); err != nil {
				return err
			}
		}
	}

	return nil
}

// ConvertAuto auto-detects and converts the input format (JSON/JSONC/JSONL) to TOON.
func (c *Converter) ConvertAuto(data []byte) error {
	format := DetectFormat(data)
	switch format {
	case "jsonc":
		return c.ConvertJSONC(data)
	case "jsonl":
		return c.ConvertJSONL(data)
	default:
		return c.ConvertJSON(data)
	}
}

func (c *Converter) Close() error {
	return c.encoder.Flush()
}

func (c *Converter) decodeObject(decoder *json.Decoder) error {
	first := true
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}

		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", keyToken)
		}

		valueToken, err := decoder.Token()
		if err != nil {
			return err
		}

		if !first {
			if _, err := c.w.Write([]byte("\n")); err != nil {
				return err
			}
		}
		first = false

		indent := strings.Repeat(" ", c.encoder.depth*2)
		if _, err := c.w.Write([]byte(indent)); err != nil {
			return err
		}

		if _, err := c.w.Write([]byte(key)); err != nil {
			return err
		}

		_, err = c.writeValueStream(valueToken, decoder)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) writeValueStream(token interface{}, decoder *json.Decoder) (bool, error) {
	delim, ok := token.(json.Delim)
	if ok {
		switch delim {
		case '{':
			if _, err := c.w.Write([]byte(":\n")); err != nil {
				return false, err
			}
			c.encoder.depth++
			if err := c.decodeObject(decoder); err != nil {
				return false, err
			}
			c.encoder.depth--
			// Consume the closing } that decodeObject exited before
			_, err := decoder.Token()
			if err != nil {
				return true, err
			}
			return true, nil
		case '[':
			if _, err := c.w.Write([]byte(":")); err != nil {
				return false, err
			}
			items, err := c.collectArrayItems(decoder)
			if err != nil {
				return false, err
			}
			// Consume the closing ']'
			if _, err := decoder.Token(); err != nil {
				return false, err
			}
			return false, c.writeArrayDirect(items)
		}
		return false, nil
	}

	if _, err := c.w.Write([]byte(": ")); err != nil {
		return false, err
	}
	return false, c.encodePrimitive(token)
}

func (c *Converter) collectArrayItems(decoder *json.Decoder) ([]interface{}, error) {
	items := make([]interface{}, 0)

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		if delim, ok := token.(json.Delim); ok {
			switch delim {
			case '{':
				obj, err := c.decodeObjectToMap(decoder)
				if err != nil {
					return nil, err
				}
				items = append(items, obj)
			case '[':
				subArr, err := c.collectArrayItems(decoder)
				if err != nil {
					return nil, err
				}
				// Consume the closing ']'
				if _, err := decoder.Token(); err != nil {
					return nil, err
				}
				items = append(items, subArr)
			}
		} else {
			items = append(items, token)
		}
	}

	return items, nil
}

func (c *Converter) decodeObjectToMap(decoder *json.Decoder) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		key, ok := keyToken.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key")
		}

		valueToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		if delim, ok := valueToken.(json.Delim); ok {
			switch delim {
			case '{':
				subObj, err := c.decodeObjectToMap(decoder)
				if err != nil {
					return nil, err
				}
				m[key] = subObj
			case '[':
				subArr, err := c.collectArrayItems(decoder)
				if err != nil {
					return nil, err
				}
				// Consume the closing ']'
				if _, err := decoder.Token(); err != nil {
					return nil, err
				}
				m[key] = subArr
			default:
				m[key] = valueToken
			}
		} else {
			m[key] = valueToken
		}
	}

	// Consume the closing '}'
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Converter) writeArrayDirect(items []interface{}) error {
	if len(items) == 0 {
		return nil
	}

	if c.isTabularArray(items) {
		header := "[" + strconv.Itoa(len(items)) + "]{"
		firstObj := items[0].(map[string]interface{})
		keys := makeSortedKeys(firstObj)
		header += strings.Join(keys, ",") + "}:"
		if _, err := c.w.Write([]byte(header)); err != nil {
			return err
		}
		return c.writeTabularRows(items)
	}

	if c.isAllPrimitives(items) {
		header := "[" + strconv.Itoa(len(items)) + "]: "
		if _, err := c.w.Write([]byte(header)); err != nil {
			return err
		}
		for i, item := range items {
			if i > 0 {
				if _, err := c.w.Write([]byte(",")); err != nil {
					return err
				}
			}
			if err := c.encodePrimitive(item); err != nil {
				return err
			}
		}
		return nil
	}

	return c.writeArrayListStream(items)
}

func (c *Converter) isAllPrimitives(items []interface{}) bool {
	for _, item := range items {
		switch item.(type) {
		case nil, bool, float64, string:
			continue
		default:
			return false
		}
	}
	return true
}

func (c *Converter) isTabularArray(items []interface{}) bool {
	if len(items) == 0 {
		return false
	}

	first, ok := items[0].(map[string]interface{})
	if !ok || len(first) == 0 {
		return false
	}

	firstKeys := makeSortedKeys(first)

	for i := 1; i < len(items); i++ {
		elem, ok := items[i].(map[string]interface{})
		if !ok || len(elem) != len(first) {
			return false
		}
		elemKeys := makeSortedKeys(elem)
		for j, k := range firstKeys {
			if elemKeys[j] != k {
				return false
			}
		}
		for _, v := range elem {
			if !c.isPrimitive(v) {
				return false
			}
		}
	}

	return true
}

func (c *Converter) isPrimitive(v interface{}) bool {
	switch v.(type) {
	case nil, bool, float64, string:
		return true
	}
	return false
}

// writeObjectDirect writes a nested object in list format (key: value per line).
func (c *Converter) writeObjectDirect(m map[string]interface{}) error {
	c.encoder.depth++
	first := true
	for k, v := range m {
		if !first {
			if _, err := c.w.Write([]byte("\n")); err != nil {
				return err
			}
			indent := strings.Repeat(" ", c.encoder.depth*2)
			if _, err := c.w.Write([]byte(indent)); err != nil {
				return err
			}
		}
		first = false
		if _, err := c.w.Write([]byte(k + ": ")); err != nil {
			return err
		}
		switch v2 := v.(type) {
		case []interface{}:
			if err := c.writeArrayDirect(v2); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := c.writeObjectDirect(v2); err != nil {
				return err
			}
		default:
			if err := c.encodePrimitive(v); err != nil {
				return err
			}
		}
	}
	c.encoder.depth--
	return nil
}

func (c *Converter) writeArrayListStream(items []interface{}) error {
	for _, item := range items {
		if _, err := c.w.Write([]byte("\n")); err != nil {
			return err
		}
		indent := strings.Repeat(" ", c.encoder.depth*2)
		if _, err := c.w.Write([]byte(indent)); err != nil {
			return err
		}
		if _, err := c.w.Write([]byte("- ")); err != nil {
			return err
		}

		switch val := item.(type) {
		case map[string]interface{}:
			c.encoder.depth++
			first := true
			for k, v := range val {
				if !first {
					if _, err := c.w.Write([]byte("\n")); err != nil {
						return err
					}
					indent2 := strings.Repeat(" ", c.encoder.depth*2)
					if _, err := c.w.Write([]byte(indent2)); err != nil {
						return err
					}
				}
				first = false
				if _, err := c.w.Write([]byte(k + ": ")); err != nil {
					return err
				}
				switch v2 := v.(type) {
				case []interface{}:
					if err := c.writeArrayDirect(v2); err != nil {
						return err
					}
				case map[string]interface{}:
					if err := c.writeObjectDirect(v2); err != nil {
						return err
					}
				default:
					if err := c.encodePrimitive(v); err != nil {
						return err
					}
				}
			}
			c.encoder.depth--
		case []interface{}:
			c.encoder.depth++
			if err := c.writeArrayDirect(val); err != nil {
				return err
			}
			c.encoder.depth--
		default:
			if err := c.encodePrimitive(val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Converter) writeTabularRows(items []interface{}) error {
	if len(items) == 0 {
		return nil
	}

	delim := string(c.encoder.opts.Delimiter)
	firstObj := items[0].(map[string]interface{})
	keys := makeSortedKeys(firstObj)

	for _, item := range items {
		obj := item.(map[string]interface{})
		if _, err := c.w.Write([]byte("\n")); err != nil {
			return err
		}
		indent := strings.Repeat(" ", c.encoder.depth*2)
		if _, err := c.w.Write([]byte(indent)); err != nil {
			return err
		}

		values := make([]string, len(keys))
		for i, k := range keys {
			values[i] = c.formatPrimitiveValue(obj[k], delim)
		}
		if _, err := c.w.Write([]byte(strings.Join(values, delim))); err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) formatPrimitiveValue(v interface{}, delim string) string {
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

func (c *Converter) encodePrimitive(v interface{}) error {
	switch val := v.(type) {
	case nil:
		_, err := c.w.Write([]byte("null"))
		return err
	case bool:
		if val {
			_, err := c.w.Write([]byte("true"))
			return err
		}
		_, err := c.w.Write([]byte("false"))
		return err
	case float64:
		_, err := c.w.Write([]byte(normalizeNumber(val)))
		return err
	case string:
		if needsQuoting(val, c.encoder.opts.Delimiter) {
			_, err := c.w.Write([]byte(`"` + escapeString(val) + `"`))
			return err
		}
		_, err := c.w.Write([]byte(val))
		return err
	default:
		return fmt.Errorf("unknown primitive type: %T", v)
	}
}

func makeSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// sameKeys checks if two key slices contain the same elements.
func sameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	slices.Sort(a)
	slices.Sort(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
