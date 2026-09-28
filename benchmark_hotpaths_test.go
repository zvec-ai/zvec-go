//go:build integration

package zvec

import (
	"fmt"
	"path/filepath"
	"testing"
)

func benchmarkHotpathCollection(benchmark *testing.B, dimension, count int) (*Collection, []string) {
	benchmark.Helper()
	schema := benchmarkCreateSchema(uint32(dimension))
	benchmark.Cleanup(schema.Destroy)
	collection, err := CreateAndOpen(filepath.Join(benchmark.TempDir(), "collection"), schema, nil)
	if err != nil {
		benchmark.Fatalf("CreateAndOpen() failed: %v", err)
	}
	benchmark.Cleanup(func() { _ = collection.Close() })

	keys := make([]string, count)
	docs := make([]*Doc, count)
	for index := range docs {
		keys[index] = fmt.Sprintf("doc_%d", index)
		document := NewDoc()
		if document == nil {
			FreeDocs(docs)
			benchmark.Fatal("NewDoc() returned nil")
		}
		docs[index] = document
		document.SetPK(keys[index])
		if err := document.AddStringField("id", keys[index]); err != nil {
			FreeDocs(docs)
			benchmark.Fatalf("AddStringField() failed: %v", err)
		}
		vector := generateRandomVector(dimension)
		vector[0] = float32(index%16) * 0.01
		if err := document.AddVectorFP32Field("embedding", vector); err != nil {
			FreeDocs(docs)
			benchmark.Fatalf("AddVectorFP32Field() failed: %v", err)
		}
	}
	result, err := collection.Insert(docs)
	FreeDocs(docs)
	if err != nil || result.ErrorCount != 0 {
		benchmark.Fatalf("Insert() returned result=%v, err=%v", result, err)
	}
	if err := collection.Flush(); err != nil {
		benchmark.Fatalf("Flush() failed: %v", err)
	}
	return collection, keys
}

func BenchmarkCollectionQueryHotpath(benchmark *testing.B) {
	for _, dimension := range []int{128, 768} {
		benchmark.Run(fmt.Sprintf("Dim%d", dimension), func(dimensionBenchmark *testing.B) {
			collection, _ := benchmarkHotpathCollection(dimensionBenchmark, dimension, 256)
			for _, topK := range []int{1, 10, 100} {
				dimensionBenchmark.Run(fmt.Sprintf("TopK%d", topK), func(queryBenchmark *testing.B) {
					query := NewSearchQuery()
					if query == nil {
						queryBenchmark.Fatal("NewSearchQuery() returned nil")
					}
					defer query.Destroy()
					if err := query.SetFieldName("embedding"); err != nil {
						queryBenchmark.Fatalf("SetFieldName() failed: %v", err)
					}
					if err := query.SetTopK(topK); err != nil {
						queryBenchmark.Fatalf("SetTopK() failed: %v", err)
					}
					if err := query.SetQueryVector(generateRandomVector(dimension)); err != nil {
						queryBenchmark.Fatalf("SetQueryVector() failed: %v", err)
					}
					for warmup := 0; warmup < 16; warmup++ {
						docs, err := collection.Query(query)
						if err != nil {
							queryBenchmark.Fatalf("Query() warmup failed: %v", err)
						}
						FreeDocs(docs)
					}
					queryBenchmark.ReportAllocs()
					queryBenchmark.ResetTimer()
					for iteration := 0; iteration < queryBenchmark.N; iteration++ {
						docs, err := collection.Query(query)
						if err != nil {
							queryBenchmark.Fatalf("Query() failed: %v", err)
						}
						FreeDocs(docs)
					}
				})
			}
		})
	}
}

func BenchmarkCollectionFetchHotpath(benchmark *testing.B) {
	collection, insertedKeys := benchmarkHotpathCollection(benchmark, 128, 256)
	keys := make([]string, 1000)
	for index := range keys {
		keys[index] = insertedKeys[index%len(insertedKeys)]
	}
	for _, count := range []int{1, 10, 100, 1000} {
		benchmark.Run(fmt.Sprintf("Keys%d", count), func(fetchBenchmark *testing.B) {
			selected := keys[:count]
			for warmup := 0; warmup < 16; warmup++ {
				docs, err := collection.Fetch(selected, nil)
				if err != nil {
					fetchBenchmark.Fatalf("Fetch() warmup failed: %v", err)
				}
				FreeDocs(docs)
			}
			fetchBenchmark.ReportAllocs()
			fetchBenchmark.ResetTimer()
			for iteration := 0; iteration < fetchBenchmark.N; iteration++ {
				docs, err := collection.Fetch(selected, nil)
				if err != nil {
					fetchBenchmark.Fatalf("Fetch() failed: %v", err)
				}
				FreeDocs(docs)
			}
		})
	}
}

