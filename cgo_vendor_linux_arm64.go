//go:build !source && linux && arm64

package zvec

// Vendor mode (default): link against pre-built libraries in lib/ directory.
// This enables "go get" out-of-the-box usage without building zvec from source.
// To use source mode instead, build with: go build -tags source

/*
#cgo CFLAGS: -I${SRCDIR}/lib/include
#cgo LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lzvec_c_api -Wl,-rpath,${SRCDIR}/lib/linux_arm64
*/
import "C"
