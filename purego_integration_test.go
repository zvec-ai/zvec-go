//go:build (purego || !cgo) && integration

package zvec

import (
	"path/filepath"
	"testing"
)

func ensurePuregoInitialized(t *testing.T) {
	t.Helper()
	if IsInitialized() {
		return
	}
	if err := Initialize(nil); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
}

func newPuregoIntegrationSchema(t *testing.T, name string) *CollectionSchema {
	t.Helper()

	schema := NewCollectionSchema(name)
	if schema == nil {
		t.Fatal("NewCollectionSchema() returned nil")
	}

	idParams, err := NewInvertIndexParams(true, false)
	if err != nil {
		t.Fatalf("NewInvertIndexParams() failed: %v", err)
	}
	defer idParams.Destroy()
	idField := NewFieldSchema("id", DataTypeString, false, 0)
	if idField == nil {
		t.Fatal("NewFieldSchema(id) returned nil")
	}
	defer idField.Destroy()
	if err := idField.SetIndexParams(idParams); err != nil {
		t.Fatalf("id.SetIndexParams() failed: %v", err)
	}
	if err := schema.AddField(idField); err != nil {
		t.Fatalf("schema.AddField(id) failed: %v", err)
	}

	textField := NewFieldSchema("text", DataTypeString, true, 0)
	if textField == nil {
		t.Fatal("NewFieldSchema(text) returned nil")
	}
	defer textField.Destroy()
	if err := schema.AddField(textField); err != nil {
		t.Fatalf("schema.AddField(text) failed: %v", err)
	}

	vectorParams, err := NewFlatIndexParams(MetricTypeCosine)
	if err != nil {
		t.Fatalf("NewFlatIndexParams() failed: %v", err)
	}
	defer vectorParams.Destroy()
	vectorField := NewFieldSchema("embedding", DataTypeVectorFP32, false, 4)
	if vectorField == nil {
		t.Fatal("NewFieldSchema(embedding) returned nil")
	}
	defer vectorField.Destroy()
	if err := vectorField.SetIndexParams(vectorParams); err != nil {
		t.Fatalf("embedding.SetIndexParams() failed: %v", err)
	}
	if err := schema.AddField(vectorField); err != nil {
		t.Fatalf("schema.AddField(embedding) failed: %v", err)
	}

	return schema
}

func newPuregoIntegrationDoc(t *testing.T, pk, text string, vector []float32) *Doc {
	t.Helper()

	doc := NewDoc()
	if doc == nil {
		t.Fatal("NewDoc() returned nil")
	}
	doc.SetPK(pk)
	if err := doc.AddStringField("id", pk); err != nil {
		t.Fatalf("doc.AddStringField(id) failed: %v", err)
	}
	if err := doc.AddStringField("text", text); err != nil {
		t.Fatalf("doc.AddStringField(text) failed: %v", err)
	}
	if err := doc.AddVectorFP32Field("embedding", vector); err != nil {
		t.Fatalf("doc.AddVectorFP32Field(embedding) failed: %v", err)
	}
	return doc
}

