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

// ConvertJSONL converts JSON Lines (JSONL) to TOON bytes.
func ConvertJSONL(jsonlBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	converter := NewConverter(&buf)
	if err := converter.ConvertJSONL(jsonlBytes); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertJSONLStream converts JSON Lines to TOON bytes using streaming mode.
// Each line is processed and written immediately, separated by "---".
// This is more memory-efficient for large JSONL files.
func ConvertJSONLStream(jsonlBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	converter := NewConverter(&buf)
	if err := converter.ConvertJSONLStream(bytes.NewReader(jsonlBytes)); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertJSONLStreamString converts JSON Lines string to TOON using streaming mode.
func ConvertJSONLStreamString(jsonlStr string) (string, error) {
	result, err := ConvertJSONLStream([]byte(jsonlStr))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ConvertJSONLString converts a JSONL string to TOON string.
func ConvertJSONLString(jsonlStr string) (string, error) {
	result, err := ConvertJSONL([]byte(jsonlStr))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ConvertJSONLWithOptions converts JSONL to TOON with custom options.
func ConvertJSONLWithOptions(jsonlBytes []byte, opts ...ConverterOption) ([]byte, error) {
	options := DefaultConverterOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var buf bytes.Buffer
	converter := NewConverterWithOptions(&buf, options)
	if err := converter.ConvertJSONL(jsonlBytes); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertJSONLStreamWithOptions converts JSONL to TOON with streaming mode and custom options.
func ConvertJSONLStreamWithOptions(jsonlBytes []byte, opts ...ConverterOption) ([]byte, error) {
	options := DefaultConverterOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var buf bytes.Buffer
	converter := NewConverterWithOptions(&buf, options)
	if err := converter.ConvertJSONLStream(bytes.NewReader(jsonlBytes)); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertAuto auto-detects and converts input format (JSON/JSONC/JSONL) to TOON.
func ConvertAuto(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	converter := NewConverter(&buf)
	if err := converter.ConvertAuto(data); err != nil {
		return nil, err
	}
	if err := converter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ConvertAutoString converts input string to TOON with auto format detection.
func ConvertAutoString(s string) (string, error) {
	result, err := ConvertAuto([]byte(s))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ConvertAutoWithOptions converts to TOON with auto format detection and custom options.
func ConvertAutoWithOptions(data []byte, opts ...ConverterOption) ([]byte, error) {
	options := DefaultConverterOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var buf bytes.Buffer
	converter := NewConverterWithOptions(&buf, options)
	if err := converter.ConvertAuto(data); err != nil {
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

// ConvertAutoFile converts a file with auto-detected format to TOON bytes.
func ConvertAutoFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ConvertAuto(data)
}

// --- TOON to JSON (Reverse Conversion) ---

// ConvertToJSON converts TOON bytes to JSON bytes.
func ConvertToJSON(toonBytes []byte) ([]byte, error) {
	return DecodeToJSON(toonBytes)
}

// ConvertToJSONString converts TOON string to JSON string.
func ConvertToJSONString(toonStr string) (string, error) {
	return DecodeToJSONString(toonStr)
}

// ConvertToJSONFromReader converts TOON from an io.Reader to JSON bytes.
func ConvertToJSONFromReader(r io.Reader) ([]byte, error) {
	return DecodeFromReader(r)
}

// ConvertToJSONFile converts a TOON file to JSON bytes.
func ConvertToJSONFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecodeToJSON(data)
}

// ConvertToJSONFileString converts a TOON file to JSON string.
func ConvertToJSONFileString(path string) (string, error) {
	result, err := ConvertToJSONFile(path)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// --- Decoder Options ---

// DecoderOption configures the decoder behavior.
type DecodeOption func(*DecodeOptions)

// DecodeOptions holds decoder configuration.
type DecodeOptions struct {
	Strict      bool
	ExpandPaths bool
}

// DefaultDecodeOptions returns the default decoder options.
func DefaultDecodeOptions() DecodeOptions {
	return DecodeOptions{
		Strict:      true,
		ExpandPaths: false,
	}
}

// WithDecodeStrict enables or disables strict validation when decoding.
func WithDecodeStrict(strict bool) DecodeOption {
	return func(opts *DecodeOptions) {
		opts.Strict = strict
	}
}

// WithExpandPaths enables path expansion for reconstructing nested structures from folded keys.
func WithExpandPaths(expand bool) DecodeOption {
	return func(opts *DecodeOptions) {
		opts.ExpandPaths = expand
	}
}

// ConvertToJSONWithOptions converts TOON to JSON with custom options.
func ConvertToJSONWithOptions(toonBytes []byte, opts ...DecodeOption) ([]byte, error) {
	options := DefaultDecodeOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return DecodeToJSONWithOptions(toonBytes, options)
}
