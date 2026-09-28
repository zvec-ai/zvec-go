//go:build integration && cgo && !purego

package zvec

import (
	"strconv"
	"testing"
)

// TestDocFieldNameInternFallback verifies the cgo field accessors keep
// working when the intern table is full and names degrade to fresh
// C.CStrings allocated and freed per call.
func TestDocFieldNameInternFallback(t *testing.T) {
	doc := NewDoc()
	if doc == nil {
		t.Fatal("NewDoc() returned nil")
	}
	defer doc.Destroy()

	// Blow past the intern cap with distinct field names; every accessor
	// below takes the fresh-CString fallback path internally but must
	// behave identically.
	const extra = 32
	for i := 0; i < maxInternedFieldNames+extra; i++ {
		if err := doc.AddStringField("field_"+strconv.Itoa(i), "v"); err != nil {
			t.Fatalf("AddStringField failed at %d: %v", i, err)
		}
	}
	for i := 0; i < 3; i++ {
		name := "field_" + strconv.Itoa(i)
		if !doc.HasField(name) {
			t.Fatalf("HasField(%q) = false, want true", name)
		}
		value, err := doc.GetStringField(name)
		if err != nil {
			t.Fatalf("GetStringField failed: %v", err)
		}
		if value != "v" {
			t.Fatalf("GetStringField(%q) = %q, want %q", name, value, "v")
		}
	}
}
