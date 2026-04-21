//go:build source && windows

package zvec

// Source mode (Windows): link against the zvec C-API library built from the git submodule.
// This mode is for developers who want to build zvec from source, use a custom
// version, or contribute to zvec-go development.
//
// On Windows, build with MSVC:
//   1. Build the C-API library: cmake --build zvec/build --target zvec_c_api
//   2. Build with source tag:   go build -tags source ./...
//
// The zvec_c_api.dll must be in the same directory as the executable
// or in a directory listed in the PATH environment variable at runtime.

/*
#cgo CFLAGS: -I${SRCDIR}/zvec/src/include
#cgo LDFLAGS: -L${SRCDIR}/zvec/build/lib -lzvec_c_api
*/
import "C"