func TestPuregoConfigDataBindings(t *testing.T) {
	ensurePuregoInitialized(t)

	config := NewConfigData()
	if config == nil {
		t.Fatal("NewConfigData() returned nil")
	}
	defer config.Destroy()

	if err := config.SetMemoryLimit(64 * 1024 * 1024); err != nil {
		t.Fatalf("SetMemoryLimit() failed: %v", err)
	}
	if got := config.GetMemoryLimit(); got != 64*1024*1024 {
		t.Fatalf("GetMemoryLimit() = %d, want %d", got, uint64(64*1024*1024))
	}
	if err := config.SetQueryThreadCount(2); err != nil {
		t.Fatalf("SetQueryThreadCount() failed: %v", err)
	}
	if got := config.GetQueryThreadCount(); got != 2 {
		t.Fatalf("GetQueryThreadCount() = %d, want 2", got)
	}
	if err := config.SetOptimizeThreadCount(1); err != nil {
		t.Fatalf("SetOptimizeThreadCount() failed: %v", err)
	}
	if got := config.GetOptimizeThreadCount(); got != 1 {
		t.Fatalf("GetOptimizeThreadCount() = %d, want 1", got)
	}
	if err := config.SetFTSBruteForceByKeysRatio(0.25); err != nil {
		t.Fatalf("SetFTSBruteForceByKeysRatio() failed: %v", err)
	}
	if got := config.GetFTSBruteForceByKeysRatio(); got != 0.25 {
		t.Fatalf("GetFTSBruteForceByKeysRatio() = %v, want 0.25", got)
	}

	dictDir := filepath.Join(t.TempDir(), "jieba")
	if err := config.SetJiebaDictDir(dictDir); err != nil {
		t.Fatalf("SetJiebaDictDir() failed: %v", err)
	}
	if got := config.GetJiebaDictDir(); got != dictDir {
		t.Fatalf("GetJiebaDictDir() = %q, want %q", got, dictDir)
	}

	oldDefault := GetDefaultJiebaDictDir()
	t.Cleanup(func() { SetDefaultJiebaDictDir(oldDefault) })
	SetDefaultJiebaDictDir(dictDir)
	if got := GetDefaultJiebaDictDir(); got != dictDir {
		t.Fatalf("GetDefaultJiebaDictDir() = %q, want %q", got, dictDir)
	}

	if err := config.SetConsoleLog(LogLevelInfo); err != nil {
		t.Fatalf("SetConsoleLog() failed: %v", err)
	}
	if err := config.SetFileLog(LogLevelWarn, t.TempDir(), "zvec-test", 1, 1); err != nil {
		t.Fatalf("SetFileLog() failed: %v", err)
	}
}

