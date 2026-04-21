//go:build integration

package zvec

import (
	"strings"
	"testing"
)

// cgoExpectedString returns the expected string after a cgo round-trip.
// C strings are null-terminated, so any embedded null bytes cause truncation.
func cgoExpectedString(s string) string {
	if idx := strings.IndexByte(s, 0); idx >= 0 {
		return s[:idx]
	}
	return s
}

// FuzzDocStringField tests AddStringField/GetStringField with arbitrary strings.
// This catches issues with null bytes, unicode boundaries, and extremely long strings
// crossing the cgo boundary.
func FuzzDocStringField(f *testing.F) {
	f.Add("hello world")
	f.Add("")
	f.Add("你好世界")
	f.Add("emoji: 🚀🎉")
	f.Add("null\x00byte")
	f.Add("line\nbreak\ttab")
	f.Add("special chars: <>&\"'\\")

	f.Fuzz(func(t *testing.T, value string) {
		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		fieldName := "fuzz_field"
		err := doc.AddStringField(fieldName, value)
		if err != nil {
			return
		}

		got, err := doc.GetStringField(fieldName)
		if err != nil {
			t.Fatalf("GetStringField failed after successful AddStringField: %v", err)
		}

		// C.CString truncates at null byte, so AddStringField stores truncated value.
		// Only verify round-trip for strings without embedded null bytes.
		if !strings.ContainsRune(value, 0) {
			if got != value {
				t.Errorf("round-trip mismatch: got %q, want %q", got, value)
			}
		}
	})
}

// FuzzDocFieldName tests field operations with arbitrary field names.
// This catches issues with special characters in field names crossing the cgo boundary.
func FuzzDocFieldName(f *testing.F) {
	f.Add("normal_field")
	f.Add("")
	f.Add("field with spaces")
	f.Add("字段名")
	f.Add("field\x00name")
	f.Add("very_long_field_name_" + "abcdefghij_abcdefghij_abcdefghij_abcdefghij_abcdefghij")

	f.Fuzz(func(t *testing.T, fieldName string) {
		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		err := doc.AddStringField(fieldName, "test_value")
		if err != nil {
			return
		}

		// C truncates at null byte, so query with the truncated name
		cFieldName := cgoExpectedString(fieldName)
		if !doc.HasField(cFieldName) {
			t.Errorf("HasField(%q) returned false after successful AddStringField", cFieldName)
		}
	})
}

// FuzzDocVectorFP32Field tests AddVectorFP32Field with arbitrary float32 slices.
// This catches issues with NaN, Inf, subnormal floats, and extreme dimensions.
func FuzzDocVectorFP32Field(f *testing.F) {
	f.Add(float32(0.1), float32(0.2), float32(0.3), float32(0.4))
	f.Add(float32(0.0), float32(0.0), float32(0.0), float32(0.0))
	f.Add(float32(-1.0), float32(1.0), float32(-0.5), float32(0.5))

	f.Fuzz(func(t *testing.T, v1, v2, v3, v4 float32) {
		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		vector := []float32{v1, v2, v3, v4}
		err := doc.AddVectorFP32Field("fuzz_vector", vector)
		if err != nil {
			return
		}

		got, err := doc.GetVectorFP32Field("fuzz_vector")
		if err != nil {
			t.Fatalf("GetVectorFP32Field failed after successful Add: %v", err)
		}

		if len(got) != len(vector) {
			t.Fatalf("vector length mismatch: got %d, want %d", len(got), len(vector))
		}
	})
}

// FuzzDocPrimaryKey tests SetPK/GetPK with arbitrary primary key strings.
func FuzzDocPrimaryKey(f *testing.F) {
	f.Add("pk-123")
	f.Add("")
	f.Add("pk_with_特殊字符")
	f.Add("pk\x00null")
	f.Add("a]b[c{d}e")

	f.Fuzz(func(t *testing.T, pk string) {
		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		doc.SetPK(pk)
		got := doc.GetPK()

		// C.CString truncates at null byte; only verify for clean strings
		if !strings.ContainsRune(pk, 0) {
			if got != pk {
				t.Errorf("PK round-trip mismatch: got %q, want %q", got, pk)
			}
		}
	})
}

// FuzzDocInt32Field tests AddInt32Field/GetInt32Field with arbitrary int32 values.
func FuzzDocInt32Field(f *testing.F) {
	f.Add(int32(0))
	f.Add(int32(42))
	f.Add(int32(-42))
	f.Add(int32(2147483647))
	f.Add(int32(-2147483648))

	f.Fuzz(func(t *testing.T, value int32) {
		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		fieldName := "fuzz_int32"
		err := doc.AddInt32Field(fieldName, value)
		if err != nil {
			return
		}

		got, err := doc.GetInt32Field(fieldName)
		if err != nil {
			t.Fatalf("GetInt32Field failed after successful Add: %v", err)
		}

		if got != value {
			t.Errorf("int32 round-trip mismatch: got %d, want %d", got, value)
		}
	})
}

// FuzzDocInt64Field tests AddInt64Field/GetInt64Field with arbitrary int64 values.
func FuzzDocInt64Field(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(1234567890))
	f.Add(int64(-1234567890))
	f.Add(int64(9223372036854775807))
	f.Add(int64(-9223372036854775808))

	f.Fuzz(func(t *testing.T, value int64) {
		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		fieldName := "fuzz_int64"
		err := doc.AddInt64Field(fieldName, value)
		if err != nil {
			return
		}

		got, err := doc.GetInt64Field(fieldName)
		if err != nil {
			t.Fatalf("GetInt64Field failed after successful Add: %v", err)
		}

		if got != value {
			t.Errorf("int64 round-trip mismatch: got %d, want %d", got, value)
		}
	})
}

// FuzzDocBinaryField tests AddBinaryField with arbitrary byte slices.
func FuzzDocBinaryField(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add([]byte{0x01, 0x02, 0x03})
	f.Add([]byte{0xFF, 0xFE, 0xFD})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}

		doc := NewDoc()
		if doc == nil {
			t.Fatal("NewDoc() returned nil")
		}
		defer doc.Destroy()

		err := doc.AddBinaryField("fuzz_binary", data)
		if err != nil {
			return
		}

		if !doc.HasField("fuzz_binary") {
			t.Error("HasField returned false after successful AddBinaryField")
		}
	})
}
