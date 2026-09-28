//go:build integration

package zvec

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func benchmarkFTSDiskUsage(benchmark *testing.B, root string) (int64, map[string]int64) {
	byDirectory := make(map[string]int64)
	var total int64
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(relative, string(os.PathSeparator))
		byDirectory[strings.Join(parts[:min(len(parts), 2)], "/")] += info.Size()
		total += info.Size()
		return nil
	})
	if err != nil {
		benchmark.Fatalf("WalkDir() failed: %v", err)
	}
	return total, byDirectory
}
