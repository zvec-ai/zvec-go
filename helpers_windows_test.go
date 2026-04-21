//go:build integration && windows

package zvec

import (
	"syscall"
	"testing"
	"unsafe"
)

// testTempDir returns a temporary directory suitable for use with zvec.
// On Windows, t.TempDir() may return 8.3 short path names (e.g. RUNNER~1)
// which contain '~' that fails zvec's path regex validation.
// This function converts the short path to a long path using GetLongPathNameW.
func testTempDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLongPathNameW := kernel32.NewProc("GetLongPathNameW")

	shortPath, err := syscall.UTF16PtrFromString(tmpDir)
	if err != nil {
		t.Fatalf("failed to convert path to UTF-16: %v", err)
	}

	// First call to get the required buffer size
	requiredSize, _, _ := getLongPathNameW.Call(
		uintptr(unsafe.Pointer(shortPath)),
		0,
		0,
	)
	if requiredSize == 0 {
		// GetLongPathNameW failed; fall back to the original path
		return tmpDir
	}

	longPathBuf := make([]uint16, requiredSize)
	written, _, _ := getLongPathNameW.Call(
		uintptr(unsafe.Pointer(shortPath)),
		uintptr(unsafe.Pointer(&longPathBuf[0])),
		uintptr(requiredSize),
	)
	if written == 0 {
		return tmpDir
	}

	return syscall.UTF16ToString(longPathBuf[:written])
}
