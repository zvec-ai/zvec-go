//go:build integration

package zvec

import (
	"testing"
)

// FuzzVectorQueryFilter tests SetFilter with arbitrary filter expressions.
// This catches issues with malformed expressions, injection-like strings,
// and special characters crossing the cgo boundary into the filter parser.
func FuzzVectorQueryFilter(f *testing.F) {
	f.Add("category == 'test'")
	f.Add("")
	f.Add("id > 0 AND id < 100")
	f.Add("name == '你好'")
	f.Add("field == 'value with \"quotes\"'")
	f.Add("a == 'b' OR (c > 1 AND d < 2)")
	f.Add("'; DROP TABLE --")
	f.Add("field\x00name == 'value'")

	f.Fuzz(func(t *testing.T, filter string) {
		query := NewVectorQuery()
		if query == nil {
			t.Fatal("NewVectorQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetFilter(filter)
		if err != nil {
			return
		}

		got := query.GetFilter()
		expectedFilter := cgoExpectedString(filter)
		if got != expectedFilter {
			t.Errorf("filter round-trip mismatch: got %q, want %q", got, expectedFilter)
		}
	})
}

// FuzzVectorQueryFieldName tests SetFieldName with arbitrary field names.
func FuzzVectorQueryFieldName(f *testing.F) {
	f.Add("embedding")
	f.Add("")
	f.Add("field_with_特殊字符")
	f.Add("very_long_" + "abcdefghij_abcdefghij_abcdefghij_abcdefghij_abcdefghij")
	f.Add("field\x00name")

	f.Fuzz(func(t *testing.T, fieldName string) {
		query := NewVectorQuery()
		if query == nil {
			t.Fatal("NewVectorQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetFieldName(fieldName)
		if err != nil {
			return
		}

		got := query.GetFieldName()
		expectedName := cgoExpectedString(fieldName)
		if got != expectedName {
			t.Errorf("field name round-trip mismatch: got %q, want %q", got, expectedName)
		}
	})
}

// FuzzVectorQueryTopK tests SetTopK with arbitrary integer values.
// This catches issues with negative values, zero, and extremely large values.
func FuzzVectorQueryTopK(f *testing.F) {
	f.Add(1)
	f.Add(10)
	f.Add(100)
	f.Add(0)
	f.Add(-1)
	f.Add(1000000)

	f.Fuzz(func(t *testing.T, topk int) {
		query := NewVectorQuery()
		if query == nil {
			t.Fatal("NewVectorQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetTopK(topk)
		if err != nil {
			return
		}

		got := query.GetTopK()
		if got != topk {
			t.Errorf("topk round-trip mismatch: got %d, want %d", got, topk)
		}
	})
}

// FuzzGroupByVectorQueryFilter tests GroupByVectorQuery.SetFilter with arbitrary expressions.
func FuzzGroupByVectorQueryFilter(f *testing.F) {
	f.Add("category == 'test'")
	f.Add("")
	f.Add("id > 0")
	f.Add("'; DROP TABLE --")

	f.Fuzz(func(t *testing.T, filter string) {
		query := NewGroupByVectorQuery()
		if query == nil {
			t.Fatal("NewGroupByVectorQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetFilter(filter)
		if err != nil {
			return
		}
	})
}

// FuzzGroupByVectorQueryGroupCount tests SetGroupCount with arbitrary uint32 values.
func FuzzGroupByVectorQueryGroupCount(f *testing.F) {
	f.Add(uint32(1))
	f.Add(uint32(10))
	f.Add(uint32(0))
	f.Add(uint32(4294967295))

	f.Fuzz(func(t *testing.T, count uint32) {
		query := NewGroupByVectorQuery()
		if query == nil {
			t.Fatal("NewGroupByVectorQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetGroupCount(count)
		if err != nil {
			return
		}
	})
}