func TestPuregoCollectionManagementDMLBindings(t *testing.T) {
	ensurePuregoInitialized(t)

	schema := newPuregoIntegrationSchema(t, "purego_dml")
	defer schema.Destroy()

	options := NewCollectionOptions()
	if options == nil {
		t.Fatal("NewCollectionOptions() returned nil")
	}
	defer options.Destroy()
	if err := options.SetEnableMmap(false); err != nil {
		t.Fatalf("SetEnableMmap() failed: %v", err)
	}

	collection, err := CreateAndOpen(filepath.Join(t.TempDir(), "collection"), schema, options)
	if err != nil {
		t.Fatalf("CreateAndOpen() failed: %v", err)
	}
	defer func() { _ = collection.Close() }()

	gotSchema, err := collection.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() failed: %v", err)
	}
	if !gotSchema.HasField("embedding") {
		t.Fatal("GetSchema() missing embedding field")
	}
	gotSchema.Destroy()

	gotOptions, err := collection.GetOptions()
	if err != nil {
		t.Fatalf("GetOptions() failed: %v", err)
	}
	gotOptions.Destroy()

	docs := []*Doc{
		newPuregoIntegrationDoc(t, "doc1", "alpha", []float32{0.1, 0.2, 0.3, 0.4}),
		newPuregoIntegrationDoc(t, "doc2", "beta", []float32{0.2, 0.3, 0.4, 0.5}),
		newPuregoIntegrationDoc(t, "doc3", "gamma", []float32{0.3, 0.4, 0.5, 0.6}),
	}
	defer FreeDocs(docs)
	insertResult, err := collection.Insert(docs)
	if err != nil {
		t.Fatalf("Insert() failed: %v", err)
	}
	if insertResult.SuccessCount != 3 {
		t.Fatalf("Insert().SuccessCount = %d, want 3", insertResult.SuccessCount)
	}

	fetched, err := collection.Fetch([]string{"doc1", "doc2"}, &FetchOptions{OutputFields: []string{"id", "text"}})
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}
	defer FreeDocs(fetched)
	if len(fetched) != 2 {
		t.Fatalf("Fetch() returned %d docs, want 2", len(fetched))
	}

	updateDoc := newPuregoIntegrationDoc(t, "doc1", "alpha updated", []float32{0.9, 0.8, 0.7, 0.6})
	defer updateDoc.Destroy()
	updateResult, err := collection.Update([]*Doc{updateDoc})
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if updateResult.SuccessCount != 1 {
		t.Fatalf("Update().SuccessCount = %d, want 1", updateResult.SuccessCount)
	}

	updated, err := collection.Fetch([]string{"doc1"}, &FetchOptions{OutputFields: []string{"text"}})
	if err != nil {
		t.Fatalf("Fetch(updated) failed: %v", err)
	}
	defer FreeDocs(updated)
	if len(updated) != 1 {
		t.Fatalf("Fetch(updated) returned %d docs, want 1", len(updated))
	}
	text, err := updated[0].GetStringField("text")
	if err != nil {
		t.Fatalf("GetStringField(text) failed: %v", err)
	}
	if text != "alpha updated" {
		t.Fatalf("updated text = %q, want %q", text, "alpha updated")
	}

	deleteResult, err := collection.Delete([]string{"doc3"})
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
	if deleteResult.SuccessCount != 1 {
		t.Fatalf("Delete().SuccessCount = %d, want 1", deleteResult.SuccessCount)
	}
	deleted, err := collection.Fetch([]string{"doc3"}, nil)
	if err != nil {
		t.Fatalf("Fetch(deleted) failed: %v", err)
	}
	if len(deleted) != 0 {
		FreeDocs(deleted)
		t.Fatalf("Fetch(deleted) returned %d docs, want 0", len(deleted))
	}

	if err := collection.DeleteByFilter("id = 'doc2'"); err != nil {
		t.Fatalf("DeleteByFilter() failed: %v", err)
	}
	stats, err := collection.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}
	if stats.DocCount != 1 {
		t.Fatalf("GetStats().DocCount = %d, want 1", stats.DocCount)
	}
}

func TestPuregoCollectionDDLBindings(t *testing.T) {
	ensurePuregoInitialized(t)

	schema := newPuregoIntegrationSchema(t, "purego_ddl")
	defer schema.Destroy()

	collection, err := CreateAndOpen(filepath.Join(t.TempDir(), "collection"), schema, nil)
	if err != nil {
		t.Fatalf("CreateAndOpen() failed: %v", err)
	}
	defer func() { _ = collection.Close() }()

	scoreField := NewFieldSchema("score", DataTypeFloat, true, 0)
	if scoreField == nil {
		t.Fatal("NewFieldSchema(score) returned nil")
	}
	defer scoreField.Destroy()
	if err := collection.AddColumn(scoreField, "0.5"); err != nil {
		t.Fatalf("AddColumn() failed: %v", err)
	}

	scoreIndex, err := NewInvertIndexParams(true, false)
	if err != nil {
		t.Fatalf("NewInvertIndexParams(score) failed: %v", err)
	}
	defer scoreIndex.Destroy()
	if err := collection.CreateIndex("score", scoreIndex); err != nil {
		t.Fatalf("CreateIndex() failed: %v", err)
	}
	if err := collection.Optimize(); err != nil {
		t.Fatalf("Optimize() failed: %v", err)
	}
	if err := collection.AlterColumn("score", "score_new", nil); err != nil {
		t.Fatalf("AlterColumn() failed: %v", err)
	}
	if err := collection.DropIndex("score_new"); err != nil {
		t.Fatalf("DropIndex() failed: %v", err)
	}
	if err := collection.DropColumn("score_new"); err != nil {
		t.Fatalf("DropColumn() failed: %v", err)
	}
}

