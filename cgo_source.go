//go:build source && !windows

package zvec

// Source mode (Unix): link against the zvec C-API library built from the git submodule.
// This mode is for developers who want to build zvec from source, use a custom
// version, or contribute to zvec-go development.
//
// Usage:
//   1. Build the C-API library: make build-zvec
//   2. Build with source tag:   go build -tags source ./...
//   3. Run tests:               go test -tags "source integration" ./...

/*
#cgo CFLAGS: -I${SRCDIR}/zvec/src/include
#cgo LDFLAGS: -L${SRCDIR}/zvec/build/lib -lzvec_c_api -Wl,-rpath,${SRCDIR}/zvec/build/lib
*/
import "C"
