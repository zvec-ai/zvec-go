// Package main demonstrates CRUD operations in zvec.
//
// This example shows how to:
// - Insert single and multiple documents
// - Update existing documents
// - Upsert documents (insert or update)
// - Delete documents by primary key
// - Delete documents by filter expression
// - Fetch documents by primary key
// - Use WriteResult to track operation results
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/crud_operations
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec CRUD Operations Example ===")
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
	schema := zvec.NewCollectionSchema("crud_example_collection")
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

	// Add text field
	textField := zvec.NewFieldSchema("text", zvec.DataTypeString, true, 0)
	defer textField.Destroy()
	if err := textField.SetIndexParams(invertParams); err != nil {
		log.Fatalf("Failed to set index params for text field: %v", err)
	}
	if err := schema.AddField(textField); err != nil {
		log.Fatalf("Failed to add text field: %v", err)
	}

	// Add embedding field
	embeddingField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 3)
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
	collectionPath := "./test_crud_collection"
	defer os.RemoveAll(collectionPath)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, nil)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer collection.Close()
	fmt.Println("✓ Collection created successfully")

	// === Insert Single Document ===
	fmt.Println("\n--- Insert Single Document ---")
	doc1 := zvec.NewDoc()
	defer doc1.Destroy()
	doc1.SetPK("doc1")
	doc1.AddStringField("id", "doc1")
	doc1.AddStringField("text", "First document")
	doc1.AddVectorFP32Field("embedding", []float32{0.1, 0.2, 0.3})

	result, err := collection.Insert([]*zvec.Doc{doc1})
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	fmt.Printf("✓ Single document inserted — Success: %d, Failed: %d\n", result.SuccessCount, result.ErrorCount)

	// === Insert Multiple Documents (Batch) ===
	fmt.Println("\n--- Insert Multiple Documents (Batch) ---")
	doc2 := zvec.NewDoc()
	defer doc2.Destroy()
	doc2.SetPK("doc2")
	doc2.AddStringField("id", "doc2")
	doc2.AddStringField("text", "Second document")
	doc2.AddVectorFP32Field("embedding", []float32{0.4, 0.5, 0.6})

	doc3 := zvec.NewDoc()
	defer doc3.Destroy()
	doc3.SetPK("doc3")
	doc3.AddStringField("id", "doc3")
	doc3.AddStringField("text", "Third document")
	doc3.AddVectorFP32Field("embedding", []float32{0.7, 0.8, 0.9})

	result, err = collection.Insert([]*zvec.Doc{doc2, doc3})
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}
	fmt.Printf("✓ Batch insert completed — Success: %d, Failed: %d\n", result.SuccessCount, result.ErrorCount)

	// Flush to ensure data is persisted
	if err := collection.Flush(); err != nil {
		log.Fatalf("Failed to flush collection: %v", err)
	}

	// === Fetch Documents by Primary Key ===
	fmt.Println("\n--- Fetch Documents by Primary Key ---")
	fetchedDocs, err := collection.Fetch([]string{"doc1", "doc2"})
	if err != nil {
		log.Fatalf("Failed to fetch documents: %v", err)
	}
	defer zvec.FreeDocs(fetchedDocs)

	fmt.Printf("✓ Fetched %d documents\n", len(fetchedDocs))
	for i, doc := range fetchedDocs {
		pk := doc.GetPK()
		text, _ := doc.GetStringField("text")
		fmt.Printf("  %d. PK=%s, Text=%s\n", i+1, pk, text)
	}

	// === Update Document ===
	fmt.Println("\n--- Update Document ---")
	docUpdate := zvec.NewDoc()
	defer docUpdate.Destroy()
	docUpdate.SetPK("doc1")
	docUpdate.AddStringField("id", "doc1")
	docUpdate.AddStringField("text", "Updated first document")
	docUpdate.AddVectorFP32Field("embedding", []float32{0.15, 0.25, 0.35})

	result, err = collection.Update([]*zvec.Doc{docUpdate})
	if err != nil {
		log.Fatalf("Failed to update document: %v", err)
	}
	fmt.Printf("✓ Document updated — Success: %d, Failed: %d\n", result.SuccessCount, result.ErrorCount)

	// Verify update
	fetchedDoc, err := collection.Fetch([]string{"doc1"})
	if err != nil {
		log.Fatalf("Failed to fetch updated document: %v", err)
	}
	defer zvec.FreeDocs(fetchedDoc)

	if len(fetchedDoc) > 0 {
		updatedText, _ := fetchedDoc[0].GetStringField("text")
		fmt.Printf("✓ Verified update: Text is now '%s'\n", updatedText)
	}

	// === Upsert Document (Insert or Update) ===
	fmt.Println("\n--- Upsert Document ---")
	// Upsert an existing document (should update)
	docUpsertUpdate := zvec.NewDoc()
	defer docUpsertUpdate.Destroy()
	docUpsertUpdate.SetPK("doc2")
	docUpsertUpdate.AddStringField("id", "doc2")
	docUpsertUpdate.AddStringField("text", "Upserted second document")
	docUpsertUpdate.AddVectorFP32Field("embedding", []float32{0.45, 0.55, 0.65})

	result, err = collection.Upsert([]*zvec.Doc{docUpsertUpdate})
	if err != nil {
		log.Fatalf("Failed to upsert existing document: %v", err)
	}
	fmt.Printf("✓ Existing document upserted (updated) — Success: %d, Failed: %d\n", result.SuccessCount, result.ErrorCount)

	// Upsert a new document (should insert)
	docUpsertInsert := zvec.NewDoc()
	defer docUpsertInsert.Destroy()
	docUpsertInsert.SetPK("doc4")
	docUpsertInsert.AddStringField("id", "doc4")
	docUpsertInsert.AddStringField("text", "New upserted document")
	docUpsertInsert.AddVectorFP32Field("embedding", []float32{1.0, 1.1, 1.2})

	result, err = collection.Upsert([]*zvec.Doc{docUpsertInsert})
	if err != nil {
		log.Fatalf("Failed to upsert new document: %v", err)
	}
	fmt.Printf("✓ New document upserted (inserted) — Success: %d, Failed: %d\n", result.SuccessCount, result.ErrorCount)

	// Flush to persist upserts
	if err := collection.Flush(); err != nil {
		log.Fatalf("Failed to flush collection: %v", err)
	}

	// === Delete Document by Primary Key ===
	fmt.Println("\n--- Delete Document by Primary Key ---")
	result, err = collection.Delete([]string{"doc3"})
	if err != nil {
		log.Fatalf("Failed to delete document: %v", err)
	}
	fmt.Printf("✓ Document deleted by PK — Success: %d, Failed: %d\n", result.SuccessCount, result.ErrorCount)

	// Verify deletion
	fetchedAfterDelete, err := collection.Fetch([]string{"doc3"})
	if err != nil {
		log.Fatalf("Failed to fetch after delete: %v", err)
	}
	defer zvec.FreeDocs(fetchedAfterDelete)

	if len(fetchedAfterDelete) == 0 {
		fmt.Println("✓ Verified deletion: doc3 no longer exists")
	}

	// === Delete Documents by Filter ===
	fmt.Println("\n--- Delete Documents by Filter ---")
	// Delete documents where text contains "second"
	filter := "text == 'Upserted second document'"
	err = collection.DeleteByFilter(filter)
	if err != nil {
		log.Fatalf("Failed to delete by filter: %v", err)
	}
	fmt.Printf("✓ Documents deleted by filter '%s'\n", filter)

	// Verify filter deletion
	fetchedAfterFilterDelete, err := collection.Fetch([]string{"doc2"})
	if err != nil {
		log.Fatalf("Failed to fetch after filter delete: %v", err)
	}
	defer zvec.FreeDocs(fetchedAfterFilterDelete)

	if len(fetchedAfterFilterDelete) == 0 {
		fmt.Println("✓ Verified filter deletion: doc2 no longer exists")
	}

	// === Check Final Collection Stats ===
	fmt.Println("\n--- Final Collection Stats ---")
	stats, err := collection.GetStats()
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	fmt.Printf("✓ Final document count: %d\n", stats.DocCount)

	// === List Remaining Documents ===
	fmt.Println("\n--- Remaining Documents ---")
	query := zvec.NewVectorQuery()
	defer query.Destroy()
	query.SetFieldName("embedding")
	query.SetQueryVector([]float32{0.1, 0.2, 0.3})
	query.SetTopK(10)
	query.SetFilter("")
	query.SetIncludeVector(false)
	query.SetIncludeDocID(true)

	results, err := collection.Query(query)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer zvec.FreeDocs(results)

	fmt.Printf("✓ Query returned %d documents\n", len(results))
	for i, doc := range results {
		pk := doc.GetPK()
		text, _ := doc.GetStringField("text")
		fmt.Printf("  %d. PK=%s, Text=%s\n", i+1, pk, text)
	}

	fmt.Println()
	fmt.Println("✓ CRUD Operations example completed successfully!")
}
