# ZVec Go SDK

English | [中文](README_CN.md)

Go bindings for the [zvec](https://github.com/alibaba/zvec) vector database, powered by cgo wrapping the zvec C-API.

## Introduction

zvec is a high-performance vector database supporting multiple index types (HNSW, IVF, Flat, Invert) and rich data types. zvec-go provides complete Go language bindings, allowing you to easily leverage zvec's powerful capabilities in your Go projects.

## Prerequisites

- **Go** ≥ 1.21
- **C compiler** (gcc or clang) for cgo
- **CMake** ≥ 3.20 and **Ninja** (for building the C-API library)

## Quick Start

```bash
# Clone with submodules
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# Build the C-API library using Makefile
make build-zvec

# Run tests
make test
```

Or use the full build commands:

```bash
# Clone with submodules
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# Build the C-API library from submodule
cd zvec && mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release -DBUILD_C_BINDINGS=ON -G Ninja
cmake --build . -j$(nproc 2>/dev/null || sysctl -n hw.ncpu) --target zvec_c_api
cd ../..

# Run tests
go test -tags integration -count=1 -v ./...
```

## Installation

zvec-go provides **two build modes** to suit different users:

### Mode 1: Vendor Mode (Default — `go get` + `go generate`)

Pre-built libraries are distributed via GitHub Releases. Use `go get` to fetch the code, then `go generate` to download the pre-built library for your platform:

```bash
# 1. Add the dependency
go get github.com/zvec-ai/zvec-go

# 2. Download pre-built library for your platform
#    (downloads from GitHub Releases, extracts to lib/)
go generate github.com/zvec-ai/zvec-go

# 3. Build (cgo is required)
CGO_ENABLED=1 go build .
```

Supported platforms: **Linux (x64, ARM64)**, **macOS (ARM64)**, **Windows (x64)**.

You can also specify a version explicitly:

```bash
go run github.com/zvec-ai/zvec-go/cmd/download-libs@latest -version v0.3.1
```

### Mode 2: Source Mode (Build from Source)

For developers who want to use a custom zvec version, contribute to the project, or build for unsupported platforms:

```bash
# Clone with submodules
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# Build the C-API library
make build-zvec

# Use in your project with replace directive
# In your project's go.mod:
#   require github.com/zvec-ai/zvec-go v0.0.0
#   replace github.com/zvec-ai/zvec-go => /path/to/zvec-go

# Build with source tag
CGO_ENABLED=1 go build -tags source ./...

# Run tests
go test -tags "source integration" -v ./...
```

### Which Mode Should I Use?

| Scenario | Mode | Build Tag |
|----------|------|-----------|
| Just want to use zvec-go in my project | **Vendor** (default) | _(none)_ |
| Contributing to zvec-go development | **Source** | `-tags source` |
| Need a custom/latest zvec version | **Source** | `-tags source` |
| Building for an unsupported platform | **Source** | `-tags source` |
| AI/LLM agent integrating zvec-go | **Vendor** (default) | _(none)_ |

## Usage

```go
package main

import (
    "fmt"
    "log"

    zvec "github.com/zvec-ai/zvec-go"
)

func main() {
    // Initialize zvec
    if err := zvec.Initialize(nil); err != nil {
        log.Fatal(err)
    }
    defer zvec.Shutdown()

    // Create a collection schema
    schema := zvec.NewCollectionSchema("example")
    defer schema.Destroy()

    // Add an ID field (primary key, with invert index)
    idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
    idField.SetIndexParams(zvec.NewInvertIndexParams(true, false))
    schema.AddField(idField)

    // Add a vector field (with HNSW index)
    embField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 4)
    embField.SetIndexParams(zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200))
    schema.AddField(embField)

    // Create and open a collection
    collection, err := zvec.CreateAndOpen("./my_data", schema, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer collection.Close()

    // Insert a document
    doc := zvec.NewDoc()
    doc.SetPK("doc1")
    doc.AddStringField("id", "doc1")
    doc.AddVectorFP32Field("embedding", []float32{0.1, 0.2, 0.3, 0.4})
    collection.Insert([]*zvec.Doc{doc})
    doc.Destroy()

    // Vector query
    query := zvec.NewVectorQuery()
    query.SetFieldName("embedding")
    query.SetQueryVector([]float32{0.4, 0.3, 0.3, 0.1})
    query.SetTopK(10)

    results, _ := collection.Query(query)
    query.Destroy()
    defer zvec.FreeDocs(results)

    for _, r := range results {
        fmt.Printf("PK=%s Score=%.4f\n", r.GetPK(), r.GetScore())
    }
}
```

## API Reference

### Initialization & Configuration

| API | Description |
|-----|-------------|
| `Initialize(config)` | Initialize the zvec library |
| `Shutdown()` | Shut down the zvec library and release resources |
| `IsInitialized()` | Check if the library is initialized |
| `GetVersion()` | Get the version string |
| `GetVersionMajor()` | Get the major version number |
| `GetVersionMinor()` | Get the minor version number |
| `GetVersionPatch()` | Get the patch version number |
| `CheckVersion(major, minor, patch)` | Check if the version is compatible |

### Schema & Index

| API | Description |
|-----|-------------|
| `NewCollectionSchema(name)` | Create a collection schema |
| `NewFieldSchema(name, dataType, nullable, dim)` | Create a field schema |
| `NewHNSWIndexParams(metricType, M, efConstruction)` | Create HNSW index parameters |
| `NewIVFIndexParams(metricType, nlist, nIters, useSoar)` | Create IVF index parameters |
| `NewFlatIndexParams(metricType)` | Create Flat index parameters |
| `NewInvertIndexParams(enable, wildcard)` | Create invert index parameters |
| `SetIndexParams(params)` | Set field index parameters |

### Collection Operations

| API | Description |
|-----|-------------|
| `CreateAndOpen(path, schema, options)` | Create and open a collection |
| `Open(path, options)` | Open an existing collection |
| `Close()` | Close a collection |
| `Destroy(path)` | Destroy a collection |
| `Flush()` | Flush data to disk |
| `Optimize()` | Optimize the collection |
| `GetStats()` | Get collection statistics |
| `GetSchema()` | Get the collection schema |
| `GetOptions()` | Get collection options |
| `AddColumn(field)` | Add a column |
| `DropColumn(fieldName)` | Drop a column |
| `AlterColumn(fieldName, field)` | Alter a column |
| `CreateIndex(fieldName, params)` | Create an index |
| `DropIndex(fieldName)` | Drop an index |

### Document Operations

| API | Description |
|-----|-------------|
| `NewDoc()` | Create a new document |
| `Destroy()` | Destroy a document and release resources |
| `SetPK(pk)` | Set the primary key |
| `GetPK()` | Get the primary key |
| `GetDocID()` | Get the document ID |
| `AddStringField(name, value)` | Add a string field |
| `AddBoolField(name, value)` | Add a boolean field |
| `AddInt32Field(name, value)` | Add an Int32 field |
| `AddInt64Field(name, value)` | Add an Int64 field |
| `AddFloatField(name, value)` | Add a Float field |
| `AddDoubleField(name, value)` | Add a Double field |
| `AddVectorFP32Field(name, value)` | Add an FP32 vector field |
| `SetFieldNull(name)` | Set a field to NULL |
| `RemoveField(name)` | Remove a field |
| `HasField(name)` | Check if a field exists |

### Write Operations

| API | Description |
|-----|-------------|
| `Insert(docs)` | Insert documents |
| `Update(docs)` | Update documents |
| `Upsert(docs)` | Insert or update documents |
| `Delete(pks)` | Delete documents by primary keys |
| `DeleteByFilter(filter)` | Delete documents by filter expression |

### Query Operations

| API | Description |
|-----|-------------|
| `NewVectorQuery()` | Create a vector query object |
| `SetFieldName(name)` | Set the query field name |
| `SetQueryVector(vector)` | Set the query vector |
| `SetTopK(k)` | Set the number of results to return |
| `SetFilter(filter)` | Set the filter expression |
| `SetOutputFields(fields)` | Set the output fields |
| `SetIncludeVector(include)` | Whether to include vector data |
| `SetIncludeDocID(include)` | Whether to include document ID |
| `Query(query)` | Execute a query |
| `GroupByVectorQuery(query)` | Group-by vector query |
| `Fetch(pks)` | Fetch documents by primary keys |
| `FreeDocs(docs)` | Free query result memory |

### Data Types

| Type | Description |
|------|-------------|
| `DataTypeString` | String type |
| `DataTypeBool` | Boolean type |
| `DataTypeInt32` | 32-bit integer |
| `DataTypeInt64` | 64-bit integer |
| `DataTypeUint32` | 32-bit unsigned integer |
| `DataTypeUint64` | 64-bit unsigned integer |
| `DataTypeFloat` | Single-precision float |
| `DataTypeDouble` | Double-precision float |
| `DataTypeVectorFP32` | FP32 vector |
| `DataTypeBinary` | Binary data |
| `DataTypeArray` | Array type |
| `DataTypeSparseVector` | Sparse vector |

### Index Types & Metrics

| Type | Description |
|------|-------------|
| `MetricTypeL2` | L2 distance |
| `MetricTypeIP` | Inner product |
| `MetricTypeCosine` | Cosine similarity |
| `MetricTypeMIPSL2` | MIPSL2 distance |
| `QuantizeTypeFP16` | FP16 quantization |
| `QuantizeTypeInt8` | Int8 quantization |
| `QuantizeTypeInt4` | Int4 quantization |

### Error Handling

| API | Description |
|-----|-------------|
| `Error.Code()` | Get the error code |
| `Error.Message()` | Get the error message |
| `IsNotFound(err)` | Check if it is a "not found" error |
| `IsAlreadyExists(err)` | Check if it is an "already exists" error |
| `IsInvalidArgument(err)` | Check if it is an "invalid argument" error |

## Examples

The project provides rich example code to help you get started quickly:

- **examples/basic** — Basic usage example, demonstrating initialization, schema definition, CRUD operations, and vector queries
- **examples/schema_and_index** — Schema and index configuration, showing how to define different field and index types
- **examples/crud_operations** — Complete CRUD operations, including insert, update, delete, and more
- **examples/vector_query** — Vector query example, demonstrating various query parameters and filter expressions
- **examples/collection_management** — Collection management, showing creation, opening, optimization, and more
- **examples/error_handling** — Error handling example, showing how to properly handle various error scenarios
- **examples/configuration** — Global configuration example, demonstrating memory limits, thread counts, and other options

Run an example:

```bash
cd examples/basic
go run main.go
```

## Development Guide

If you want to contribute to zvec-go, please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for the detailed contribution guide.

## Syncing with zvec Core

This repository uses a **git submodule** to track the [zvec](https://github.com/alibaba/zvec) core library. To update:

```bash
# Update to latest main
./scripts/sync-zvec.sh

# Update to a specific tag
./scripts/sync-zvec.sh v0.4.0
```

[Dependabot](https://docs.github.com/en/code-security/dependabot) is also configured to automatically create PRs when the zvec submodule has new commits.

## Makefile Commands

The project provides convenient Makefile commands for managing build, test, and development tasks:

| Command | Description |
|---------|-------------|
| `make build-zvec` | Build the zvec C-API library |
| `make build` | Build the C-API library and verify Go compilation |
| `make test` | Run all Go tests |
| `make test-short` | Run tests in short mode (skip long-running tests) |
| `make test-race` | Run tests with race detector |
| `make test-cover` | Run tests and generate coverage report |
| `make bench` | Run performance benchmarks |
| `make fuzz` | Run fuzz tests (default 30s per target, set `FUZZ_TIME` to customize) |
| `make lint` | Run all linter checks |
| `make vet` | Run go vet checks |
| `make fmt` | Format Go source files |
| `make fmt-check` | Check Go file formatting (CI-friendly) |
| `make sync-zvec` | Sync zvec submodule to latest main |
| `make sync-zvec-build` | Sync zvec submodule + rebuild + test |
| `make check-zvec` | Check for upstream C-API changes (no update) |
| `make clean` | Clean build artifacts |
| `make deps` | Download Go module dependencies |
| `make install-tools` | Install development tools (golangci-lint, gofumpt) |
| `make all` | Run full CI check (build, test, lint) |
| `make help` | Show help message |

## Supported Platforms

- Linux (x86_64, ARM64)
- macOS (ARM64)
- Windows (x86_64)

## License

Apache License 2.0
