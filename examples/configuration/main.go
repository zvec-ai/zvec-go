// Package main demonstrates global configuration options with the zvec Go SDK.
//
// This example covers:
//   - Version information retrieval
//   - Version compatibility checking
//   - Memory limit configuration
//   - Thread count configuration (query and optimize)
//   - Console and file logging configuration
//
// Before running this example, build the zvec C-API library:
//
//	cd /path/to/zvec-go && make build-zvec
//
// Run:
//
//	cd examples/configuration
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	zvec "github.com/zvec-ai/zvec-go"
)

func main() {
	fmt.Println("=== ZVec Go SDK — Configuration Example ===")
	fmt.Println()

	// -----------------------------------------------------------------------
	// 1. Version information (available before Initialize)
	// -----------------------------------------------------------------------
	fmt.Println("--- 1. Version Information ---")

	versionString := zvec.GetVersion()
	majorVersion := zvec.GetVersionMajor()
	minorVersion := zvec.GetVersionMinor()
	patchVersion := zvec.GetVersionPatch()

	fmt.Printf("✓ ZVec version: %s\n", versionString)
	fmt.Printf("✓ Parsed: major=%d, minor=%d, patch=%d\n", majorVersion, minorVersion, patchVersion)

	// -----------------------------------------------------------------------
	// 2. Version compatibility check
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 2. Version Compatibility Check ---")

	isCompatible := zvec.CheckVersion(0, 1, 0)
	fmt.Printf("✓ Compatible with v0.1.0: %v\n", isCompatible)

	isCompatibleFuture := zvec.CheckVersion(99, 0, 0)
	fmt.Printf("✓ Compatible with v99.0.0: %v\n", isCompatibleFuture)

	// -----------------------------------------------------------------------
	// 3. Default initialization (no custom config)
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 3. Default Initialization ---")

	if err := zvec.Initialize(nil); err != nil {
		log.Fatalf("Failed to initialize with defaults: %v", err)
	}
	fmt.Printf("✓ Initialized with defaults, IsInitialized=%v\n", zvec.IsInitialized())

	// Shutdown to reinitialize with custom config
	if err := zvec.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown: %v", err)
	}
	fmt.Println("✓ Shutdown complete")

	// -----------------------------------------------------------------------
	// 4. Custom configuration — memory and threads
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 4. Custom Configuration ---")

	config := zvec.NewConfigData()
	if config == nil {
		log.Fatal("Failed to create config data")
	}
	defer config.Destroy()

	// Set memory limit (e.g., 512 MB)
	memoryLimitBytes := uint64(512 * 1024 * 1024)
	if err := config.SetMemoryLimit(memoryLimitBytes); err != nil {
		log.Fatalf("Failed to set memory limit: %v", err)
	}
	fmt.Printf("✓ Memory limit set to %d bytes (%.0f MB)\n",
		config.GetMemoryLimit(), float64(config.GetMemoryLimit())/(1024*1024))

	// Set query thread count
	if err := config.SetQueryThreadCount(4); err != nil {
		log.Fatalf("Failed to set query thread count: %v", err)
	}
	fmt.Printf("✓ Query thread count: %d\n", config.GetQueryThreadCount())

	// Set optimize thread count
	if err := config.SetOptimizeThreadCount(2); err != nil {
		log.Fatalf("Failed to set optimize thread count: %v", err)
	}
	fmt.Printf("✓ Optimize thread count: %d\n", config.GetOptimizeThreadCount())

	// -----------------------------------------------------------------------
	// 5. Console logging configuration
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 5. Console Logging ---")

	if err := config.SetConsoleLog(zvec.LogLevelInfo); err != nil {
		log.Fatalf("Failed to set console log: %v", err)
	}
	fmt.Println("✓ Console logging enabled at INFO level")

	// Initialize with custom config
	if err := zvec.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize with custom config: %v", err)
	}
	fmt.Printf("✓ Initialized with custom config, IsInitialized=%v\n", zvec.IsInitialized())

	// Quick verification: create a small collection
	collectionPath := "./test_config_collection"
	defer os.RemoveAll(collectionPath)

	schema := zvec.NewCollectionSchema("config_test")
	defer schema.Destroy()

	idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
	defer idField.Destroy()
	invertParams := zvec.NewInvertIndexParams(true, false)
	defer invertParams.Destroy()
	idField.SetIndexParams(invertParams)
	schema.AddField(idField)

	embField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 3)
	defer embField.Destroy()
	hnswParams := zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200)
	defer hnswParams.Destroy()
	embField.SetIndexParams(hnswParams)
	schema.AddField(embField)

	collection, err := zvec.CreateAndOpen(collectionPath, schema, nil)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	doc := zvec.NewDoc()
	doc.SetPK("test1")
	doc.AddStringField("id", "test1")
	doc.AddVectorFP32Field("embedding", []float32{0.1, 0.2, 0.3})
	if _, err := collection.Insert([]*zvec.Doc{doc}); err != nil {
		log.Fatalf("Failed to insert: %v", err)
	}
	doc.Destroy()

	fmt.Println("✓ Collection operations work with custom config")

	if err := collection.Close(); err != nil {
		log.Fatalf("Failed to close: %v", err)
	}

	// -----------------------------------------------------------------------
	// 6. Available log levels
	// -----------------------------------------------------------------------
	fmt.Println("\n--- 6. Available Log Levels ---")

	logLevels := []struct {
		level zvec.LogLevel
		name  string
	}{
		{zvec.LogLevelDebug, "Debug"},
		{zvec.LogLevelInfo, "Info"},
		{zvec.LogLevelWarn, "Warn"},
		{zvec.LogLevelError, "Error"},
		{zvec.LogLevelFatal, "Fatal"},
	}
	for _, logLevel := range logLevels {
		fmt.Printf("  LogLevel%-6s = %d\n", logLevel.name, logLevel.level)
	}

	// Final shutdown
	if err := zvec.Shutdown(); err != nil {
		log.Printf("Warning: shutdown error: %v", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration example completed!")
}
