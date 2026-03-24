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
