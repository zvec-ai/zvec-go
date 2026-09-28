//go:build integration

package zvec

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func BenchmarkFTSIngestion(benchmark *testing.B) {
	rows := 2048
	if raw := os.Getenv("ZVEC_BENCH_FTS_ROWS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			benchmark.Fatalf("invalid ZVEC_BENCH_FTS_ROWS=%q", raw)
		}
		rows = parsed
	}
	text := strings.Repeat("the quick brown fox jumps over a lazy dog and searches for text ", 16)
	corpus := os.Getenv("ZVEC_BENCH_FTS_CORPUS")
	if corpus == "" {
		corpus = "repeat"
	}
	if corpus != "repeat" && corpus != "varied" {
		benchmark.Fatalf("invalid ZVEC_BENCH_FTS_CORPUS=%q", corpus)
	}
	var bufferMiB uint64
	if raw := os.Getenv("ZVEC_BENCH_FTS_BUFFER_MIB"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || parsed == 0 || parsed > 4095 {
			benchmark.Fatalf("invalid ZVEC_BENCH_FTS_BUFFER_MIB=%q", raw)
		}
		bufferMiB = parsed
	}
	vector := []float32{0, 0, 0, 0}
	variants := []struct {
		name       string
		batchSize  int
		ftsAtStart bool
		ftsAfter   bool
		extra      string
	}{
		{"Plain_Batch512", 512, false, false, ""},
		{"FTS_Batch1", 1, true, false, ""},
		{"FTS_Batch128", 128, true, false, ""},
		{"FTS_Batch512", 512, true, false, ""},
		{"FTS_After_Batch512", 512, false, true, ""},
		{"FTS_NoPos_Batch1", 1, true, false, `{"store_positions":false}`},
		{"FTS_NoPos_Batch128", 128, true, false, `{"store_positions":false}`},
		{"FTS_NoPos_Batch512", 512, true, false, `{"store_positions":false}`},
		{"FTS_NoPos_After_Batch512", 512, false, true, `{"store_positions":false}`},
	}
	for _, variant := range variants {
		benchmark.Run(variant.name, func(caseBenchmark *testing.B) {
			caseBenchmark.ReportAllocs()
			for run := 0; run < caseBenchmark.N; run++ {
				benchmarkFTSIngestCase(caseBenchmark, rows, text, corpus, vector, bufferMiB, variant.batchSize, variant.ftsAtStart, variant.ftsAfter, variant.extra)
			}
		})
	}
}

