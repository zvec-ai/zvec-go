//go:build (purego || !cgo) && (linux || darwin)

package zvec

import (
	"runtime"

	"github.com/ebitengine/purego"
)

func zvecLibraryNames() []string {
	if runtime.GOOS == "darwin" {
		return []string{"libzvec_c_api.dylib"}
	}
	return []string{"libzvec_c_api.so"}
}

func openZvecLibrary(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func closeZvecLibrary(handle uintptr) error {
	return purego.Dlclose(handle)
}

func lookupZvecSymbol(handle uintptr, name string) (uintptr, error) {
	return purego.Dlsym(handle, name)
}
