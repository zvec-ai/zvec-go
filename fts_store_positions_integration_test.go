//go:build integration

package zvec

import (
	"path/filepath"
	"testing"
)

type ftsPositionsSample struct {
	id      string
	content string
}

var ftsPositionsSamples = []ftsPositionsSample{
	{"doc1", "the quick brown fox jumps over the lazy dog"},
	{"doc2", "a fast red fox runs through the forest"},
	{"doc3", "the lazy cat sleeps on the couch all day"},
	{"doc4", "dogs and cats are popular household pets"},
	{"doc5", "the fox and the hound is a classic story"},
}

// ftsPositionsQueryDocs runs an FTS query with either a boolean query string
// or a natural-language match string, so callers can assert either successful
// results or a rejection.
func ftsPositionsQueryDocs(t *testing.T, collection *Collection, queryString, matchString string) ([]*Doc, error) {
	t.Helper()
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery() returned nil")
	}
	defer query.Destroy()
	if err := query.SetFieldName("content"); err != nil {
		t.Fatalf("SetFieldName() failed: %v", err)
	}
	if err := query.SetTopK(10); err != nil {
		t.Fatalf("SetTopK() failed: %v", err)
	}
	fts := NewFTS()
	if fts == nil {
		t.Fatal("NewFTS() returned nil")
	}
	defer fts.Destroy()
	if queryString != "" {
		if err := fts.SetQueryString(queryString); err != nil {
			t.Fatalf("SetQueryString(%q) failed: %v", queryString, err)
		}
	}
	if matchString != "" {
		if err := fts.SetMatchString(matchString); err != nil {
			t.Fatalf("SetMatchString(%q) failed: %v", matchString, err)
		}
	}
	if err := query.SetFTS(fts); err != nil {
		t.Fatalf("SetFTS() failed: %v", err)
	}
	return collection.Query(query)
}

func ftsPositionsCreateCollection(t *testing.T, path string, extraParams string) *Collection {
	t.Helper()
	schema := NewCollectionSchema("fts_positions")
	if schema == nil {
		t.Fatal("NewCollectionSchema() returned nil")
	}
	defer schema.Destroy()
	content := NewFieldSchema("content", DataTypeString, false, 0)
	if content == nil {
		t.Fatal("NewFieldSchema(content) returned nil")
	}
	defer content.Destroy()
	ftsIndex, err := NewFTSIndexParams("whitespace", []string{"lowercase"}, extraParams)
	if err != nil {
		t.Fatalf("NewFTSIndexParams(content) failed: %v", err)
	}
	defer ftsIndex.Destroy()
	if err := content.SetIndexParams(ftsIndex); err != nil {
		t.Fatalf("content.SetIndexParams() failed: %v", err)
	}
	if err := schema.AddField(content); err != nil {
		t.Fatalf("AddField(content) failed: %v", err)
	}

	collection, err := CreateAndOpen(path, schema, nil)
	if err != nil {
		t.Fatalf("CreateAndOpen() failed: %v", err)
	}
	return collection
}

func ftsPositionsInsertSamples(t *testing.T, collection *Collection) {
	t.Helper()
	docs := make([]*Doc, 0, len(ftsPositionsSamples))
	defer FreeDocs(docs)
	for _, sample := range ftsPositionsSamples {
		doc := NewDoc()
		if doc == nil {
			t.Fatalf("NewDoc(%s) returned nil", sample.id)
		}
		docs = append(docs, doc)
		doc.SetPK(sample.id)
		if err := doc.AddStringField("content", sample.content); err != nil {
			t.Fatalf("AddStringField(content) for %s failed: %v", sample.id, err)
		}
	}
	result, err := collection.Insert(docs)
	if err != nil || result.SuccessCount != uint64(len(ftsPositionsSamples)) {
		t.Fatalf("Insert() returned result=%v, err=%v", result, err)
	}
}

func TestFTSPhraseQueryWithPositions(t *testing.T) {
	collection := ftsPositionsCreateCollection(t, filepath.Join(testTempDir(t), "collection"), "")
	defer func() { _ = collection.Close() }()
	ftsPositionsInsertSamples(t, collection)

	docs, err := ftsPositionsQueryDocs(t, collection, `"lazy dog"`, "")
	if err != nil {
		t.Fatalf("phrase Query() failed: %v", err)
	}
	defer FreeDocs(docs)
	if len(docs) != 1 || docs[0].GetPK() != "doc1" {
		t.Fatalf("phrase Query() returned %d documents, want exactly doc1", len(docs))
	}
}

func TestFTSStorePositionsDisabled(t *testing.T) {
	path := filepath.Join(testTempDir(t), "collection")
	collection := ftsPositionsCreateCollection(t, path, `{"store_positions":false}`)
	ftsPositionsInsertSamples(t, collection)
	if err := collection.Flush(); err != nil {
		t.Fatalf("Flush() failed: %v", err)
	}

	// Term queries must keep working with positions disabled.
	docs, err := ftsPositionsQueryDocs(t, collection, "", "fox")
	if err != nil {
		t.Fatalf("term Query() failed: %v", err)
	}
	defer FreeDocs(docs)
	if len(docs) != 3 {
		t.Fatalf("term Query() returned %d documents, want 3", len(docs))
	}

	// Phrase queries are rejected rather than silently returning wrong data.
	if _, err := ftsPositionsQueryDocs(t, collection, `"lazy dog"`, ""); err == nil {
		t.Fatal("phrase Query() unexpectedly succeeded, want rejection when positions are not stored")
	}

	if err := collection.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Persistence: the disabled flag must survive a reopen, both for term
	// queries and for the phrase rejection.
	reopened, err := Open(path, nil)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	docs, err = ftsPositionsQueryDocs(t, reopened, "", "fox")
	if err != nil {
		t.Fatalf("term Query() after reopen failed: %v", err)
	}
	defer FreeDocs(docs)
	if len(docs) != 3 {
		t.Fatalf("term Query() after reopen returned %d documents, want 3", len(docs))
	}
	if _, err := ftsPositionsQueryDocs(t, reopened, `"lazy dog"`, ""); err == nil {
		t.Fatal("phrase Query() after reopen unexpectedly succeeded, want rejection when positions are not stored")
	}
}
