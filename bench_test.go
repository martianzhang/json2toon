package json2toon

import (
	"bytes"
	"testing"
)

func BenchmarkConvertSimple(b *testing.B) {
	json := []byte(`{"id": 123, "name": "Ada", "active": true}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Convert(json)
	}
}

func BenchmarkConvertNested(b *testing.B) {
	json := []byte(`{"user": {"id": 123, "name": "Ada", "profile": {"email": "ada@example.com"}}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Convert(json)
	}
}

func BenchmarkConvertArray(b *testing.B) {
	// Generate array with 100 items
	var buf bytes.Buffer
	buf.WriteString(`{"items": [`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			buf.WriteString(`,`)
		}
		buf.WriteString(`"item`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`"`)
	}
	buf.WriteString(`]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Convert(buf.Bytes())
	}
}

func BenchmarkConvertTabular(b *testing.B) {
	// Generate tabular array with 100 rows
	var buf bytes.Buffer
	buf.WriteString(`{"users": [`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			buf.WriteString(`,`)
		}
		buf.WriteString(`{"id":`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`,`)
		buf.WriteString(`"name":"User`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`"}`)
	}
	buf.WriteString(`]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Convert(buf.Bytes())
	}
}

func BenchmarkConvertLargeFile(b *testing.B) {
	// Generate large nested structure
	var buf bytes.Buffer
	buf.WriteString(`{"data":{`)
	for i := 0; i < 50; i++ {
		if i > 0 {
			buf.WriteString(`,`)
		}
		buf.WriteString(`"key`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`":{"items":[`)
		for j := 0; j < 10; j++ {
			if j > 0 {
				buf.WriteString(`,`)
			}
			buf.WriteString(`{"id":`)
			buf.WriteString(string(rune('0' + (j % 10))))
			buf.WriteString(`}`)
		}
		buf.WriteString(`]}`)
	}
	buf.WriteString(`}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Convert(buf.Bytes())
	}
}

func BenchmarkConvertJSONC(b *testing.B) {
	jsonc := []byte(`{
  // This is a comment
  "name": "Test",
  /* Multi-line
     comment */
  "value": 42,
  "items": ["a", "b", "c"]
}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertJSONC(jsonc)
	}
}

func BenchmarkStreamingVsBatch(b *testing.B) {
	// Generate moderate-sized JSON
	var buf bytes.Buffer
	buf.WriteString(`{"items": [`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			buf.WriteString(`,`)
		}
		buf.WriteString(`{"id":`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`,`)
		buf.WriteString(`"name":"Item`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`"}`)
	}
	buf.WriteString(`]}`)

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Convert(buf.Bytes())
		}
	})

	b.Run("Streaming", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var out bytes.Buffer
			converter := NewConverter(&out)
			_ = converter.ConvertJSON(buf.Bytes())
			_ = converter.Close()
		}
	})
}

func BenchmarkEncoderOptions(b *testing.B) {
	json := []byte(`{"name": "Test", "value": 42}`)

	b.Run("Default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Convert(json)
		}
	})

	b.Run("WithIndent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ConvertWithOptions(json, WithIndent(4))
		}
	})

	b.Run("WithDelimiter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ConvertWithOptions(json, WithDelimiter('|'))
		}
	})

	b.Run("WithAllOptions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ConvertWithOptions(json, WithIndent(4), WithDelimiter('|'))
		}
	})
}

func BenchmarkConvertJSONL(b *testing.B) {
	// Generate JSONL with 50 lines
	var buf bytes.Buffer
	for i := 0; i < 50; i++ {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(`{"id":`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`,"name":"Item`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`","active":true}`)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertJSONL(buf.Bytes())
	}
}

func BenchmarkConvertJSONLStream(b *testing.B) {
	// Generate JSONL with 50 lines
	var buf bytes.Buffer
	for i := 0; i < 50; i++ {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(`{"id":`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`,"name":"Item`)
		buf.WriteString(string(rune('0' + (i % 10))))
		buf.WriteString(`","active":true}`)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertJSONLStream(buf.Bytes())
	}
}

func BenchmarkConvertAuto(b *testing.B) {
	// Test with JSON (auto-detect should identify as JSON)
	json := []byte(`{"id": 123, "name": "Ada", "items": [{"id": 1}, {"id": 2}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertAuto(json)
	}
}

func BenchmarkConvertToJSON(b *testing.B) {
	// First convert JSON to TOON
	json := []byte(`{"id": 123, "name": "Ada", "active": true, "items": ["a", "b", "c"]}`)
	toon, _ := Convert(json)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToJSON(toon)
	}
}

func BenchmarkConvertToJSONFromReader(b *testing.B) {
	// First convert JSON to TOON
	json := []byte(`{"id": 123, "name": "Ada", "active": true, "items": ["a", "b", "c"]}`)
	toon, _ := Convert(json)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertToJSONFromReader(bytes.NewReader(toon))
	}
}
