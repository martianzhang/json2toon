package json2toon

// EncoderOptions configures the TOON encoder behavior.
type EncoderOptions struct {
	// Indent is the number of spaces per indentation level.
	// Defaults to 2.
	Indent int

	// Delimiter is the character used to separate values in tabular arrays.
	// Options: ',', '\t', '|'
	// Defaults to ','.
	Delimiter rune

	// KeyFolding enables key folding: "off" or "safe".
	// "safe" collapses nested single-key objects into dotted paths.
	// Defaults to "off".
	KeyFolding string

	// FlattenDepth limits key folding depth. 0 means unlimited.
	// Defaults to 0 (unlimited).
	FlattenDepth int
}

// DefaultEncoderOptions returns the default encoder options.
func DefaultEncoderOptions() EncoderOptions {
	return EncoderOptions{
		Indent:     2,
		Delimiter:  ',',
		KeyFolding: "off",
	}
}

// ConverterOptions configures the JSON to TOON converter.
type ConverterOptions struct {
	// Encoder configures the TOON encoder.
	Encoder EncoderOptions

	// Strict enables strict TOON spec validation.
	// When true, enforces strict compliance with TOON specification.
	// Defaults to true.
	Strict bool
}

// DefaultConverterOptions returns the default converter options.
func DefaultConverterOptions() ConverterOptions {
	return ConverterOptions{
		Encoder: DefaultEncoderOptions(),
		Strict:  true,
	}
}

// DecodeOption configures the decoder behavior.
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
