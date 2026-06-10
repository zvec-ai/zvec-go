//go:build integration

package zvec

import (
	"testing"
)

// FuzzSearchQueryFilter tests SetFilter with arbitrary filter expressions.
// This catches issues with malformed expressions, injection-like strings,
// and special characters crossing the cgo boundary into the filter parser.
func FuzzSearchQueryFilter(f *testing.F) {
	f.Add("category == 'test'")
	f.Add("")
	f.Add("id > 0 AND id < 100")
	f.Add("name == '你好'")
	f.Add("field == 'value with \"quotes\"'")
	f.Add("a == 'b' OR (c > 1 AND d < 2)")
	f.Add("'; DROP TABLE --")
	f.Add("field\x00name == 'value'")

	f.Fuzz(func(t *testing.T, filter string) {
		query := NewSearchQuery()
		if query == nil {
			t.Fatal("NewSearchQuery returned nil")
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

// FuzzSearchQueryFieldName tests SetFieldName with arbitrary field names.
func FuzzSearchQueryFieldName(f *testing.F) {
	f.Add("embedding")
	f.Add("")
	f.Add("field_with_特殊字符")
	f.Add("very_long_" + "abcdefghij_abcdefghij_abcdefghij_abcdefghij_abcdefghij")
	f.Add("field\x00name")

	f.Fuzz(func(t *testing.T, fieldName string) {
		query := NewSearchQuery()
		if query == nil {
			t.Fatal("NewSearchQuery returned nil")
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

// FuzzSearchQueryTopK tests SetTopK with arbitrary integer values.
// This catches issues with negative values, zero, and extremely large values.
func FuzzSearchQueryTopK(f *testing.F) {
	f.Add(1)
	f.Add(10)
	f.Add(100)
	f.Add(0)
	f.Add(-1)
	f.Add(1000000)

	f.Fuzz(func(t *testing.T, topk int) {
		query := NewSearchQuery()
		if query == nil {
			t.Fatal("NewSearchQuery returned nil")
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

// FuzzGroupBySearchQueryFilter tests GroupBySearchQuery.SetFilter with arbitrary expressions.
func FuzzGroupBySearchQueryFilter(f *testing.F) {
	f.Add("category == 'test'")
	f.Add("")
	f.Add("id > 0")
	f.Add("'; DROP TABLE --")

	f.Fuzz(func(t *testing.T, filter string) {
		query := NewGroupBySearchQuery()
		if query == nil {
			t.Fatal("NewGroupBySearchQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetFilter(filter)
		if err != nil {
			return
		}
	})
}

// FuzzGroupBySearchQueryGroupCount tests SetGroupCount with arbitrary uint32 values.
func FuzzGroupBySearchQueryGroupCount(f *testing.F) {
	f.Add(uint32(1))
	f.Add(uint32(10))
	f.Add(uint32(0))
	f.Add(uint32(4294967295))

	f.Fuzz(func(t *testing.T, count uint32) {
		query := NewGroupBySearchQuery()
		if query == nil {
			t.Fatal("NewGroupBySearchQuery returned nil")
		}
		defer query.Destroy()

		err := query.SetGroupCount(count)
		if err != nil {
			return
		}
	})
}
