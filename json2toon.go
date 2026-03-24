package json2toon

import (
	"bytes"
	"io"
	"os"
)

// ConverterOption configures the converter behavior.
type ConverterOption func(*ConverterOptions)

// WithIndent sets the indentation level.
func WithIndent(n int) ConverterOption {
	return func(opts *ConverterOptions) {
		opts.Encoder.Indent = n
	}
}

// WithDelimiter sets the table delimiter character.
func WithDelimiter(r rune) ConverterOption {
	return func(opts *ConverterOptions) {
		opts.Encoder.Delimiter = r
	}
}

// WithKeyFolding sets the key folding mode.
func WithKeyFolding(mode string) ConverterOption {
	return func(opts *ConverterOptions) {
		opts.Encoder.KeyFolding = mode
	}
}

// WithStrict enables or disables strict mode.
func WithStrict(strict bool) ConverterOption {
	return func(opts *ConverterOptions) {
		opts.Strict = strict
	}
}

// Convert converts JSON bytes to TOON bytes.
func Convert(jsonBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	converter := NewConverter(&buf)
	if err := converter.ConvertJSON(jsonBytes); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertString converts a JSON string to TOON string.
func ConvertString(jsonStr string) (string, error) {
	result, err := Convert([]byte(jsonStr))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ConvertWithOptions converts JSON bytes to TOON with custom options.
func ConvertWithOptions(jsonBytes []byte, opts ...ConverterOption) ([]byte, error) {
	options := DefaultConverterOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var buf bytes.Buffer
	converter := NewConverterWithOptions(&buf, options)
	if err := converter.ConvertJSON(jsonBytes); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertJSONC converts JSONC (JSON with comments) to TOON bytes.
func ConvertJSONC(jsoncBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	converter := NewConverter(&buf)
	if err := converter.ConvertJSONC(jsoncBytes); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertJSONCString converts a JSONC string to TOON string.
func ConvertJSONCString(jsoncStr string) (string, error) {
	result, err := ConvertJSONC([]byte(jsoncStr))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ConvertJSONCWithOptions converts JSONC to TOON with custom options.
func ConvertJSONCWithOptions(jsoncBytes []byte, opts ...ConverterOption) ([]byte, error) {
	options := DefaultConverterOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var buf bytes.Buffer
	converter := NewConverterWithOptions(&buf, options)
	if err := converter.ConvertJSONC(jsoncBytes); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertFile converts a JSON file to TOON bytes.
func ConvertFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Convert(data)
}

// ConvertFileToWriter converts a JSON file and writes to an io.Writer.
func ConvertFileToWriter(path string, w io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	converter := NewConverter(w)
	if err := converter.ConvertJSON(data); err != nil {
		return err
	}
	return converter.Close()
}

// ConvertFileWithOptions converts a JSON file with custom options.
func ConvertFileWithOptions(path string, opts ...ConverterOption) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ConvertWithOptions(data, opts...)
}
