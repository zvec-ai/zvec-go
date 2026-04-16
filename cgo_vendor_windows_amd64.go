//go:build !source && windows && amd64

package zvec

// Vendor mode (default): link against pre-built libraries in lib/ directory.
// This enables "go get" out-of-the-box usage without building zvec from source.
// To use source mode instead, build with: go build -tags source
//
// On Windows, the zvec_c_api.dll must be in the same directory as the executable
// or in a directory listed in the PATH environment variable at runtime.

/*
#cgo CFLAGS: -I${SRCDIR}/lib/include
#cgo LDFLAGS: -L${SRCDIR}/lib/windows_amd64 -lzvec_c_api
*/
import "C"
