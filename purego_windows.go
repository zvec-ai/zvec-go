//go:build purego && windows

package zvec

import "syscall"

func zvecLibraryNames() []string {
	return []string{"zvec_c_api.dll"}
}

func openZvecLibrary(path string) (uintptr, error) {
	handle, err := syscall.LoadLibrary(path)
	return uintptr(handle), err
}

func closeZvecLibrary(handle uintptr) error {
	return syscall.FreeLibrary(syscall.Handle(handle))
}
