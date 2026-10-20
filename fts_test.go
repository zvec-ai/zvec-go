//go:build integration

package zvec

import (
	"path/filepath"
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
	defer got.Destroy()
	if !got.owned {
		t.Fatal("GetFTS returned a borrowed payload, want an owned copy")
	}
	query.Destroy()
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

// queryFTSMatch runs a match query against the "content" field and returns the
// matching documents.
func queryFTSMatch(t *testing.T, collection *Collection, term string) []*Doc {
	t.Helper()

	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()
	if err := query.SetFieldName("content"); err != nil {
		t.Fatalf("SetFieldName() failed: %v", err)
	}
	if err := query.SetTopK(10); err != nil {
		t.Fatalf("SetTopK() failed: %v", err)
	}
	if err := query.SetOutputFields([]string{"id", "content"}); err != nil {
		t.Fatalf("SetOutputFields() failed: %v", err)
	}

	payload := NewFTS()
	if payload == nil {
		t.Fatal("NewFTS returned nil")
	}
	defer payload.Destroy()
	if err := payload.SetMatchString(term); err != nil {
		t.Fatalf("SetMatchString() failed: %v", err)
	}
	if err := query.SetFTS(payload); err != nil {
		t.Fatalf("SetFTS() failed: %v", err)
	}

	docs, err := collection.Query(query)
	if err != nil {
		t.Fatalf("Query(%q) failed: %v", term, err)
	}
	return docs
}

// TestFTSJiebaTokenizerEmbeddedDicts verifies that the `jieba` tokenizer works
// out of the box. Since zvec v0.8.0 the dictionary files are embedded in the
// C-API library, so no external dictionary directory is required.
func TestFTSJiebaTokenizerEmbeddedDicts(t *testing.T) {
	schema := NewCollectionSchema("fts_jieba_collection")
	if schema == nil {
		t.Fatal("NewCollectionSchema returned nil")
	}
	defer schema.Destroy()

	idField := NewFieldSchema("id", DataTypeString, false, 0)
	defer idField.Destroy()
	idIndex, err := NewInvertIndexParams(true, false)
	if err != nil {
		t.Fatalf("NewInvertIndexParams() failed: %v", err)
	}
	defer idIndex.Destroy()
	if err := idField.SetIndexParams(idIndex); err != nil {
		t.Fatalf("SetIndexParams(id) failed: %v", err)
	}
	if err := schema.AddField(idField); err != nil {
		t.Fatalf("AddField(id) failed: %v", err)
	}

	contentField := NewFieldSchema("content", DataTypeString, false, 0)
	defer contentField.Destroy()
	ftsIndex, err := NewFTSIndexParams("jieba", nil, "")
	if err != nil {
		t.Fatalf("NewFTSIndexParams(jieba) failed: %v", err)
	}
	defer ftsIndex.Destroy()
	if err := contentField.SetIndexParams(ftsIndex); err != nil {
		t.Fatalf("SetIndexParams(content) failed: %v", err)
	}
	if err := schema.AddField(contentField); err != nil {
		t.Fatalf("AddField(content) failed: %v", err)
	}

	path := filepath.Join(testTempDir(t), "fts_jieba")
	collection, err := CreateAndOpen(path, schema, nil)
	if err != nil {
		t.Fatalf("CreateAndOpen() failed: %v", err)
	}
	t.Cleanup(func() { _ = collection.Close() })

	contents := [][2]string{
		{"doc1", "向量数据库支持全文检索功能"},
		{"doc2", "全文检索依赖分词器切分中文"},
		{"doc3", "数据库优化需要索引结构"},
	}
	docs := make([]*Doc, 0, len(contents))
	for _, item := range contents {
		doc := NewDoc()
		if doc == nil {
			t.Fatalf("NewDoc(%s) returned nil", item[0])
		}
		docs = append(docs, doc)
		doc.SetPK(item[0])
		if err := doc.AddStringField("id", item[0]); err != nil {
			t.Fatalf("AddStringField(id) failed: %v", err)
		}
		if err := doc.AddStringField("content", item[1]); err != nil {
			t.Fatalf("AddStringField(content) failed: %v", err)
		}
	}
	defer FreeDocs(docs)

	result, err := collection.Insert(docs)
	if err != nil {
		t.Fatalf("Insert() failed: %v", err)
	}
	if result.SuccessCount != uint64(len(contents)) || result.ErrorCount != 0 {
		t.Fatalf("Insert() success=%d error=%d, want success=%d error=0",
			result.SuccessCount, result.ErrorCount, len(contents))
	}
	if err := collection.Flush(); err != nil {
		t.Fatalf("Flush() failed: %v", err)
	}

	for _, tc := range []struct {
		term     string
		wantHits int
	}{
		{"检索", 2},
		{"数据库", 2},
		{"分词器", 1},
	} {
		hits := queryFTSMatch(t, collection, tc.term)
		if len(hits) != tc.wantHits {
			t.Errorf("Query(%q) returned %d hits, want %d", tc.term, len(hits), tc.wantHits)
		}
		FreeDocs(hits)
	}
}
