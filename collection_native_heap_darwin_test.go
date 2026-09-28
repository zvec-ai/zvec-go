//go:build integration && darwin && cgo && !purego

package zvec

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCollectionQueryFetchNativeHeapDoesNotGrowPerResult(test *testing.T) {
	schema := createTestSchema()
	defer schema.Destroy()
	collection, err := CreateAndOpen(filepath.Join(testTempDir(test), "collection"), schema, nil)
	if err != nil {
		test.Fatalf("CreateAndOpen() failed: %v", err)
	}
	defer func() { _ = collection.Close() }()

	primaryKeys := make([]string, 10)
	docs := make([]*Doc, len(primaryKeys))
	for index := range primaryKeys {
		primaryKeys[index] = fmt.Sprintf("doc_%d", index)
		docs[index] = createTestDoc(primaryKeys[index], "hello", []float32{1, float32(index + 1), 2, 3})
	}
	result, err := collection.Insert(docs)
	FreeDocs(docs)
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
	if err := query.SetTopK(len(primaryKeys)); err != nil {
		test.Fatalf("SetTopK() failed: %v", err)
	}
	if err := query.SetQueryVector([]float32{1, 1, 2, 3}); err != nil {
		test.Fatalf("SetQueryVector() failed: %v", err)
	}

	runQuery := func() {
		test.Helper()
		results, err := collection.Query(query)
		if err != nil {
			test.Fatalf("Query() failed: %v", err)
		}
		if len(results) != len(primaryKeys) {
			FreeDocs(results)
			test.Fatalf("Query() returned %d documents, want %d", len(results), len(primaryKeys))
		}
		FreeDocs(results)
	}
	runFetch := func() {
		test.Helper()
		results, err := collection.Fetch(primaryKeys, nil)
		if err != nil {
			test.Fatalf("Fetch() failed: %v", err)
		}
		if len(results) != len(primaryKeys) {
			FreeDocs(results)
			test.Fatalf("Fetch() returned %d documents, want %d", len(results), len(primaryKeys))
		}
		FreeDocs(results)
	}

	for iteration := 0; iteration < 1000; iteration++ {
		runQuery()
		runFetch()
	}

	// Prime the allocator to steady state before measuring. The malloc-zone
	// in-use counter drifts upward during the first few thousand identical
	// call cycles as native heap fragmentation reaches its high-water mark —
	// that one-time drift scales with system memory pressure and is not
	// per-result growth. A second 20,000-call warmup window absorbs it; a
	// genuine per-result leak would still grow the measured windows below.
	for iteration := 0; iteration < 20000; iteration++ {
		runQuery()
	}
	runtime.GC()
	beforeQuery := nativeHeapInUse()
	if beforeQuery == 0 {
		test.Fatal("Darwin malloc zone statistics unavailable")
	}
	for iteration := 0; iteration < 20000; iteration++ {
		runQuery()
	}
	runtime.GC()
	afterQuery := nativeHeapInUse()
	for iteration := 0; iteration < 20000; iteration++ {
		runFetch()
	}
	runtime.GC()
	afterFetch := nativeHeapInUse()

	// Steady-state tolerance: after the warmup windows above, residual drift
	// comes from malloc-zone accounting noise (observed up to ~0.9MB under
	// system memory pressure), while any realistic per-result leak dwarfs
	// this — 50 bytes per freed ten-result call would already grow a
	// measured window by 10MB.
	const maxNativeHeapGrowth = 2_000_000
	queryGrowth := int64(afterQuery) - int64(beforeQuery)
	fetchGrowth := int64(afterFetch) - int64(afterQuery)
	test.Logf("native heap growth after 20,000 ten-result calls: Query %d B, Fetch %d B", queryGrowth, fetchGrowth)
	if queryGrowth > maxNativeHeapGrowth {
		test.Errorf("Query() native heap grew %d B after 20,000 freed ten-result calls", queryGrowth)
	}
	if fetchGrowth > maxNativeHeapGrowth {
		test.Errorf("Fetch() native heap grew %d B after 20,000 freed ten-result calls", fetchGrowth)
	}
}
