# Contributing to zvec-go

Thank you for your interest in contributing to zvec-go! This guide covers the development workflow, project architecture, and best practices.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | ≥ 1.21 | Go SDK development |
| C compiler | gcc / clang | cgo compilation |
| CMake | ≥ 3.20 | Building zvec C-API |
| Ninja | latest | CMake build backend |
| golangci-lint | latest | Go linting (optional) |

## Project Architecture

```
zvec-go/
├── zvec/                  # git submodule → github.com/alibaba/zvec
│   └── src/include/zvec/
│       └── c_api.h        # C-API header (the contract between zvec and zvec-go)
├── zvec.go                # Library initialization, version, config
├── collection.go          # Collection CRUD, insert/update/upsert/delete/query/fetch
├── doc.go                 # Document field operations
├── query.go               # VectorQuery, GroupByVectorQuery, query params (HNSW/IVF/Flat)
├── schema.go              # IndexParams, FieldSchema, CollectionSchema
├── types.go               # Go enum types (DataType, IndexType, MetricType, etc.)
├── errors.go              # Error handling, C→Go error conversion
├── scripts/
│   └── sync-zvec.sh       # Submodule sync with C-API change detection
├── examples/              # Usage examples
├── Makefile               # Build, test, lint automation
└── .github/
    ├── workflows/ci.yml   # CI pipeline (macOS-arm64, linux-x64, linux-arm64)
    └── dependabot.yml     # Auto-update submodule & GitHub Actions
```

### How cgo Bindings Work

The Go SDK wraps the zvec C-API via cgo directives in each `.go` file:

```go
/*
#cgo CFLAGS: -I${SRCDIR}/zvec/src/include
#cgo LDFLAGS: -L${SRCDIR}/zvec/build/lib -lzvec_c_api -Wl,-rpath,${SRCDIR}/zvec/build/lib
#include "zvec/c_api.h"
*/
import "C"
```

- **CFLAGS**: Points to the C-API header in the submodule
- **LDFLAGS**: Links against the built `libzvec_c_api` library
- **rpath**: Ensures the dynamic linker finds the library at runtime

## Getting Started

```bash
# 1. Clone with submodules
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# 2. Build the C-API library
make build-zvec

# 3. Run tests
make test

# 4. Run linter
make lint
```

## Development Workflow

### Daily Development

```bash
# Build + test in one command
make all

# Run tests with verbose output
make test

# Run benchmarks
make bench

# Check code quality
make lint

# Format code
make fmt
```

### When zvec Has C-API Changes

The zvec C-API header (`zvec/src/include/zvec/c_api.h`) is the contract between zvec and zvec-go. When it changes, Go bindings may need updates.

#### Step 1: Check for Changes

```bash
# Check if upstream has C-API changes (without updating)
./scripts/sync-zvec.sh --check-only
```

#### Step 2: Update and Verify

```bash
# Update submodule + rebuild + test
./scripts/sync-zvec.sh --build

# Or update to a specific version
./scripts/sync-zvec.sh v0.5.0 --build
```

#### Step 3: Update Go Bindings (if needed)

If the script reports C-API changes, review the diff and update the corresponding Go files:

| C-API Area | Go File | What to Update |
|------------|---------|----------------|
| `zvec_initialize`, `zvec_shutdown`, `zvec_config_*` | `zvec.go` | Init/config functions |
| `zvec_collection_*` | `collection.go` | Collection operations |
| `zvec_doc_*` | `doc.go` | Document field operations |
| `zvec_vector_query_*`, `zvec_*_query_params_*` | `query.go` | Query types and params |
| `zvec_*_schema_*`, `zvec_index_params_*` | `schema.go` | Schema and index params |
| `zvec_error_code_t` enum values | `errors.go` | Error codes |
| `zvec_data_type_t`, `zvec_index_type_t`, etc. | `types.go` | Enum constants |

#### Step 4: Commit

```bash
git add zvec
git commit -m "chore(deps): update zvec submodule to <version>"
```

### Automated Updates via Dependabot

Dependabot is configured to automatically create PRs when:
- The zvec submodule has new commits (weekly check)
- GitHub Actions have new versions

These PRs will trigger CI, which builds the C-API and runs all tests.

## Adding New C-API Bindings

When a new function is added to `c_api.h`, follow this pattern:

```go
// 1. Add the Go wrapper function with proper documentation
// FunctionName does something useful.
func FunctionName(param string) error {
    // 2. Convert Go types to C types
    cParam := C.CString(param)
    defer C.free(unsafe.Pointer(cParam))

    // 3. Call the C function and convert the error
    return toError(C.zvec_function_name(cParam))
}
```

Key conventions:
- **Always** use `toError()` to convert C error codes
- **Always** free C strings with `defer C.free(unsafe.Pointer(...))`
- **Always** check for nil handles before operations
- Use `unsafe.Slice()` for converting C arrays to Go slices
- Transfer ownership explicitly (set handle to nil after transfer)

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use descriptive variable names (no single-letter names)
- Add godoc comments to all exported types and functions
- Keep cgo blocks minimal — put Go logic outside `import "C"` blocks
- Use `//go:build integration` tag for tests that require the C library

## Running CI Locally

```bash
# Full CI equivalent
make clean build-zvec test bench lint
```

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
