// Package main demonstrates collection management operations in zvec.
//
// This example shows how to:
// - Create and open collections with different options
// - Open existing collections
// - Configure collection options (Mmap, MaxBufferSize, ReadOnly)
// - Get collection statistics
// - Flush and optimize collections
// - Add, drop, and alter columns dynamically
// - Create and drop indexes dynamically
// - Destroy collections (delete data)
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/collection_management
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec Collection Management Example ===")
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

	// === Example 1: Create Collection with Options ===
	fmt.Println("--- Example 1: Create Collection with Options ---")

	// Create collection schema
	schema := zvec.NewCollectionSchema("managed_collection")
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

	// Create collection options
	options := zvec.NewCollectionOptions()
	defer options.Destroy()
	options.SetEnableMmap(true)
	options.SetMaxBufferSize(1024 * 1024 * 100) // 100MB
	options.SetReadOnly(false)

	fmt.Printf("✓ Collection options: Mmap=%v, MaxBufferSize=%d, ReadOnly=%v\n",
		options.GetEnableMmap(), options.GetMaxBufferSize(), options.GetReadOnly())

	// Create and open collection with options
	collectionPath := "./test_managed_collection"
	defer os.RemoveAll(collectionPath)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, options)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer collection.Close()
	fmt.Println("✓ Collection created with options")

	// Insert sample data
	doc := zvec.NewDoc()
	defer doc.Destroy()
	doc.SetPK("doc1")
	doc.AddStringField("id", "doc1")
	vector := make([]float32, 128)
	for i := range vector {
		vector[i] = float32(i) * 0.01
	}
	doc.AddVectorFP32Field("embedding", vector)

	result, err := collection.Insert([]*zvec.Doc{doc})
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	fmt.Printf("✓ Inserted %d document(s)\n", result.SuccessCount)

	// === Example 2: Get Collection Statistics ===
	fmt.Println("\n--- Example 2: Get Collection Statistics ---")
	stats, err := collection.GetStats()
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	fmt.Printf("✓ Collection Stats:\n")
	fmt.Printf("  - Document Count: %d\n", stats.DocCount)
	fmt.Printf("  - Index Count: %d\n", stats.IndexCount)
	fmt.Printf("  - Index Names: %v\n", stats.IndexNames)
	if len(stats.IndexCompleteness) > 0 {
		fmt.Printf("  - Index Completeness: %v\n", stats.IndexCompleteness)
	}

	// === Example 3: Flush Collection ===
	fmt.Println("\n--- Example 3: Flush Collection ---")
	if err := collection.Flush(); err != nil {
		log.Fatalf("Failed to flush collection: %v", err)
	}
	fmt.Println("✓ Collection flushed to disk")

	// === Example 4: Open Existing Collection ===
	fmt.Println("\n--- Example 4: Open Existing Collection ---")

	// Close current collection
	if err := collection.Close(); err != nil {
		log.Fatalf("Failed to close collection: %v", err)
	}

	// Reopen with read-only options
	readOnlyOptions := zvec.NewCollectionOptions()
	defer readOnlyOptions.Destroy()
	readOnlyOptions.SetEnableMmap(true)
	readOnlyOptions.SetReadOnly(true)

	reopenedCollection, err := zvec.Open(collectionPath, readOnlyOptions)
	if err != nil {
		log.Fatalf("Failed to open collection: %v", err)
	}
	defer reopenedCollection.Close()
	fmt.Println("✓ Collection reopened in read-only mode")

	// Verify data
	fetchedDocs, err := reopenedCollection.Fetch([]string{"doc1"})
	if err != nil {
		log.Fatalf("Failed to fetch documents: %v", err)
	}
	defer zvec.FreeDocs(fetchedDocs)
	fmt.Printf("✓ Fetched %d document(s) from reopened collection\n", len(fetchedDocs))

	// === Example 5: Add Column ===
	fmt.Println("\n--- Example 5: Add Column ---")

	// Close read-only collection and reopen in read-write mode
	if err := reopenedCollection.Close(); err != nil {
		log.Fatalf("Failed to close read-only collection: %v", err)
	}

	readWriteOptions := zvec.NewCollectionOptions()
	defer readWriteOptions.Destroy()
	readWriteOptions.SetEnableMmap(true)
	readWriteOptions.SetReadOnly(false)

	readWriteCollection, err := zvec.Open(collectionPath, readWriteOptions)
	if err != nil {
		log.Fatalf("Failed to open collection in read-write mode: %v", err)
	}
	defer readWriteCollection.Close()

	// Add a new column
	newField := zvec.NewFieldSchema("category", zvec.DataTypeString, true, 0)
	defer newField.Destroy()
	if err := readWriteCollection.AddColumn(newField, ""); err != nil {
		log.Fatalf("Failed to add column: %v", err)
	}
	fmt.Println("✓ Added 'category' column to collection")

	// Verify new column exists
	retrievedSchema, err := readWriteCollection.GetSchema()
	if err != nil {
		log.Fatalf("Failed to get schema: %v", err)
	}
	defer retrievedSchema.Destroy()

	if retrievedSchema.HasField("category") {
		fmt.Println("✓ Verified 'category' column exists in schema")
	}

	// === Example 6: Create Index Dynamically ===
	fmt.Println("\n--- Example 6: Create Index Dynamically ---")

	// Create index for the new category field
	categoryIndexParams := zvec.NewInvertIndexParams(true, false)
	defer categoryIndexParams.Destroy()
	if err := readWriteCollection.CreateIndex("category", categoryIndexParams); err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	fmt.Println("✓ Created index for 'category' field")

	// === Example 7: Optimize Collection ===
	fmt.Println("\n--- Example 7: Optimize Collection ---")
	if err := readWriteCollection.Optimize(); err != nil {
		log.Fatalf("Failed to optimize collection: %v", err)
	}
	fmt.Println("✓ Collection optimized")

	// === Example 8: Alter Column ===
	fmt.Println("\n--- Example 8: Alter Column ---")

	// Alter column (rename)
	if err := readWriteCollection.AlterColumn("category", "category_new", nil); err != nil {
		log.Fatalf("Failed to alter column: %v", err)
	}
	fmt.Println("✓ Renamed 'category' column to 'category_new'")

	// Verify rename
	if !retrievedSchema.HasField("category") && retrievedSchema.HasField("category_new") {
		fmt.Println("✓ Verified column rename")
	}

	// === Example 9: Drop Index ===
	fmt.Println("\n--- Example 9: Drop Index ---")
	if err := readWriteCollection.DropIndex("category_new"); err != nil {
		log.Fatalf("Failed to drop index: %v", err)
	}
	fmt.Println("✓ Dropped index for 'category_new' field")

	// === Example 10: Drop Column ===
	fmt.Println("\n--- Example 10: Drop Column ---")
	if err := readWriteCollection.DropColumn("category_new"); err != nil {
		log.Fatalf("Failed to drop column: %v", err)
	}
	fmt.Println("✓ Dropped 'category_new' column from collection")

	// === Example 11: Destroy Collection ===
	fmt.Println("\n--- Example 11: Destroy Collection ---")

	// Close collection first
	if err := readWriteCollection.Close(); err != nil {
		log.Fatalf("Failed to close collection: %v", err)
	}

	// Open again to destroy
	destroyCollection, err := zvec.Open(collectionPath, nil)
	if err != nil {
		log.Fatalf("Failed to open collection for destruction: %v", err)
	}

	// Destroy collection (deletes data on disk)
	if err := destroyCollection.Destroy(); err != nil {
		log.Fatalf("Failed to destroy collection: %v", err)
	}
	fmt.Println("✓ Collection destroyed (data deleted from disk)")

	// Verify collection is destroyed
	_, err = zvec.Open(collectionPath, nil)
	if err != nil {
		fmt.Println("✓ Verified collection is destroyed (cannot open)")
	}

	fmt.Println()
	fmt.Println("✓ Collection Management example completed successfully!")
}
