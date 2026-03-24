package json2toon

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
