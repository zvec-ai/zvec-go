//go:build integration && !windows

package zvec

import "testing"

// testTempDir returns a temporary directory suitable for use with zvec.
// On non-Windows platforms, this simply delegates to t.TempDir().
func testTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
