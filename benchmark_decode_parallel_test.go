//go:build integration

package zvec

import (
	"fmt"
	"testing"
)

// The benchmarks below close the measurement gaps left by the
// QueryOnly/FreeDocsOnly phase split in benchmark_hotpaths_test.go: the
// user-visible decode half of a search round trip (reading PK, score and
// vector off every result doc), the single-transition FFI floor, concurrent
// query throughput, and the batched update/upsert write paths.

// BenchmarkCgoCallFloor_GetScore measures the irreducible cost of ONE FFI
// transition: GetScore is the cheapest native call in the SDK (single
// transition, no argument marshaling, no native allocation). Divide any
// other benchmark's transition count by this floor to see its unavoidable
// FFI share: ns/op ≈ transitions × floor + native work.
func BenchmarkCgoCallFloor_GetScore(benchmark *testing.B) {
	doc := NewDoc()
	if doc == nil {
		benchmark.Fatal("NewDoc() returned nil")
	}
	benchmark.Cleanup(doc.Destroy)
	doc.SetScore(1.5)

	benchmark.ReportAllocs()
	for warmup := 0; warmup < 32; warmup++ {
		_ = doc.GetScore()
	}
	benchmark.ResetTimer()
	for iteration := 0; iteration < benchmark.N; iteration++ {
		_ = doc.GetScore()
	}
}

// BenchmarkCgoCallFloor_IsInitialized is a handle-free variant of the FFI
// transition floor probe.
func BenchmarkCgoCallFloor_IsInitialized(benchmark *testing.B) {
	benchmark.ReportAllocs()
	for warmup := 0; warmup < 32; warmup++ {
		_ = IsInitialized()
	}
	benchmark.ResetTimer()
	for iteration := 0; iteration < benchmark.N; iteration++ {
		_ = IsInitialized()
	}
}

// benchmarkDecodeResultDoc reads every per-result value a real consumer of
// a search wants: the PK, the score, and the vector (into a shared buffer,
// so vector decoding itself is allocation-free).
func benchmarkDecodeResultDoc(benchmark *testing.B, doc *Doc, buffer []float32) {
	if doc.GetPK() == "" {
		benchmark.Fatal("GetPK() returned empty primary key")
	}
	_ = doc.GetScore()
	if _, err := doc.GetVectorFP32FieldInto("embedding", buffer); err != nil {
		benchmark.Fatalf("GetVectorFP32FieldInto() failed: %v", err)
	}
}

// BenchmarkCollectionQueryResultDecode measures the full user-visible
// search round trip: Query + per-doc decode (PK, score, vector into a
// shared buffer) + FreeDocs, all inside the timed region. Comparing this
// against the QueryOnly phase reveals how much of end-to-end search
// latency the decode path costs.
func BenchmarkCollectionQueryResultDecode(benchmark *testing.B) {
	const dimension = 128
	collection, _ := benchmarkHotpathCollection(benchmark, dimension, 256)

	decodeBuffer := make([]float32, dimension)
	for _, topK := range []int{10, 100} {
		benchmark.Run(fmt.Sprintf("TopK%d", topK), func(decodeBenchmark *testing.B) {
			query := NewSearchQuery()
			if query == nil {
				decodeBenchmark.Fatal("NewSearchQuery() returned nil")
			}
			decodeBenchmark.Cleanup(query.Destroy)
			if err := query.SetFieldName("embedding"); err != nil {
				decodeBenchmark.Fatalf("SetFieldName() failed: %v", err)
			}
			if err := query.SetTopK(topK); err != nil {
				decodeBenchmark.Fatalf("SetTopK() failed: %v", err)
			}
			if err := query.SetIncludeVector(true); err != nil {
				decodeBenchmark.Fatalf("SetIncludeVector() failed: %v", err)
			}
			if err := query.SetQueryVector(generateRandomVector(dimension)); err != nil {
				decodeBenchmark.Fatalf("SetQueryVector() failed: %v", err)
			}

			decodeDocs := func() {
				docs, err := collection.Query(query)
				if err != nil {
					decodeBenchmark.Fatalf("Query() failed: %v", err)
				}
				if len(docs) != topK {
					FreeDocs(docs)
					decodeBenchmark.Fatalf("Query() returned %d documents, want %d", len(docs), topK)
				}
				for _, doc := range docs {
					benchmarkDecodeResultDoc(decodeBenchmark, doc, decodeBuffer)
				}
				FreeDocs(docs)
			}
			for warmup := 0; warmup < 16; warmup++ {
				decodeDocs()
			}

			decodeBenchmark.ReportAllocs()
			decodeBenchmark.ResetTimer()
			for iteration := 0; iteration < decodeBenchmark.N; iteration++ {
				decodeDocs()
			}
			decodeBenchmark.ReportMetric(
				float64(decodeBenchmark.Elapsed().Nanoseconds())/float64(decodeBenchmark.N)/float64(topK),
				"ns/doc")
		})
	}
}

