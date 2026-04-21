// Package main demonstrates vector query operations in zvec.
//
// This example shows how to:
// - Perform basic vector similarity search (TopK)
// - Set query vector and field name
// - Use filter expressions to narrow results
// - Configure output fields
// - Include vectors and document IDs in results
// - Use HNSW query parameters (ef, radius)
// - Use IVF query parameters (nprobe)
// - Use GroupByVectorQuery for grouping results
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/vector_query
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec Vector Query Example ===")
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

	// Create collection schema
	schema := zvec.NewCollectionSchema("vector_query_example")
	defer schema.Destroy()

	// Add ID field
	idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
	defer idField.Destroy()
	invertParams := zvec.NewInvertIndexParams(true, false)
	defer invertParams.Destroy()
	if err := idField.SetIndexParams(invertParams); err != nil {
		log.Fatalf("Failed to set index params for id field: %v", err)
	}
	if err := schema.AddField(idField); err != nil {
		log.Fatalf("Failed to add id field: %v", err)
	}

	// Add category field for filtering
	categoryField := zvec.NewFieldSchema("category", zvec.DataTypeString, true, 0)
	defer categoryField.Destroy()
	if err := categoryField.SetIndexParams(invertParams); err != nil {
		log.Fatalf("Failed to set index params for category field: %v", err)
	}
	if err := schema.AddField(categoryField); err != nil {
		log.Fatalf("Failed to add category field: %v", err)
	}

	// Add embedding field
	embeddingField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 128)
	defer embeddingField.Destroy()
	hnswParams := zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200)
	defer hnswParams.Destroy()
	if err := embeddingField.SetIndexParams(hnswParams); err != nil {
		log.Fatalf("Failed to set index params for embedding field: %v", err)
	}
	if err := schema.AddField(embeddingField); err != nil {
		log.Fatalf("Failed to add embedding field: %v", err)
	}

	// Create and open collection
	collectionPath := "./test_vector_query_collection"
	defer os.RemoveAll(collectionPath)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, nil)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer collection.Close()
	fmt.Println("✓ Collection created successfully")

	// Insert sample documents
	fmt.Println("\n--- Insert Sample Documents ---")
	docs := make([]*zvec.Doc, 5)
	for i := 0; i < 5; i++ {
		doc := zvec.NewDoc()
		defer doc.Destroy()
		doc.SetPK(fmt.Sprintf("doc%d", i+1))
		doc.AddStringField("id", fmt.Sprintf("doc%d", i+1))
		doc.AddStringField("category", []string{"electronics", "books", "clothing", "electronics", "books"}[i])

		// Create different vectors for each document
		vector := make([]float32, 128)
		for j := range vector {
			vector[j] = float32(i+1) * 0.1
		}
		doc.AddVectorFP32Field("embedding", vector)

		docs[i] = doc
	}

	result, err := collection.Insert(docs)
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}
	fmt.Printf("✓ Inserted %d documents\n", result.SuccessCount)

	// Flush to ensure data is indexed
	if err := collection.Flush(); err != nil {
		log.Fatalf("Failed to flush collection: %v", err)
	}

	// === Example 1: Basic Vector Query ===
	fmt.Println("\n--- Example 1: Basic Vector Query ---")
	query := zvec.NewVectorQuery()
	defer query.Destroy()
	query.SetFieldName("embedding")
	query.SetTopK(3)

	// Query vector similar to doc1
	queryVector := make([]float32, 128)
	for i := range queryVector {
		queryVector[i] = 0.1
	}
	query.SetQueryVector(queryVector)
	query.SetFilter("")
	query.SetIncludeVector(false)
	query.SetIncludeDocID(true)

	results, err := collection.Query(query)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer zvec.FreeDocs(results)

	fmt.Printf("✓ Query returned %d results\n", len(results))
	for i, doc := range results {
		pk := doc.GetPK()
		category, _ := doc.GetStringField("category")
		fmt.Printf("  %d. PK=%s, Category=%s, Score=%.4f\n", i+1, pk, category, doc.GetScore())
	}

	// === Example 2: Query with Filter ===
	fmt.Println("\n--- Example 2: Query with Filter ---")
	queryWithFilter := zvec.NewVectorQuery()
	defer queryWithFilter.Destroy()
	queryWithFilter.SetFieldName("embedding")
	queryWithFilter.SetTopK(10)
	queryWithFilter.SetQueryVector(queryVector)
	queryWithFilter.SetFilter("category == 'electronics'")
	queryWithFilter.SetIncludeVector(false)
	queryWithFilter.SetIncludeDocID(true)

	filteredResults, err := collection.Query(queryWithFilter)
	if err != nil {
		log.Fatalf("Failed to query with filter: %v", err)
	}
	defer zvec.FreeDocs(filteredResults)

	fmt.Printf("✓ Query with filter 'category == \"electronics\"' returned %d results\n", len(filteredResults))
	for i, doc := range filteredResults {
		pk := doc.GetPK()
		fmt.Printf("  %d. PK=%s, Score=%.4f\n", i+1, pk, doc.GetScore())
	}

	// === Example 3: Query with Output Fields ===
	fmt.Println("\n--- Example 3: Query with Output Fields ---")
	queryWithOutput := zvec.NewVectorQuery()
	defer queryWithOutput.Destroy()
	queryWithOutput.SetFieldName("embedding")
	queryWithOutput.SetTopK(3)
	queryWithOutput.SetQueryVector(queryVector)
	queryWithOutput.SetFilter("")
	queryWithOutput.SetIncludeVector(false)
	queryWithOutput.SetIncludeDocID(true)
	queryWithOutput.SetOutputFields([]string{"id", "category"})

	outputResults, err := collection.Query(queryWithOutput)
	if err != nil {
		log.Fatalf("Failed to query with output fields: %v", err)
	}
	defer zvec.FreeDocs(outputResults)

	fmt.Printf("✓ Query with output fields returned %d results\n", len(outputResults))
	for i, doc := range outputResults {
		pk := doc.GetPK()
		id, _ := doc.GetStringField("id")
		category, _ := doc.GetStringField("category")
		fmt.Printf("  %d. PK=%s, ID=%s, Category=%s, Score=%.4f\n", i+1, pk, id, category, doc.GetScore())
	}

	// === Example 4: Query with Include Vector ===
	fmt.Println("\n--- Example 4: Query with Include Vector ---")
	queryWithVector := zvec.NewVectorQuery()
	defer queryWithVector.Destroy()
	queryWithVector.SetFieldName("embedding")
	queryWithVector.SetTopK(2)
	queryWithVector.SetQueryVector(queryVector)
	queryWithVector.SetFilter("")
	queryWithVector.SetIncludeVector(true)
	queryWithVector.SetIncludeDocID(true)

	vectorResults, err := collection.Query(queryWithVector)
	if err != nil {
		log.Fatalf("Failed to query with include vector: %v", err)
	}
	defer zvec.FreeDocs(vectorResults)

	fmt.Printf("✓ Query with include vector returned %d results\n", len(vectorResults))
	for i, doc := range vectorResults {
		pk := doc.GetPK()
		vector, _ := doc.GetVectorFP32Field("embedding")
		fmt.Printf("  %d. PK=%s, Score=%.4f, Vector=[%.2f, %.2f, ...]\n", i+1, pk, doc.GetScore(), vector[0], vector[1])
	}

	// === Example 5: Query with HNSW Parameters ===
	fmt.Println("\n--- Example 5: Query with HNSW Parameters ---")
	queryWithHNSW := zvec.NewVectorQuery()
	defer queryWithHNSW.Destroy()
	queryWithHNSW.SetFieldName("embedding")
	queryWithHNSW.SetTopK(3)
	queryWithHNSW.SetQueryVector(queryVector)
	queryWithHNSW.SetFilter("")
	queryWithHNSW.SetIncludeVector(false)
	queryWithHNSW.SetIncludeDocID(true)

	// Set HNSW query parameters
	hnswQueryParams := zvec.NewHNSWQueryParams(100, 0.5, false, false)
	defer hnswQueryParams.Destroy()
	queryWithHNSW.SetHNSWParams(hnswQueryParams)

	hnswResults, err := collection.Query(queryWithHNSW)
	if err != nil {
		log.Fatalf("Failed to query with HNSW params: %v", err)
	}
	defer zvec.FreeDocs(hnswResults)

	fmt.Printf("✓ Query with HNSW params (ef=100, radius=0.5) returned %d results\n", len(hnswResults))
	for i, doc := range hnswResults {
		pk := doc.GetPK()
		fmt.Printf("  %d. PK=%s, Score=%.4f\n", i+1, pk, doc.GetScore())
	}

	// === Example 6: Query with IVF Parameters ===
	fmt.Println("\n--- Example 6: Query with IVF Parameters ---")
	queryWithIVF := zvec.NewVectorQuery()
	defer queryWithIVF.Destroy()
	queryWithIVF.SetFieldName("embedding")
	queryWithIVF.SetTopK(3)
	queryWithIVF.SetQueryVector(queryVector)
	queryWithIVF.SetFilter("")
	queryWithIVF.SetIncludeVector(false)
	queryWithIVF.SetIncludeDocID(true)

	// Set IVF query parameters
	ivfQueryParams := zvec.NewIVFQueryParams(10, false, 1.0)
	defer ivfQueryParams.Destroy()
	queryWithIVF.SetIVFParams(ivfQueryParams)

	ivfResults, err := collection.Query(queryWithIVF)
	if err != nil {
		log.Fatalf("Failed to query with IVF params: %v", err)
	}
	defer zvec.FreeDocs(ivfResults)

	fmt.Printf("✓ Query with IVF params (nprobe=10) returned %d results\n", len(ivfResults))
	for i, doc := range ivfResults {
		pk := doc.GetPK()
		fmt.Printf("  %d. PK=%s, Score=%.4f\n", i+1, pk, doc.GetScore())
	}

	// === Example 7: GroupByVectorQuery ===
	fmt.Println("\n--- Example 7: GroupByVectorQuery ---")
	groupByQuery := zvec.NewGroupByVectorQuery()
	defer groupByQuery.Destroy()
	groupByQuery.SetFieldName("embedding")
	groupByQuery.SetGroupByFieldName("category")
	groupByQuery.SetGroupCount(2)
	groupByQuery.SetGroupTopK(2)
	groupByQuery.SetQueryVector(queryVector)
	groupByQuery.SetFilter("")
	groupByQuery.SetIncludeVector(false)
	groupByQuery.SetOutputFields([]string{"id", "category"})

	fmt.Println("✓ GroupByVectorQuery configured:")
	fmt.Println("  - Group by field: category")
	fmt.Println("  - Group count: 2")
	fmt.Println("  - Group top-k: 2")
	fmt.Println("  (Note: GroupByVectorQuery uses a separate query API — see collection_management example)")

	fmt.Println()
	fmt.Println("✓ Vector Query example completed successfully!")
}
