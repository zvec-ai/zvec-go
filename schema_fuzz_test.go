//go:build integration

package zvec

import (
	"testing"
)

// FuzzFieldSchemaName tests NewFieldSchema and SetName with arbitrary field names.
// This catches issues with special characters, null bytes, and extremely long names
// crossing the cgo boundary.
func FuzzFieldSchemaName(f *testing.F) {
	f.Add("normal_field")
	f.Add("")
	f.Add("field with spaces")
	f.Add("字段名")
	f.Add("field\x00name")
	f.Add("very_long_" + "abcdefghij_abcdefghij_abcdefghij_abcdefghij_abcdefghij")
	f.Add("special!@#$%^&*()")

	f.Fuzz(func(t *testing.T, name string) {
		field := NewFieldSchema(name, DataTypeString, false, 0)
		if field == nil {
			return
		}
		defer field.Destroy()

		got := field.GetName()
		expected := cgoExpectedString(name)
		if got != expected {
			t.Errorf("field name round-trip mismatch: got %q, want %q", got, expected)
		}
	})
}

// FuzzFieldSchemaSetName tests SetName with arbitrary names on an existing field.
func FuzzFieldSchemaSetName(f *testing.F) {
	f.Add("updated_name")
	f.Add("")
	f.Add("名前")
	f.Add("name\x00with\x00nulls")

	f.Fuzz(func(t *testing.T, newName string) {
		field := NewFieldSchema("original", DataTypeString, false, 0)
		if field == nil {
			t.Fatal("NewFieldSchema returned nil")
		}
		defer field.Destroy()

		err := field.SetName(newName)
		if err != nil {
			return
		}

		got := field.GetName()
		expected := cgoExpectedString(newName)
		if got != expected {
			t.Errorf("SetName round-trip mismatch: got %q, want %q", got, expected)
		}
	})
}

// FuzzCollectionSchemaName tests NewCollectionSchema with arbitrary collection names.
func FuzzCollectionSchemaName(f *testing.F) {
	f.Add("my_collection")
	f.Add("")
	f.Add("集合名称")
	f.Add("collection\x00name")
	f.Add("a/b/c")
	f.Add("col with spaces")

	f.Fuzz(func(t *testing.T, name string) {
		schema := NewCollectionSchema(name)
		if schema == nil {
			return
		}
		defer schema.Destroy()

		got := schema.GetName()
		expected := cgoExpectedString(name)
		if got != expected {
			t.Errorf("collection name round-trip mismatch: got %q, want %q", got, expected)
		}
	})
}

// FuzzHNSWIndexParams tests NewHNSWIndexParams with arbitrary M and efConstruction values.
// This catches issues with invalid parameter combinations and boundary values.
func FuzzHNSWIndexParams(f *testing.F) {
	f.Add(16, 200)
	f.Add(0, 0)
	f.Add(1, 1)
	f.Add(-1, -1)
	f.Add(1000, 10000)
	f.Add(2, 50)

	f.Fuzz(func(t *testing.T, m, efConstruction int) {
		params := NewHNSWIndexParams(MetricTypeCosine, m, efConstruction)
		if params == nil {
			return
		}
		defer params.Destroy()

		gotM := params.GetHNSWM()
		gotEf := params.GetHNSWEfConstruction()

		if gotM != m {
			t.Errorf("HNSW M round-trip mismatch: got %d, want %d", gotM, m)
		}
		if gotEf != efConstruction {
			t.Errorf("HNSW efConstruction round-trip mismatch: got %d, want %d", gotEf, efConstruction)
		}
	})
}

// FuzzFieldSchemaDimension tests SetDimension with arbitrary uint32 values.
func FuzzFieldSchemaDimension(f *testing.F) {
	f.Add(uint32(0))
	f.Add(uint32(4))
	f.Add(uint32(128))
	f.Add(uint32(768))
	f.Add(uint32(4294967295))

	f.Fuzz(func(t *testing.T, dimension uint32) {
		field := NewFieldSchema("vector_field", DataTypeVectorFP32, false, dimension)
		if field == nil {
			return
		}
		defer field.Destroy()

		got := field.GetDimension()
		if got != dimension {
			t.Errorf("dimension round-trip mismatch: got %d, want %d", got, dimension)
		}
	})
}