func TestPuregoQueryBuilderBindings(t *testing.T) {
	ensurePuregoInitialized(t)

	groupBy := NewGroupBySearchQuery()
	if groupBy == nil {
		t.Fatal("NewGroupBySearchQuery() returned nil")
	}
	defer groupBy.Destroy()
	if err := groupBy.SetFieldName("embedding"); err != nil {
		t.Fatalf("GroupBy.SetFieldName() failed: %v", err)
	}
	if err := groupBy.SetGroupByFieldName("id"); err != nil {
		t.Fatalf("GroupBy.SetGroupByFieldName() failed: %v", err)
	}
	if err := groupBy.SetGroupCount(2); err != nil {
		t.Fatalf("GroupBy.SetGroupCount() failed: %v", err)
	}
	if err := groupBy.SetTopkPerGroup(1); err != nil {
		t.Fatalf("GroupBy.SetTopkPerGroup() failed: %v", err)
	}
	if err := groupBy.SetQueryVector([]float32{0.1, 0.2, 0.3, 0.4}); err != nil {
		t.Fatalf("GroupBy.SetQueryVector() failed: %v", err)
	}
	if err := groupBy.SetFilter("id != ''"); err != nil {
		t.Fatalf("GroupBy.SetFilter() failed: %v", err)
	}
	if err := groupBy.SetIncludeVector(true); err != nil {
		t.Fatalf("GroupBy.SetIncludeVector() failed: %v", err)
	}
	if err := groupBy.SetOutputFields([]string{"id", "text"}); err != nil {
		t.Fatalf("GroupBy.SetOutputFields() failed: %v", err)
	}
	hnsw := NewHNSWQueryParams(32, 0, false, false)
	if hnsw == nil {
		t.Fatal("NewHNSWQueryParams() returned nil")
	}
	if err := groupBy.SetHNSWParams(hnsw); err != nil {
		t.Fatalf("GroupBy.SetHNSWParams() failed: %v", err)
	}
	if hnsw.handle != nil {
		t.Fatal("GroupBy.SetHNSWParams() did not transfer ownership")
	}

	groupByIVF := NewGroupBySearchQuery()
	if groupByIVF == nil {
		t.Fatal("NewGroupBySearchQuery(IVF) returned nil")
	}
	defer groupByIVF.Destroy()
	ivf := NewIVFQueryParams(4, false, 1)
	if ivf == nil {
		t.Fatal("NewIVFQueryParams() returned nil")
	}
	if err := groupByIVF.SetIVFParams(ivf); err != nil {
		t.Fatalf("GroupBy.SetIVFParams() failed: %v", err)
	}
	if ivf.handle != nil {
		t.Fatal("GroupBy.SetIVFParams() did not transfer ownership")
	}

	groupByFlat := NewGroupBySearchQuery()
	if groupByFlat == nil {
		t.Fatal("NewGroupBySearchQuery(Flat) returned nil")
	}
	defer groupByFlat.Destroy()
	flat := NewFlatQueryParams(false, 1)
	if flat == nil {
		t.Fatal("NewFlatQueryParams() returned nil")
	}
	if err := groupByFlat.SetFlatParams(flat); err != nil {
		t.Fatalf("GroupBy.SetFlatParams() failed: %v", err)
	}
	if flat.handle != nil {
		t.Fatal("GroupBy.SetFlatParams() did not transfer ownership")
	}

	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery() returned nil")
	}
	defer sub.Destroy()
	if err := sub.SetFieldName("sparse_embedding"); err != nil {
		t.Fatalf("SubQuery.SetFieldName() failed: %v", err)
	}
	if err := sub.SetSparseVector([]uint32{0, 5, 10}, []float32{0.1, 0.5, 0.9}); err != nil {
		t.Fatalf("SubQuery.SetSparseVector() failed: %v", err)
	}
	if err := sub.SetSparseVector([]uint32{0}, []float32{}); err == nil {
		t.Fatal("SubQuery.SetSparseVector() mismatch returned nil error")
	}
}
