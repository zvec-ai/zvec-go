//go:build (purego || !cgo) && integration

package zvec

import (
	"path/filepath"
	"strings"
	"testing"
)

type puregoFTSSample struct {
	id      string
	content string
	vector  []float32
}

func TestPuregoFTSAndHybridQuery(t *testing.T) {
	ensurePuregoInitialized(t)

	schema := NewCollectionSchema("purego_fts_hybrid")
	if schema == nil {
		t.Fatal("schema stage: NewCollectionSchema() returned nil")
	}
	defer schema.Destroy()

	idField := NewFieldSchema("id", DataTypeString, false, 0)
	if idField == nil {
		t.Fatal("schema stage: NewFieldSchema(id) returned nil")
	}
	defer idField.Destroy()
	idIndex, err := NewInvertIndexParams(true, false)
	if err != nil {
		t.Fatalf("schema stage: NewInvertIndexParams(id) failed: %v", err)
	}
	defer idIndex.Destroy()
	if err := idField.SetIndexParams(idIndex); err != nil {
		t.Fatalf("schema stage: id.SetIndexParams() failed: %v", err)
	}
	if err := schema.AddField(idField); err != nil {
		t.Fatalf("schema stage: AddField(id) failed: %v", err)
	}

	contentField := NewFieldSchema("content", DataTypeString, false, 0)
	if contentField == nil {
		t.Fatal("schema stage: NewFieldSchema(content) returned nil")
	}
	defer contentField.Destroy()
	ftsIndex, err := NewFTSIndexParams("whitespace", []string{"lowercase"}, "")
	if err != nil {
		t.Fatalf("schema stage: NewFTSIndexParams(content) failed: %v", err)
	}
	defer ftsIndex.Destroy()
	if err := contentField.SetIndexParams(ftsIndex); err != nil {
		t.Fatalf("schema stage: content.SetIndexParams() failed: %v", err)
	}
	if err := schema.AddField(contentField); err != nil {
		t.Fatalf("schema stage: AddField(content) failed: %v", err)
	}

	vectorField := NewFieldSchema("embedding", DataTypeVectorFP32, false, 4)
	if vectorField == nil {
		t.Fatal("schema stage: NewFieldSchema(embedding) returned nil")
	}
	defer vectorField.Destroy()
	vectorIndex, err := NewFlatIndexParams(MetricTypeIP)
	if err != nil {
		t.Fatalf("schema stage: NewFlatIndexParams(embedding) failed: %v", err)
	}
	defer vectorIndex.Destroy()
	if err := vectorField.SetIndexParams(vectorIndex); err != nil {
		t.Fatalf("schema stage: embedding.SetIndexParams() failed: %v", err)
	}
	if err := schema.AddField(vectorField); err != nil {
		t.Fatalf("schema stage: AddField(embedding) failed: %v", err)
	}

	collection, err := CreateAndOpen(filepath.Join(t.TempDir(), "collection"), schema, nil)
	if err != nil {
		t.Fatalf("collection stage: CreateAndOpen() failed: %v", err)
	}
	defer func() {
		if err := collection.Close(); err != nil {
			t.Errorf("collection cleanup stage: Close() failed: %v", err)
		}
	}()

	samples := []puregoFTSSample{
		{"doc1", "The quick brown fox jumps over the lazy dog", []float32{1.0, 0.0, 0.0, 0.0}},
		{"doc2", "A fast red fox runs through the forest", []float32{0.9, 0.1, 0.0, 0.0}},
		{"doc3", "The lazy cat sleeps on the couch all day", []float32{0.0, 1.0, 0.0, 0.0}},
		{"doc4", "Dogs and cats are popular household pets", []float32{0.0, 0.8, 0.2, 0.0}},
		{"doc5", "The fox and the hound is a classic story", []float32{0.7, 0.2, 0.0, 0.1}},
	}
	wantContent := make(map[string]string, len(samples))
	docs := make([]*Doc, 0, len(samples))
	defer func() { FreeDocs(docs) }()
	for _, sample := range samples {
		wantContent[sample.id] = sample.content
		doc := NewDoc()
		if doc == nil {
			t.Fatalf("insert stage: NewDoc(%s) returned nil", sample.id)
		}
		docs = append(docs, doc)
		doc.SetPK(sample.id)
		if err := doc.AddStringField("id", sample.id); err != nil {
			t.Fatalf("insert stage: AddStringField(id) for %s failed: %v", sample.id, err)
		}
		if err := doc.AddStringField("content", sample.content); err != nil {
			t.Fatalf("insert stage: AddStringField(content) for %s failed: %v", sample.id, err)
		}
		if err := doc.AddVectorFP32Field("embedding", sample.vector); err != nil {
			t.Fatalf("insert stage: AddVectorFP32Field(embedding) for %s failed: %v", sample.id, err)
		}
	}

	insertResult, err := collection.Insert(docs)
	if err != nil {
		t.Fatalf("insert stage: Insert() failed: %v", err)
	}
	if insertResult.SuccessCount != uint64(len(samples)) || insertResult.ErrorCount != 0 {
		t.Fatalf(
			"insert stage: Insert() success=%d error=%d, want success=%d error=0",
			insertResult.SuccessCount,
			insertResult.ErrorCount,
			len(samples),
		)
	}
	if err := collection.Flush(); err != nil {
		t.Fatalf("flush stage: Flush() failed: %v", err)
	}

	ftsQuery := NewSearchQuery()
	if ftsQuery == nil {
		t.Fatal("FTS query stage: NewSearchQuery() returned nil")
	}
	defer ftsQuery.Destroy()
	if err := ftsQuery.SetFieldName("content"); err != nil {
		t.Fatalf("FTS query stage: SetFieldName() failed: %v", err)
	}
	if err := ftsQuery.SetTopK(10); err != nil {
		t.Fatalf("FTS query stage: SetTopK() failed: %v", err)
	}
	if err := ftsQuery.SetOutputFields([]string{"id", "content"}); err != nil {
		t.Fatalf("FTS query stage: SetOutputFields() failed: %v", err)
	}
	ftsPayload := NewFTS()
	if ftsPayload == nil {
		t.Fatal("FTS query stage: NewFTS() returned nil")
	}
	defer ftsPayload.Destroy()
	if err := ftsPayload.SetMatchString("fox"); err != nil {
		t.Fatalf("FTS query stage: SetMatchString() failed: %v", err)
	}
	if err := ftsQuery.SetFTS(ftsPayload); err != nil {
		t.Fatalf("FTS query stage: SetFTS() failed: %v", err)
	}

	ftsResults, err := collection.Query(ftsQuery)
	if err != nil {
		t.Fatalf("FTS query stage: Query() failed: %v", err)
	}
	defer FreeDocs(ftsResults)
	assertPuregoFTSResults(t, "FTS query stage", ftsResults, 3, wantContent, true)

	hybridQuery := NewMultiQuery()
	if hybridQuery == nil {
		t.Fatal("hybrid query stage: NewMultiQuery() returned nil")
	}
	defer hybridQuery.Destroy()
	if err := hybridQuery.SetTopK(3); err != nil {
		t.Fatalf("hybrid query stage: SetTopK() failed: %v", err)
	}
	if err := hybridQuery.SetOutputFields([]string{"id", "content"}); err != nil {
		t.Fatalf("hybrid query stage: SetOutputFields() failed: %v", err)
	}
	if err := hybridQuery.SetRerankRRF(60); err != nil {
		t.Fatalf("hybrid query stage: SetRerankRRF() failed: %v", err)
	}

	vectorSubQuery := NewSubQuery()
	if vectorSubQuery == nil {
		t.Fatal("hybrid vector stage: NewSubQuery() returned nil")
	}
	defer vectorSubQuery.Destroy()
	if err := vectorSubQuery.SetFieldName("embedding"); err != nil {
		t.Fatalf("hybrid vector stage: SetFieldName() failed: %v", err)
	}
	if err := vectorSubQuery.SetNumCandidates(5); err != nil {
		t.Fatalf("hybrid vector stage: SetNumCandidates() failed: %v", err)
	}
	if err := vectorSubQuery.SetQueryVector([]float32{1.0, 0.0, 0.0, 0.0}); err != nil {
		t.Fatalf("hybrid vector stage: SetQueryVector() failed: %v", err)
	}
	if err := hybridQuery.AddSubQuery(vectorSubQuery); err != nil {
		t.Fatalf("hybrid vector stage: AddSubQuery() failed: %v", err)
	}

	ftsSubQuery := NewSubQuery()
	if ftsSubQuery == nil {
		t.Fatal("hybrid FTS stage: NewSubQuery() returned nil")
	}
	defer ftsSubQuery.Destroy()
	if err := ftsSubQuery.SetFieldName("content"); err != nil {
		t.Fatalf("hybrid FTS stage: SetFieldName() failed: %v", err)
	}
	if err := ftsSubQuery.SetNumCandidates(5); err != nil {
		t.Fatalf("hybrid FTS stage: SetNumCandidates() failed: %v", err)
	}
	hybridFTS := NewFTS()
	if hybridFTS == nil {
		t.Fatal("hybrid FTS stage: NewFTS() returned nil")
	}
	defer hybridFTS.Destroy()
	if err := hybridFTS.SetMatchString("fox"); err != nil {
		t.Fatalf("hybrid FTS stage: SetMatchString() failed: %v", err)
	}
	if err := ftsSubQuery.SetFTS(hybridFTS); err != nil {
		t.Fatalf("hybrid FTS stage: SetFTS() failed: %v", err)
	}
	if err := hybridQuery.AddSubQuery(ftsSubQuery); err != nil {
		t.Fatalf("hybrid FTS stage: AddSubQuery() failed: %v", err)
	}

	hybridResults, err := collection.MultiQuery(hybridQuery)
	if err != nil {
		t.Fatalf("hybrid query stage: MultiQuery() failed: %v", err)
	}
	defer FreeDocs(hybridResults)
	assertPuregoFTSResults(t, "hybrid query stage", hybridResults, 3, wantContent, true)
}

