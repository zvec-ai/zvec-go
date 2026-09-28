//go:build purego || !cgo

package zvec

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"
)

func TestCStringArrayPacksStorage(test *testing.T) {
	values := []string{"short", "", "日本語-key"}
	pointers, storage := cStringArray(values)
	wantSize := len(values)
	for _, value := range values {
		wantSize += len(value)
	}
	if len(pointers) != len(values) || len(storage) != wantSize {
		test.Fatalf("cStringArray() returned %d pointers and %d bytes, want %d", len(pointers), len(storage), wantSize)
	}
	runtime.GC()
	for index, value := range []string{"short", "", "日本語-key"} {
		actual := unsafe.Slice((*byte)(pointers[index]), len(value)+1)
		want := append([]byte(value), 0)
		if !bytes.Equal(actual, want) {
			test.Fatalf("key %d = %q, want %q", index, actual, want)
		}
		if index+1 < len(pointers) && uintptr(pointers[index+1])-uintptr(pointers[index]) != uintptr(len(value)+1) {
			test.Fatalf("key %d is not adjacent to key %d", index, index+1)
		}
	}
	runtime.KeepAlive(storage)
}