// benchmarkWriteBatchDocs rebuilds a batch of docs with the given primary
// keys so every iteration writes fresh document values over a stable set
// of keys; callers exclude it from the timed region via StopTimer.
func benchmarkWriteBatchDocs(benchmark *testing.B, keys []string) []*Doc {
	benchmark.Helper()
	docs := make([]*Doc, len(keys))
	vector := generateRandomVector(128)
	for index, key := range keys {
		doc := NewDoc()
		if doc == nil {
			FreeDocs(docs)
			benchmark.Fatal("NewDoc() returned nil")
		}
		doc.SetPK(key)
		if err := doc.AddStringField("id", key); err != nil {
			FreeDocs(docs)
			benchmark.Fatalf("AddStringField() failed: %v", err)
		}
		if err := doc.AddVectorFP32Field("embedding", vector); err != nil {
			FreeDocs(docs)
			benchmark.Fatalf("AddVectorFP32Field() failed: %v", err)
		}
		docs[index] = doc
	}
	return docs
}

const benchmarkWriteBatchSize = 100

// benchmarkWriteCollectionOp drives a batched write operation against a
// fixed 256-doc corpus: documents for the same 100 keys are rebuilt outside
// the timed region, so ns/op isolates the batched FFI call plus native
// write work (reported as ns/key).
func benchmarkWriteCollectionOp(benchmark *testing.B, name string, op func(docs []*Doc) error, collection *Collection, keys []string) {
	benchmark.Helper()
	batch := keys[:benchmarkWriteBatchSize]

	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for iteration := 0; iteration < benchmark.N; iteration++ {
		benchmark.StopTimer()
		docs := benchmarkWriteBatchDocs(benchmark, batch)
		benchmark.StartTimer()

		if err := op(docs); err != nil {
			benchmark.Fatalf("%s() failed: %v", name, err)
		}

		benchmark.StopTimer()
		FreeDocs(docs)
		benchmark.StartTimer()
	}
	benchmark.ReportMetric(
		float64(benchmark.Elapsed().Nanoseconds())/float64(benchmark.N)/float64(benchmarkWriteBatchSize),
		"ns/key")
}

// BenchmarkCollectionUpdateBatch measures batched Update against a fixed
// 256-doc corpus.
func BenchmarkCollectionUpdateBatch(benchmark *testing.B) {
	collection, keys := benchmarkHotpathCollection(benchmark, 128, 256)
	benchmarkWriteCollectionOp(benchmark, "Update", func(docs []*Doc) error {
		_, err := collection.Update(docs)
		return err
	}, collection, keys)
}

// BenchmarkCollectionUpsertBatch measures batched Upsert against a fixed
// 256-doc corpus.
func BenchmarkCollectionUpsertBatch(benchmark *testing.B) {
	collection, keys := benchmarkHotpathCollection(benchmark, 128, 256)
	benchmarkWriteCollectionOp(benchmark, "Upsert", func(docs []*Doc) error {
		_, err := collection.Upsert(docs)
		return err
	}, collection, keys)
}

// BenchmarkCollectionQueryParallel exercises concurrent query throughput
// against one shared collection — the shape of an in-process vector
// service (each goroutine issues its own query object, per-request
// style). Nothing else in the package measures the SDK under goroutine
// concurrency.
func BenchmarkCollectionQueryParallel(benchmark *testing.B) {
	const dimension = 128
	collection, _ := benchmarkHotpathCollection(benchmark, dimension, 256)

	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	benchmark.RunParallel(func(pb *testing.PB) {
		query := NewSearchQuery()
		if query == nil {
			benchmark.Errorf("NewSearchQuery() returned nil")
			return
		}
		defer query.Destroy()
		if err := query.SetFieldName("embedding"); err != nil {
			benchmark.Errorf("SetFieldName() failed: %v", err)
			return
		}
		if err := query.SetTopK(10); err != nil {
			benchmark.Errorf("SetTopK() failed: %v", err)
			return
		}
		if err := query.SetQueryVector(generateRandomVector(dimension)); err != nil {
			benchmark.Errorf("SetQueryVector() failed: %v", err)
			return
		}
		for pb.Next() {
			docs, err := collection.Query(query)
			if err != nil {
				benchmark.Errorf("Query() failed: %v", err)
				return
			}
			FreeDocs(docs)
		}
	})
}
