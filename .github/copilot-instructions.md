# ZVec Go SDK — AI Copilot Instructions

## Project Overview

zvec-go is a Go SDK for the [zvec](https://github.com/alibaba/zvec) vector database. It wraps the zvec C-API via cgo to provide idiomatic Go access to all zvec functionality.

## Architecture

### Dual-Mode Build System

This project supports two build modes controlled by Go build tags:

- **Vendor mode** (default, no tag): Links against pre-built libraries in `lib/{platform}/`. Enables `go get` out-of-the-box usage.
- **Source mode** (`-tags source`): Links against libraries built from the `zvec/` git submodule. For developers and contributors.

### Key Files

| File | Purpose |
|------|---------|
| `cgo_vendor_*.go` | Vendor mode cgo directives (per platform) |
| `cgo_source.go` | Source mode cgo directives |
| `zvec.go` | Library init, version, config (no cgo directives) |
| `collection.go` | Collection CRUD, insert/update/upsert/delete/query/fetch |
| `doc.go` | Document field operations |
| `query.go` | VectorQuery, GroupByVectorQuery, query params |
| `schema.go` | IndexParams, FieldSchema, CollectionSchema |
| `types.go` | Go enum types (DataType, IndexType, MetricType, etc.) |
| `errors.go` | Error handling, C→Go error conversion |

### Directory Structure

```
zvec-go/
├── lib/                        # Vendor mode: pre-built libraries (Git LFS)
│   ├── include/zvec/c_api.h    # C-API header
│   ├── darwin_arm64/           # macOS ARM64 (.dylib)
│   ├── linux_amd64/            # Linux x64 (.so)
│   ├── linux_arm64/            # Linux ARM64 (.so)
│   └── windows_amd64/          # Windows x64 (.dll + .lib)
├── zvec/                       # Source mode: git submodule → github.com/alibaba/zvec
├── cgo_vendor_*.go             # Vendor mode cgo (default, per platform)
├── cgo_source.go               # Source mode cgo — Unix (-tags source)
├── cgo_source_windows.go       # Source mode cgo — Windows (-tags source)
├── *.go                        # SDK implementation
├── *_test.go                   # Unit tests (tag: integration)
├── *_fuzz_test.go              # Fuzz tests (tag: integration)
├── examples/                   # Usage examples
├── scripts/                    # Build and sync scripts
└── Makefile                    # Build automation
```

### cgo Binding Pattern

All Go files that call C functions use this pattern:

```go
/*
#include "zvec/c_api.h"
#include <stdlib.h>
*/
import "C"
```

The `#cgo CFLAGS` and `#cgo LDFLAGS` are defined in the platform-specific `cgo_vendor_*.go` or `cgo_source.go` files, NOT in the individual `.go` files.

### Error Handling Pattern

All C-API calls return `zvec_error_code_t`. Use `toError()` to convert:

```go
func SomeOperation() error {
    return toError(C.zvec_some_operation(...))
}
```

### Memory Management Pattern

- C strings: `C.CString()` + `defer C.free(unsafe.Pointer(...))`
- C arrays: `unsafe.Slice()` for reading, manual allocation for writing
- Ownership transfer: Set handle to nil after transfer

## Development Commands

```bash
make build-zvec    # Build C-API from submodule
make test          # Run tests (source mode)
make bench         # Run benchmarks
make fuzz          # Run fuzz tests
make lint          # Run linters
make package-libs  # Package vendor libs for current platform
make all           # Full CI check
```

## Testing

Tests require the `integration` build tag (they need the C library):
- Source mode: `go test -tags "source integration" ./...`
- Vendor mode: `go test -tags integration ./...`