func BenchmarkCollectionQueryPhases(benchmark *testing.B) {
	collection, _ := benchmarkHotpathCollection(benchmark, 128, 256)
	query := NewSearchQuery()
	if query == nil {
		benchmark.Fatal("NewSearchQuery() returned nil")
	}
	benchmark.Cleanup(query.Destroy)
	if err := query.SetFieldName("embedding"); err != nil {
		benchmark.Fatalf("SetFieldName() failed: %v", err)
	}
	if err := query.SetTopK(100); err != nil {
		benchmark.Fatalf("SetTopK() failed: %v", err)
	}
	if err := query.SetQueryVector(generateRandomVector(128)); err != nil {
		benchmark.Fatalf("SetQueryVector() failed: %v", err)
	}

	benchmark.Run("QueryOnly", func(phaseBenchmark *testing.B) {
		phaseBenchmark.ReportAllocs()
		for iteration := 0; iteration < phaseBenchmark.N; iteration++ {
			docs, err := collection.Query(query)
			if err != nil {
				phaseBenchmark.Fatalf("Query() failed: %v", err)
			}
			phaseBenchmark.StopTimer()
			FreeDocs(docs)
			phaseBenchmark.StartTimer()
		}
	})
	benchmark.Run("FreeDocsOnly", func(phaseBenchmark *testing.B) {
		phaseBenchmark.ReportAllocs()
		for iteration := 0; iteration < phaseBenchmark.N; iteration++ {
			phaseBenchmark.StopTimer()
			docs, err := collection.Query(query)
			if err != nil {
				phaseBenchmark.Fatalf("Query() failed: %v", err)
			}
			phaseBenchmark.StartTimer()
			FreeDocs(docs)
		}
	})
}

func BenchmarkCollectionFetchUniqueKeys(benchmark *testing.B) {
	collection, keys := benchmarkHotpathCollection(benchmark, 128, 1000)
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for iteration := 0; iteration < benchmark.N; iteration++ {
		docs, err := collection.Fetch(keys, nil)
		if err != nil {
			benchmark.Fatalf("Fetch() failed: %v", err)
		}
		if len(docs) != len(keys) {
			FreeDocs(docs)
			benchmark.Fatalf("Fetch() returned %d documents, want %d", len(docs), len(keys))
		}
		FreeDocs(docs)
	}
}

func BenchmarkCollectionDeleteKeys(benchmark *testing.B) {
	for _, count := range []int{1, 10, 100, 1000} {
		benchmark.Run(fmt.Sprintf("Keys%d", count), func(deleteBenchmark *testing.B) {
			collection, keys := benchmarkHotpathCollection(deleteBenchmark, 128, count)
			vector := make([]float32, 128)
			docs := make([]*Doc, count)
			for index, key := range keys {
				doc := NewDoc()
				if doc == nil {
					FreeDocs(docs)
					deleteBenchmark.Fatal("NewDoc() returned nil")
				}
				docs[index] = doc
				doc.SetPK(key)
				if err := doc.AddStringField("id", key); err != nil {
					FreeDocs(docs)
					deleteBenchmark.Fatalf("AddStringField() failed: %v", err)
				}
				if err := doc.AddVectorFP32Field("embedding", vector); err != nil {
					FreeDocs(docs)
					deleteBenchmark.Fatalf("AddVectorFP32Field() failed: %v", err)
				}
			}
			deleteBenchmark.Cleanup(func() { FreeDocs(docs) })
			deleteBenchmark.ReportAllocs()
			deleteBenchmark.ResetTimer()
			for iteration := 0; iteration < deleteBenchmark.N; iteration++ {
				if iteration > 0 {
					deleteBenchmark.StopTimer()
					result, err := collection.Insert(docs)
					if err != nil || result.ErrorCount != 0 {
						deleteBenchmark.Fatalf("Insert() returned result=%v, err=%v", result, err)
					}
					deleteBenchmark.StartTimer()
				}
				result, err := collection.Delete(keys)
				if err != nil || result.SuccessCount != uint64(count) {
					deleteBenchmark.Fatalf("Delete() returned result=%v, err=%v", result, err)
				}
			}
		})
	}
}
