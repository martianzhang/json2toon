package json2toon

import (
	"bytes"
	"errors"
)

// StripComments removes JSONC-style comments from JSONC content.
// Supports:
//   - Line comments: // comment to end of line
//   - Block comments: /* multi-line comment */
//   - Comments inside strings (not stripped)
func StripComments(jsonc []byte) ([]byte, error) {
	var result bytes.Buffer
	i := 0

	for i < len(jsonc) {
		// Check for string start
		if jsonc[i] == '"' {
			if err := copyString(&result, jsonc, &i); err != nil {
				return nil, err
			}
			continue
		}

		// Check for line comment
		if i+1 < len(jsonc) && jsonc[i] == '/' && jsonc[i+1] == '/' {
			i += 2
			// Skip to end of line
			for i < len(jsonc) && jsonc[i] != '\n' {
				i++
			}
			// Keep the newline
			continue
		}

		// Check for block comment
		if i+1 < len(jsonc) && jsonc[i] == '/' && jsonc[i+1] == '*' {
			i += 2
			// Skip until end of block comment
			for i+1 < len(jsonc) {
				if jsonc[i] == '*' && jsonc[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			if i >= len(jsonc) {
				return nil, errors.New("unterminated block comment")
			}
			continue
		}

		// Regular character
		result.WriteByte(jsonc[i])
		i++
	}

	return result.Bytes(), nil
}

func copyString(result *bytes.Buffer, data []byte, i *int) error {
	result.WriteByte(data[*i])
	*i++

	for *i < len(data) {
		c := data[*i]
		if c == '\\' && *i+1 < len(data) {
			// Escape sequence
			result.WriteByte(c)
			*i++
			result.WriteByte(data[*i])
			*i++
		} else if c == '"' {
			result.WriteByte(c)
			*i++
			return nil
		} else {
			result.WriteByte(c)
			*i++
		}
	}

	return errors.New("unterminated string")
}