func benchmarkFTSIngestCase(benchmark *testing.B, rows int, text, corpus string, vector []float32, bufferMiB uint64, batchSize int, ftsAtStart, ftsAfter bool, ftsExtra string) {
	benchmark.StopTimer()
	schema := NewCollectionSchema("fts_bench")
	if schema == nil {
		benchmark.Fatal("NewCollectionSchema() returned nil")
	}
	defer schema.Destroy()
	content := NewFieldSchema("content", DataTypeString, false, 0)
	if content == nil {
		benchmark.Fatal("NewFieldSchema(content) returned nil")
	}
	defer content.Destroy()
	if ftsAtStart {
		params := benchmarkFTSIndexParams(benchmark, ftsExtra)
		defer params.Destroy()
		if err := content.SetIndexParams(params); err != nil {
			benchmark.Fatalf("SetIndexParams(content) failed: %v", err)
		}
	}
	if err := schema.AddField(content); err != nil {
		benchmark.Fatalf("AddField(content) failed: %v", err)
	}
	embedding := NewFieldSchema("embedding", DataTypeVectorFP32, false, uint32(len(vector)))
	if embedding == nil {
		benchmark.Fatal("NewFieldSchema(embedding) returned nil")
	}
	defer embedding.Destroy()
	flat, err := NewFlatIndexParams(MetricTypeL2)
	if err != nil {
		benchmark.Fatalf("NewFlatIndexParams() failed: %v", err)
	}
	defer flat.Destroy()
	if err := embedding.SetIndexParams(flat); err != nil {
		benchmark.Fatalf("SetIndexParams(embedding) failed: %v", err)
	}
	if err := schema.AddField(embedding); err != nil {
		benchmark.Fatalf("AddField(embedding) failed: %v", err)
	}
	path := filepath.Join(benchmark.TempDir(), "collection")
	var options *CollectionOptions
	if bufferMiB != 0 {
		options = NewCollectionOptions()
		if options == nil {
			benchmark.Fatal("NewCollectionOptions() returned nil")
		}
		defer options.Destroy()
		if err := options.SetMaxBufferSize(bufferMiB << 20); err != nil {
			benchmark.Fatalf("SetMaxBufferSize() failed: %v", err)
		}
	}
	collection, err := CreateAndOpen(path, schema, options)
	if err != nil {
		benchmark.Fatalf("CreateAndOpen() failed: %v", err)
	}
	defer func() { _ = collection.Close() }()

	var prepareTime, insertTime time.Duration
	benchmark.StartTimer()
	for offset := 0; offset < rows; offset += batchSize {
		started := time.Now()
		end := min(offset+batchSize, rows)
		docs := make([]*Doc, end-offset)
		for index := range docs {
			key := fmt.Sprintf("doc_%d", offset+index)
			doc := NewDoc()
			if doc == nil {
				FreeDocs(docs)
				benchmark.Fatal("NewDoc() returned nil")
			}
			docs[index] = doc
			doc.SetPK(key)
			content := text
			if corpus == "varied" {
				content = benchmarkVariedFTSText(offset + index)
			}
			if err := doc.AddStringField("content", content); err != nil {
				FreeDocs(docs)
				benchmark.Fatalf("AddStringField() failed: %v", err)
			}
			if err := doc.AddVectorFP32Field("embedding", vector); err != nil {
				FreeDocs(docs)
				benchmark.Fatalf("AddVectorFP32Field() failed: %v", err)
			}
		}
		prepareTime += time.Since(started)
		started = time.Now()
		result, err := collection.Insert(docs)
		insertTime += time.Since(started)
		FreeDocs(docs)
		if err != nil || result.SuccessCount != uint64(len(docs)) {
			benchmark.Fatalf("Insert() returned result=%v, err=%v", result, err)
		}
	}
	benchmark.StopTimer()
	started := time.Now()
	if err := collection.Flush(); err != nil {
		benchmark.Fatalf("Flush() failed: %v", err)
	}
	flushTime := time.Since(started)
	flushBytes, flushDirs := benchmarkFTSDiskUsage(benchmark, path)
	var indexTime time.Duration
	if ftsAfter {
		params := benchmarkFTSIndexParams(benchmark, ftsExtra)
		defer params.Destroy()
		started = time.Now()
		if err := collection.CreateIndex("content", params); err != nil {
			benchmark.Fatalf("CreateIndex(content) failed: %v", err)
		}
		indexTime = time.Since(started)
	}
	started = time.Now()
	if err := collection.Optimize(); err != nil {
		benchmark.Fatalf("Optimize() failed: %v", err)
	}
	optimizeTime := time.Since(started)
	optimizedBytes, optimizedDirs := benchmarkFTSDiskUsage(benchmark, path)
	if ftsAtStart || ftsAfter {
		benchmarkFTSQuery(benchmark, collection)
	}
	if err := collection.Close(); err != nil {
		benchmark.Fatalf("Close() failed: %v", err)
	}
	closedBytes, closedDirs := benchmarkFTSDiskUsage(benchmark, path)
	collection, err = Open(path, options)
	if err != nil {
		benchmark.Fatalf("Open() failed: %v", err)
	}
	if ftsAtStart || ftsAfter {
		benchmarkFTSQuery(benchmark, collection)
	}
	reopenedBytes, reopenedDirs := benchmarkFTSDiskUsage(benchmark, path)
	benchmark.ReportMetric(float64(prepareTime.Milliseconds()), "prepare_ms")
	benchmark.ReportMetric(float64(insertTime.Milliseconds()), "insert_ms")
	benchmark.ReportMetric(float64(flushTime.Milliseconds()), "flush_ms")
	benchmark.ReportMetric(float64(indexTime.Milliseconds()), "index_ms")
	benchmark.ReportMetric(float64(optimizeTime.Milliseconds()), "optimize_ms")
	benchmark.ReportMetric(float64(flushBytes), "flush_disk_B")
	benchmark.ReportMetric(float64(optimizedBytes), "optimized_disk_B")
	benchmark.ReportMetric(float64(closedBytes), "closed_disk_B")
	benchmark.ReportMetric(float64(reopenedBytes), "reopened_disk_B")
	benchmark.Logf("rows=%d text_bytes=%d corpus=%s batch=%d buffer_MiB=%d; flush dirs=%v; optimized dirs=%v; closed dirs=%v; reopened dirs=%v", rows, rows*len(text), corpus, batchSize, bufferMiB, flushDirs, optimizedDirs, closedDirs, reopenedDirs)
}

func benchmarkVariedFTSText(row int) string {
	var text strings.Builder
	text.Grow(1024)
	text.WriteString("fox ")
	for word := 0; word < 102; word++ {
		fmt.Fprintf(&text, "word%05d ", (row*131+word*1973)%65536)
	}
	return text.String()
}

func benchmarkFTSIndexParams(benchmark *testing.B, extraParams string) *IndexParams {
	params, err := NewFTSIndexParams("whitespace", []string{"lowercase"}, extraParams)
	if err != nil {
		benchmark.Fatalf("NewFTSIndexParams() failed: %v", err)
	}
	return params
}

func benchmarkFTSQuery(benchmark *testing.B, collection *Collection) {
	query := NewSearchQuery()
	if query == nil {
		benchmark.Fatal("NewSearchQuery() returned nil")
	}
	defer query.Destroy()
	if err := query.SetFieldName("content"); err != nil {
		benchmark.Fatalf("SetFieldName() failed: %v", err)
	}
	if err := query.SetTopK(10); err != nil {
		benchmark.Fatalf("SetTopK() failed: %v", err)
	}
	fts := NewFTS()
	if fts == nil {
		benchmark.Fatal("NewFTS() returned nil")
	}
	defer fts.Destroy()
	if err := fts.SetMatchString("fox"); err != nil {
		benchmark.Fatalf("SetMatchString() failed: %v", err)
	}
	if err := query.SetFTS(fts); err != nil {
		benchmark.Fatalf("SetFTS() failed: %v", err)
	}
	docs, err := collection.Query(query)
	if err != nil || len(docs) == 0 {
		FreeDocs(docs)
		benchmark.Fatalf("FTS Query() returned %d documents, err=%v", len(docs), err)
	}
	FreeDocs(docs)
}