func assertPuregoFTSResults(
	t *testing.T,
	stage string,
	docs []*Doc,
	wantCount int,
	wantContent map[string]string,
	wantFox bool,
) {
	t.Helper()
	if len(docs) != wantCount {
		t.Fatalf("%s: got %d results, want %d", stage, len(docs), wantCount)
	}

	seen := make(map[string]struct{}, len(docs))
	for i, doc := range docs {
		if doc == nil {
			t.Fatalf("%s: result %d is nil", stage, i)
		}
		pk := doc.GetPK()
		expectedContent, ok := wantContent[pk]
		if !ok {
			t.Fatalf("%s: result %d has unexpected primary key %q", stage, i, pk)
		}
		if _, duplicate := seen[pk]; duplicate {
			t.Fatalf("%s: primary key %q appeared more than once", stage, pk)
		}
		seen[pk] = struct{}{}

		id, err := doc.GetStringField("id")
		if err != nil {
			t.Fatalf("%s: GetStringField(id) for %q failed: %v", stage, pk, err)
		}
		if id != pk {
			t.Fatalf("%s: id field %q does not match primary key %q", stage, id, pk)
		}
		content, err := doc.GetStringField("content")
		if err != nil {
			t.Fatalf("%s: GetStringField(content) for %q failed: %v", stage, pk, err)
		}
		if content != expectedContent {
			t.Fatalf("%s: content for %q = %q, want %q", stage, pk, content, expectedContent)
		}
		if wantFox && !strings.Contains(strings.ToLower(content), "fox") {
			t.Fatalf("%s: content for %q does not contain the FTS term: %q", stage, pk, content)
		}
	}
}
