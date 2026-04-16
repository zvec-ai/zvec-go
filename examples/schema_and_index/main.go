// Package main demonstrates schema and index configuration in zvec.
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/schema_and_index
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec Schema and Index Configuration Example ===")
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

	// Example 1: Create schema with multiple field types
	fmt.Println("--- Example 1: Schema with Multiple Field Types ---")
	schema := zvec.NewCollectionSchema("multi_field_collection")
	defer schema.Destroy()

	// Add primary key field (String with inverted index)
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
	fmt.Println("✓ Added String field 'id' with Invert index")

	// Add Int64 field (nullable)
	ageField := zvec.NewFieldSchema("age", zvec.DataTypeInt64, true, 0)
	defer ageField.Destroy()
	if err := schema.AddField(ageField); err != nil {
		log.Fatalf("Failed to add age field: %v", err)
	}
	fmt.Println("✓ Added Int64 field 'age' (nullable)")

	// Add Float field
	scoreField := zvec.NewFieldSchema("score", zvec.DataTypeFloat, false, 0)
	defer scoreField.Destroy()
	if err := schema.AddField(scoreField); err != nil {
		log.Fatalf("Failed to add score field: %v", err)
	}
	fmt.Println("✓ Added Float field 'score'")

	// Add Bool field
	activeField := zvec.NewFieldSchema("active", zvec.DataTypeBool, false, 0)
	defer activeField.Destroy()
	if err := schema.AddField(activeField); err != nil {
		log.Fatalf("Failed to add active field: %v", err)
	}
	fmt.Println("✓ Added Bool field 'active'")

	// Add VectorFP32 field with HNSW index
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
	fmt.Println("✓ Added VectorFP32 field 'embedding' with HNSW index (Cosine, m=16, efConstruction=200)")

	// Set max documents per segment
	schema.SetMaxDocCountPerSegment(10000)
	fmt.Println("✓ Set max document count per segment to 10000")

	fmt.Println()

	// Example 2: Different index types and metrics
	fmt.Println("--- Example 2: Different Index Types and Metrics ---")

	// HNSW index with L2 metric
	hnswL2Params := zvec.NewHNSWIndexParams(zvec.MetricTypeL2, 32, 100)
	defer hnswL2Params.Destroy()
	fmt.Println("✓ Created HNSW index params with L2 metric (m=32, efConstruction=100)")

	// HNSW index with IP metric
	hnswIPParams := zvec.NewHNSWIndexParams(zvec.MetricTypeIP, 24, 150)
	defer hnswIPParams.Destroy()
	fmt.Println("✓ Created HNSW index params with IP metric (m=24, efConstruction=150)")

	// IVF index with Cosine metric
	ivfParams := zvec.NewIVFIndexParams(zvec.MetricTypeCosine, 100, 20, false)
	defer ivfParams.Destroy()
	fmt.Println("✓ Created IVF index params with Cosine metric (nList=100, nIters=20)")

	// Flat index with L2 metric
	flatParams := zvec.NewFlatIndexParams(zvec.MetricTypeL2)
	defer flatParams.Destroy()
	fmt.Println("✓ Created Flat index params with L2 metric")

	// Invert index with wildcard support
	invertWildcardParams := zvec.NewInvertIndexParams(true, true)
	defer invertWildcardParams.Destroy()
	fmt.Println("✓ Created Invert index params with wildcard support")

	fmt.Println()

	// Example 3: Quantization types
	fmt.Println("--- Example 3: Quantization Types ---")

	hnswQuantizedParams := zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200)
	defer hnswQuantizedParams.Destroy()
	if err := hnswQuantizedParams.SetQuantizeType(zvec.QuantizeTypeFP16); err != nil {
		log.Fatalf("Failed to set quantize type: %v", err)
	}
	fmt.Printf("✓ HNSW with FP16 quantization — type=%d, metric=%d\n",
		hnswQuantizedParams.GetType(), hnswQuantizedParams.GetMetricType())

	hnswInt8Params := zvec.NewHNSWIndexParams(zvec.MetricTypeL2, 16, 200)
	defer hnswInt8Params.Destroy()
	if err := hnswInt8Params.SetQuantizeType(zvec.QuantizeTypeInt8); err != nil {
		log.Fatalf("Failed to set quantize type: %v", err)
	}
	fmt.Printf("✓ HNSW with Int8 quantization — type=%d, metric=%d\n",
		hnswInt8Params.GetType(), hnswInt8Params.GetMetricType())

	fmt.Println()

	// Example 4: Create a collection and verify schema
	fmt.Println("--- Example 4: Create Collection and Verify Schema ---")

	collectionPath := "./test_schema_collection"
	defer os.RemoveAll(collectionPath)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, nil)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer collection.Close()
	fmt.Println("✓ Collection created")

	// Retrieve and inspect schema
	retrievedSchema, err := collection.GetSchema()
	if err != nil {
		log.Fatalf("Failed to get schema: %v", err)
	}
	defer retrievedSchema.Destroy()

	fmt.Printf("✓ Schema name: %s\n", retrievedSchema.GetName())
	fmt.Printf("✓ Has 'id' field: %v\n", retrievedSchema.HasField("id"))
	fmt.Printf("✓ Has 'age' field: %v\n", retrievedSchema.HasField("age"))
	fmt.Printf("✓ Has 'embedding' field: %v\n", retrievedSchema.HasField("embedding"))
	fmt.Printf("✓ Has 'nonexistent' field: %v\n", retrievedSchema.HasField("nonexistent"))

	// Inspect field properties
	embField := retrievedSchema.GetField("embedding")
	if embField != nil {
		fmt.Printf("✓ Embedding field — DataType=%d, Dimension=%d, IsVector=%v, HasIndex=%v\n",
			embField.GetDataType(), embField.GetDimension(), embField.IsVectorField(), embField.HasIndex())
	}

	ageFieldRetrieved := retrievedSchema.GetField("age")
	if ageFieldRetrieved != nil {
		fmt.Printf("✓ Age field — DataType=%d, Nullable=%v, IsVector=%v\n",
			ageFieldRetrieved.GetDataType(), ageFieldRetrieved.IsNullable(), ageFieldRetrieved.IsVectorField())
	}

	fmt.Println()
	fmt.Println("✓ Schema and Index example completed successfully!")
}
