//go:build integration

package zvec

import (
	"testing"
)

func TestNewFTS(t *testing.T) {
	fts := NewFTS()
	if fts == nil {
		t.Fatal("NewFTS returned nil")
	}
	defer fts.Destroy()
}

func TestFTSSetGetQueryString(t *testing.T) {
	fts := NewFTS()
	if fts == nil {
		t.Fatal("NewFTS returned nil")
	}
	defer fts.Destroy()

	err := fts.SetQueryString("hello AND world")
	if err != nil {
		t.Fatalf("SetQueryString failed: %v", err)
	}
	got := fts.GetQueryString()
	if got != "hello AND world" {
		t.Errorf("GetQueryString() = %q, want %q", got, "hello AND world")
	}
}

func TestFTSSetGetMatchString(t *testing.T) {
	fts := NewFTS()
	if fts == nil {
		t.Fatal("NewFTS returned nil")
	}
	defer fts.Destroy()

	err := fts.SetMatchString("natural language query")
	if err != nil {
		t.Fatalf("SetMatchString failed: %v", err)
	}
	got := fts.GetMatchString()
	if got != "natural language query" {
		t.Errorf("GetMatchString() = %q, want %q", got, "natural language query")
	}
}

func TestFTSDestroy(t *testing.T) {
	fts := NewFTS()
	if fts == nil {
		t.Fatal("NewFTS returned nil")
	}
	fts.Destroy()
	fts.Destroy() // double destroy should not panic
}

func TestNewFTSQueryParams(t *testing.T) {
	params := NewFTSQueryParams("")
	if params == nil {
		t.Fatal("NewFTSQueryParams returned nil")
	}
	defer params.Destroy()
}

func TestFTSQueryParamsWithOperator(t *testing.T) {
	params := NewFTSQueryParams("AND")
	if params == nil {
		t.Fatal("NewFTSQueryParams returned nil")
	}
	defer params.Destroy()

	got := params.GetDefaultOperator()
	if got != "AND" {
		t.Errorf("GetDefaultOperator() = %q, want %q", got, "AND")
	}
}

func TestFTSQueryParamsSetDefaultOperator(t *testing.T) {
	params := NewFTSQueryParams("")
	if params == nil {
		t.Fatal("NewFTSQueryParams returned nil")
	}
	defer params.Destroy()

	err := params.SetDefaultOperator("OR")
	if err != nil {
		t.Fatalf("SetDefaultOperator failed: %v", err)
	}
	got := params.GetDefaultOperator()
	if got != "OR" {
		t.Errorf("GetDefaultOperator() = %q, want %q", got, "OR")
	}
}

func TestFTSQueryParamsDestroy(t *testing.T) {
	params := NewFTSQueryParams("AND")
	if params == nil {
		t.Fatal("NewFTSQueryParams returned nil")
	}
	params.Destroy()
	params.Destroy() // double destroy should not panic
}

func TestSearchQuerySetFTSParams(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	params := NewFTSQueryParams("AND")
	if params == nil {
		t.Fatal("NewFTSQueryParams returned nil")
	}

	err := query.SetFTSParams(params)
	if err != nil {
		t.Fatalf("SetFTSParams failed: %v", err)
	}
	if params.handle != nil {
		t.Error("params.handle should be nil after ownership transfer")
	}
}

func TestSearchQuerySetFTS(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	fts := NewFTS()
	if fts == nil {
		t.Fatal("NewFTS returned nil")
	}
	defer fts.Destroy()

	err := fts.SetQueryString("test query")
	if err != nil {
		t.Fatalf("SetQueryString failed: %v", err)
	}

	err = query.SetFTS(fts)
	if err != nil {
		t.Fatalf("SetFTS failed: %v", err)
	}

	// fts should still be valid (copied, not transferred)
	if fts.handle == nil {
		t.Error("fts.handle should still be valid after SetFTS (copy semantics)")
	}

	got := query.GetFTS()
	if got == nil {
		t.Fatal("GetFTS returned nil")
	}
	if got.GetQueryString() != "test query" {
		t.Errorf("GetFTS().GetQueryString() = %q, want %q", got.GetQueryString(), "test query")
	}
}

func TestNewFTSIndexParams(t *testing.T) {
	params, err := NewFTSIndexParams("default", nil, "")
	if err != nil {
		t.Fatalf("NewFTSIndexParams failed: %v", err)
	}
	defer params.Destroy()

	if params.GetType() != IndexTypeFTS {
		t.Errorf("GetType() = %v, want IndexTypeFTS", params.GetType())
	}
}

func TestFTSIndexParamsWithFilters(t *testing.T) {
	filters := []string{"lowercase", "stop_words"}
	params, err := NewFTSIndexParams("jieba", filters, `{"key":"value"}`)
	if err != nil {
		t.Fatalf("NewFTSIndexParams failed: %v", err)
	}
	defer params.Destroy()

	tokenizer, gotFilters, extra, err := params.GetFTSParams()
	if err != nil {
		t.Fatalf("GetFTSParams failed: %v", err)
	}
	if tokenizer != "jieba" {
		t.Errorf("tokenizer = %q, want %q", tokenizer, "jieba")
	}
	if len(gotFilters) != 2 {
		t.Fatalf("len(filters) = %d, want 2", len(gotFilters))
	}
	if gotFilters[0] != "lowercase" || gotFilters[1] != "stop_words" {
		t.Errorf("filters = %v, want [lowercase stop_words]", gotFilters)
	}
	if extra != `{"key":"value"}` {
		t.Errorf("extra = %q, want %q", extra, `{"key":"value"}`)
	}
}
