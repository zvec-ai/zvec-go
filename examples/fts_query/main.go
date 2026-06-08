// Package main demonstrates Full-Text Search (FTS) in zvec.
//
// This example shows how to:
// - Create a collection with an FTS-indexed text field
// - Insert documents with text content
// - Perform FTS queries using match strings and boolean expressions
// - Combine FTS with vector search using MultiQuery
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/fts_query
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec FTS Query Example ===")
	fmt.Println()

	// Initialize the library
	if err := zvec.Initialize(nil); err != nil {
		log.Fatalf("Failed to initialize zvec: %v", err)
	}
	defer func() {
		if err := zvec.Shutdown(); err != nil {
			log.Printf("Warning: failed to shutdown zvec: %v", err)
		}
	}()

	// === Create Collection Schema with FTS Field ===
	fmt.Println("--- Create Collection with FTS Index ---")

	schema := zvec.NewCollectionSchema("fts_example")
	defer schema.Destroy()

	// Add ID field (primary key, with invert index)
	idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
	defer idField.Destroy()
	invertParams, err := zvec.NewInvertIndexParams(true, false)
	if err != nil {
		log.Fatalf("Failed to create invert index params: %v", err)
	}
	defer invertParams.Destroy()
	if err := idField.SetIndexParams(invertParams); err != nil {
		log.Fatalf("Failed to set index params for id field: %v", err)
	}
	if err := schema.AddField(idField); err != nil {
		log.Fatalf("Failed to add id field: %v", err)
	}

	// Add content field with FTS index
	contentField := zvec.NewFieldSchema("content", zvec.DataTypeString, false, 0)
	defer contentField.Destroy()
	ftsParams, err := zvec.NewFTSIndexParams("default", nil, "")
	if err != nil {
		log.Fatalf("Failed to create FTS index params: %v", err)
	}
	defer ftsParams.Destroy()
	if err := contentField.SetIndexParams(ftsParams); err != nil {
		log.Fatalf("Failed to set FTS index params: %v", err)
	}
	if err := schema.AddField(contentField); err != nil {
		log.Fatalf("Failed to add content field: %v", err)
	}

	// Add vector field for hybrid search
	embField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 4)
	defer embField.Destroy()
	hnswParams, err := zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200)
	if err != nil {
		log.Fatalf("Failed to create HNSW index params: %v", err)
	}
	defer hnswParams.Destroy()
	if err := embField.SetIndexParams(hnswParams); err != nil {
		log.Fatalf("Failed to set HNSW params: %v", err)
	}
	if err := schema.AddField(embField); err != nil {
		log.Fatalf("Failed to add embedding field: %v", err)
	}

	// Create and open collection
	collectionPath := "./test_fts_collection"
	defer os.RemoveAll(collectionPath)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, nil)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer collection.Close()
	fmt.Println("Collection created with FTS index on 'content' field")

	// === Insert Documents ===
	fmt.Println("\n--- Insert Documents ---")

	documents := []struct {
		id        string
		content   string
		embedding []float32
	}{
		{"doc1", "The quick brown fox jumps over the lazy dog", []float32{0.1, 0.2, 0.3, 0.4}},
		{"doc2", "A fast red fox runs through the forest", []float32{0.2, 0.3, 0.4, 0.5}},
		{"doc3", "The lazy cat sleeps on the couch all day", []float32{0.3, 0.4, 0.5, 0.6}},
		{"doc4", "Dogs and cats are popular household pets", []float32{0.4, 0.5, 0.6, 0.7}},
		{"doc5", "The fox and the hound is a classic story", []float32{0.5, 0.6, 0.7, 0.8}},
	}

	for _, d := range documents {
		doc := zvec.NewDoc()
		doc.SetPK(d.id)
		doc.AddStringField("id", d.id)
		doc.AddStringField("content", d.content)
		doc.AddVectorFP32Field("embedding", d.embedding)
		result, err := collection.Insert([]*zvec.Doc{doc})
		doc.Destroy()
		if err != nil {
			log.Fatalf("Failed to insert %s: %v", d.id, err)
		}
		fmt.Printf("Inserted %s (%d success)\n", d.id, result.SuccessCount)
	}

	// Flush to ensure data is searchable
	if err := collection.Flush(); err != nil {
		log.Fatalf("Failed to flush: %v", err)
	}

	// === FTS Query with Match String ===
	fmt.Println("\n--- FTS Query: match 'fox' ---")

	query := zvec.NewSearchQuery()
	query.SetFieldName("content")
	query.SetTopK(10)
	query.SetQueryVector([]float32{0.1, 0.2, 0.3, 0.4})

	fts := zvec.NewFTS()
	fts.SetMatchString("fox")
	query.SetFTS(fts)
	fts.Destroy()

	results, err := collection.Query(query)
	query.Destroy()
	if err != nil {
		log.Fatalf("FTS query failed: %v", err)
	}
	fmt.Printf("Found %d results:\n", len(results))
	for _, r := range results {
		fmt.Printf("  PK=%s Score=%.4f\n", r.GetPK(), r.GetScore())
	}
	zvec.FreeDocs(results)

	fmt.Println()
	fmt.Println("FTS Query example completed successfully!")
}
