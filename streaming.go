package json2toon

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Converter provides streaming JSON/JSONC/JSONL to TOON conversion.
// It writes directly to an io.Writer without buffering the entire output.
type Converter struct {
	encoder *Encoder
	w       io.Writer
	opts    ConverterOptions
}

// NewConverter creates a new Converter that writes TOON to w.
func NewConverter(w io.Writer) *Converter {
	return NewConverterWithOptions(w, DefaultConverterOptions())
}

// NewConverterWithOptions creates a new Converter with custom options.
func NewConverterWithOptions(w io.Writer, opts ConverterOptions) *Converter {
	return &Converter{
		encoder: NewEncoderWithOptions(w, opts.Encoder),
		w:       w,
		opts:    opts,
	}
}

// ConvertJSON converts JSON bytes to TOON using streaming token-based parsing.
// It is memory-efficient for large JSON documents as it doesn't load the
// entire input into memory.
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
		// JSON null value - encode it
		return c.encodePrimitive(nil)
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

// ConvertJSONC converts JSONC (JSON with comments) to TOON.
// It first strips comments (// and /* */) and then converts the result.
func (c *Converter) ConvertJSONC(jsoncBytes []byte) error {
	cleaned, err := StripComments(jsoncBytes)
	if err != nil {
		return err
	}
	return c.ConvertJSON(cleaned)
}

// ConvertJSONL converts JSON Lines (JSONL) to TOON.
// Each line is parsed as a separate JSON value.
// If all lines are objects with the same keys, output is in tabular format.
func (c *Converter) ConvertJSONL(jsonlBytes []byte) error {
	return c.ConvertJSONLFromReader(bytes.NewReader(jsonlBytes))
}

// ConvertJSONLFromReader converts JSON Lines from an io.Reader to TOON.
// This is more memory-efficient than ConvertJSONL for large inputs.
func (c *Converter) ConvertJSONLFromReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	tmpFile, err := os.CreateTemp("", "json2toon-jsonl-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer func() { _ = tmpFile.Close() }()

	isTabular := true
	var expectedKeys []string
	lineCount := 0

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		if _, err := tmpFile.Write(line); err != nil {
			return err
		}
		if _, err := tmpFile.Write([]byte("\n")); err != nil {
			return err
		}
		lineCount++

		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			isTabular = false
			continue
		}

		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}

		if expectedKeys == nil {
			expectedKeys = keys
		} else if !sameKeys(expectedKeys, keys) {
			isTabular = false
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if lineCount == 0 {
		return nil
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	secondPass := bufio.NewScanner(tmpFile)

	delim := string(c.encoder.opts.Delimiter)

	if isTabular && len(expectedKeys) > 0 {
		header := fmt.Sprintf("[%d]{%s}:\n", lineCount, strings.Join(expectedKeys, delim))
		if _, err := c.w.Write([]byte(header)); err != nil {
			return err
		}

		for secondPass.Scan() {
			line := bytes.TrimSpace(secondPass.Bytes())
			if len(line) == 0 {
				continue
			}

			var obj map[string]interface{}
			if err := json.Unmarshal(line, &obj); err != nil {
				return err
			}

			values := make([]string, len(expectedKeys))
			for i, k := range expectedKeys {
				values[i] = c.formatPrimitiveValue(obj[k], delim)
			}
			if _, err := c.w.Write([]byte("  " + strings.Join(values, delim))); err != nil {
				return err
			}
			if _, err := c.w.Write([]byte("\n")); err != nil {
				return err
			}
		}

		return secondPass.Err()
	} else {
		first := true
		for secondPass.Scan() {
			line := bytes.TrimSpace(secondPass.Bytes())
			if len(line) == 0 {
				continue
			}

			if !first {
				if _, err := c.w.Write([]byte("---\n")); err != nil {
					return err
				}
			}
			first = false

			conv := NewConverterWithOptions(c.w, c.opts)
			if err := conv.ConvertJSON(line); err != nil {
				return err
			}
		}

		return secondPass.Err()
	}
}

// ConvertJSONLStream converts JSON Lines to TOON in streaming mode.
// Each line is processed and written immediately, separated by "---".
// This is the most memory-efficient option for large JSONL files.
func (c *Converter) ConvertJSONLStream(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	first := true

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		// Validate it's valid JSON
		if !json.Valid(line) {
			return fmt.Errorf("invalid JSON on line: %s", string(line))
		}

		if !first {
			// Add separator between JSON values
			if _, err := c.w.Write([]byte("\n---\n")); err != nil {
				return err
			}
		}
		first = false

		if err := c.ConvertJSON(line); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// ConvertAuto auto-detects and converts the input format (JSON/JSONC/JSONL) to TOON.
// It examines the input to determine the format:
//   - JSONC: contains // or /* */ comments
//   - JSONL: multiple lines of valid JSON objects
//   - JSON: standard JSON format (default)
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

// Close flushes any buffered data and releases resources.
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
		firstObj, _ := items[0].(map[string]interface{})
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
			keys := makeSortedKeys(val)
			for i, k := range keys {
				v := val[k]
				if i > 0 {
					if _, err := c.w.Write([]byte("\n")); err != nil {
						return err
					}
					indent2 := strings.Repeat(" ", c.encoder.depth*2)
					if _, err := c.w.Write([]byte(indent2)); err != nil {
						return err
					}
				}
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
	firstObj, _ := items[0].(map[string]interface{})
	keys := makeSortedKeys(firstObj)

	for _, item := range items {
		obj, _ := item.(map[string]interface{})
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
	return formatPrimitiveToString(v, rune(delim[0]))
}

func (c *Converter) encodePrimitive(v interface{}) error {
	return writePrimitiveToWriter(c.w, v, c.encoder.opts.Delimiter)
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
