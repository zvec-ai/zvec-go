// Package main demonstrates error handling best practices with the zvec Go SDK.
//
// This example covers:
//   - Using type assertions to inspect zvec error codes
//   - Using helper functions (IsNotFound, IsAlreadyExists, IsInvalidArgument)
//   - Handling common error scenarios gracefully
//   - Proper resource cleanup on errors
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/error_handling
//	go run main.go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec Go SDK — Error Handling Example ===")
	fmt.Println()

	if err := zvec.Initialize(nil); err != nil {
		log.Fatalf("Failed to initialize zvec: %v", err)
	}
	defer func() {
		if err := zvec.Shutdown(); err != nil {
			log.Printf("Warning: shutdown error: %v", err)
		}
	}()

	collectionPath := "./test_error_handling"
	defer os.RemoveAll(collectionPath)

	// -----------------------------------------------------------------------
	// 1. Type assertion to inspect error details
	// -----------------------------------------------------------------------
	fmt.Println("--- 1. Type Assertion for Error Inspection ---")

	// Try to open a non-existent collection
	_, openErr := zvec.Open("/non/existent/path", nil)
	if openErr != nil {
		var zvecErr *zvec.Error
		if errors.As(openErr, &zvecErr) {
			fmt.Printf("✓ Caught zvec error:\n")
			fmt.Printf("  Code:    %d\n", zvecErr.Code)
			fmt.Printf("  Message: %s\n", zvecErr.Message)
		} else {
			fmt.Printf("  Non-zvec error: %v\n", openErr)
		}
	}

	// -----------------------------------------------------------------------
	// 2. Using IsNotFound helper
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 2. IsNotFound Helper ---")

	// Create a collection for testing
	schema := zvec.NewCollectionSchema("error_test")
	defer schema.Destroy()

	invertParams := zvec.NewInvertIndexParams(true, false)
	defer invertParams.Destroy()
	hnswParams := zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200)
	defer hnswParams.Destroy()

	idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
	defer idField.Destroy()
	idField.SetIndexParams(invertParams)
	schema.AddField(idField)

	embField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 3)
	defer embField.Destroy()
	embField.SetIndexParams(hnswParams)
	schema.AddField(embField)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, nil)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer collection.Close()

	// Insert a document
	doc := zvec.NewDoc()
	doc.SetPK("existing_doc")
	doc.AddStringField("id", "existing_doc")
	doc.AddVectorFP32Field("embedding", []float32{0.1, 0.2, 0.3})
	if _, err := collection.Insert([]*zvec.Doc{doc}); err != nil {
		log.Fatalf("Failed to insert: %v", err)
	}
	doc.Destroy()

	if err := collection.Flush(); err != nil {
		log.Fatalf("Failed to flush: %v", err)
	}

	// Fetch an existing document
	existingDocs, fetchErr := collection.Fetch([]string{"existing_doc"})
	if fetchErr != nil {
		if zvec.IsNotFound(fetchErr) {
			fmt.Println("  Document not found (unexpected)")
		} else {
			log.Fatalf("Fetch error: %v", fetchErr)
		}
	} else {
		fmt.Printf("✓ Found document: PK=%s\n", existingDocs[0].GetPK())
		zvec.FreeDocs(existingDocs)
	}

	// Fetch a non-existing document — returns empty result (not an error)
	missingDocs, fetchErr := collection.Fetch([]string{"non_existent_doc"})
	if fetchErr != nil {
		if zvec.IsNotFound(fetchErr) {
			fmt.Println("✓ IsNotFound correctly identified missing document")
		} else {
			fmt.Printf("  Fetch returned error: %v\n", fetchErr)
		}
	} else {
		fmt.Printf("✓ Fetch returned %d results for non-existent key (no error)\n", len(missingDocs))
		zvec.FreeDocs(missingDocs)
	}

	// -----------------------------------------------------------------------
	// 3. Using IsInvalidArgument helper
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 3. IsInvalidArgument Helper ---")

	// Try to add an empty vector
	badDoc := zvec.NewDoc()
	defer badDoc.Destroy()
	addErr := badDoc.AddVectorFP32Field("embedding", []float32{})
	if addErr != nil {
		if zvec.IsInvalidArgument(addErr) {
			fmt.Printf("✓ IsInvalidArgument caught: %v\n", addErr)
		} else {
			fmt.Printf("  Unexpected error type: %v\n", addErr)
		}
	}

	// -----------------------------------------------------------------------
	// 4. Graceful error recovery pattern
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 4. Graceful Error Recovery Pattern ---")

	insertOrUpdate := func(coll *zvec.Collection, primaryKey string, vector []float32) error {
		newDoc := zvec.NewDoc()
		defer newDoc.Destroy()
		newDoc.SetPK(primaryKey)
		newDoc.AddStringField("id", primaryKey)
		newDoc.AddVectorFP32Field("embedding", vector)

		// Try insert first; if already exists, fall back to upsert
		_, insertErr := coll.Insert([]*zvec.Doc{newDoc})
		if insertErr != nil {
			if zvec.IsAlreadyExists(insertErr) {
				fmt.Printf("  Document '%s' exists, upserting instead...\n", primaryKey)
				_, upsertErr := coll.Upsert([]*zvec.Doc{newDoc})
				return upsertErr
			}
			return insertErr
		}
		return nil
	}

	// First call: insert
	if err := insertOrUpdate(collection, "recovery_doc", []float32{0.4, 0.5, 0.6}); err != nil {
		log.Fatalf("insertOrUpdate failed: %v", err)
	}
	fmt.Println("✓ First insertOrUpdate succeeded (insert path)")

	// Second call: may trigger upsert path
	if err := insertOrUpdate(collection, "recovery_doc", []float32{0.7, 0.8, 0.9}); err != nil {
		log.Fatalf("insertOrUpdate failed: %v", err)
	}
	fmt.Println("✓ Second insertOrUpdate succeeded")

	// -----------------------------------------------------------------------
	// 5. Error code switch pattern
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 5. Error Code Switch Pattern ---")

	handleError := func(err error) string {
		if err == nil {
			return "success"
		}

		var zvecErr *zvec.Error
		if !errors.As(err, &zvecErr) {
			return fmt.Sprintf("non-zvec error: %v", err)
		}

		switch zvecErr.Code {
		case zvec.ErrNotFound:
			return "resource not found — check if the collection/document exists"
		case zvec.ErrAlreadyExists:
			return "resource already exists — use upsert instead of insert"
		case zvec.ErrInvalidArgument:
			return "invalid argument — check input parameters"
		case zvec.ErrPermissionDenied:
			return "permission denied — check file permissions"
		case zvec.ErrResourceExhausted:
			return "resource exhausted — consider increasing memory limit"
		case zvec.ErrInternalError:
			return "internal error — this may be a bug, please report it"
		default:
			return fmt.Sprintf("error code %d: %s", zvecErr.Code, zvecErr.Message)
		}
	}

	// Demonstrate with a real error
	_, realErr := zvec.Open("/does/not/exist", nil)
	fmt.Printf("✓ Error handling result: %s\n", handleError(realErr))
	fmt.Printf("✓ Nil error result: %s\n", handleError(nil))

	fmt.Println()
	fmt.Println("✓ Error handling example completed!")
}
