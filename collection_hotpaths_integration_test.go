//go:build integration

package zvec

import (
	"path/filepath"
	"testing"
)

func TestCollectionDeleteMultiplePrimaryKeys(test *testing.T) {
	schema := createTestSchema()
	defer schema.Destroy()
	collection, err := CreateAndOpen(filepath.Join(testTempDir(test), "collection"), schema, nil)
	if err != nil {
		test.Fatalf("CreateAndOpen() failed: %v", err)
	}
	defer func() { _ = collection.Close() }()

	keys := []string{"a", "longer-key", "another_key", "last"}
	inputs := make([]*Doc, len(keys))
	for index, key := range keys {
		inputs[index] = createTestDoc(key, "value", []float32{1, 2, 3, 4})
	}
	defer FreeDocs(inputs)
	result, err := collection.Insert(inputs)
	if err != nil || result.SuccessCount != uint64(len(keys)) {
		test.Fatalf("Insert() returned result=%v, err=%v", result, err)
	}
	result, err = collection.Delete(keys)
	if err != nil || result.SuccessCount != uint64(len(keys)) {
		test.Fatalf("Delete() returned result=%v, err=%v", result, err)
	}
	fetched, err := collection.Fetch(keys, nil)
	if err != nil || len(fetched) != 0 {
		FreeDocs(fetched)
		test.Fatalf("Fetch() after Delete() returned %d documents, err=%v", len(fetched), err)
	}
}

func TestCollectionQueryFetchResultOwnershipAndKeys(test *testing.T) {
	schema := createTestSchema()
	defer schema.Destroy()
	collection, err := CreateAndOpen(filepath.Join(testTempDir(test), "collection"), schema, nil)
	if err != nil {
		test.Fatalf("CreateAndOpen() failed: %v", err)
	}
	defer func() { _ = collection.Close() }()

	keys := []string{"short", "a_much_longer_primary_key", "doc_003"}
	inputs := make([]*Doc, len(keys))
	for index, key := range keys {
		inputs[index] = createTestDoc(key, "value", []float32{1, float32(index + 1), 2, 3})
	}
	defer FreeDocs(inputs)
	result, err := collection.Insert(inputs)
	if err != nil || result.ErrorCount != 0 {
		test.Fatalf("Insert() returned result=%v, err=%v", result, err)
	}
	if err := collection.Flush(); err != nil {
		test.Fatalf("Flush() failed: %v", err)
	}

	query := NewSearchQuery()
	if query == nil {
		test.Fatal("NewSearchQuery() returned nil")
	}
	defer query.Destroy()
	if err := query.SetFieldName("embedding"); err != nil {
		test.Fatalf("SetFieldName() failed: %v", err)
	}
	if err := query.SetTopK(len(keys)); err != nil {
		test.Fatalf("SetTopK() failed: %v", err)
	}
	if err := query.SetQueryVector([]float32{1, 1, 2, 3}); err != nil {
		test.Fatalf("SetQueryVector() failed: %v", err)
	}
	queried, err := collection.Query(query)
	if err != nil || len(queried) != len(keys) {
		FreeDocs(queried)
		test.Fatalf("Query() returned %d documents, err=%v", len(queried), err)
	}
	queried[0].Destroy()
	if queried[1].GetPK() == "" {
		test.Fatal("destroying one Query result invalidated another")
	}
	FreeDocs(queried)

	fetched, err := collection.Fetch([]string{keys[0], "missing", keys[1], keys[2]}, &FetchOptions{OutputFields: []string{"id"}, IncludeVector: true})
	if err != nil || len(fetched) != len(keys) {
		FreeDocs(fetched)
		test.Fatalf("Fetch() returned %d documents, err=%v", len(fetched), err)
	}
	seen := make(map[string]bool, len(keys))
	for _, document := range fetched {
		seen[document.GetPK()] = true
	}
	for _, key := range keys {
		if !seen[key] {
			test.Errorf("Fetch() omitted %q", key)
		}
	}
	fetched[0].Destroy()
	if fetched[1].GetPK() == "" {
		test.Fatal("destroying one Fetch result invalidated another")
	}
	FreeDocs(fetched)
}
