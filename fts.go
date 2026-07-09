//go:build !purego

package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// FTS represents an FTS query payload (query string + match string).
type FTS struct {
	handle *C.zvec_fts_t
	owned  bool
}

// NewFTS creates a new FTS query payload.
func NewFTS() *FTS {
	handle := C.zvec_fts_create()
	if handle == nil {
		return nil
	}
	return &FTS{handle: handle, owned: true}
}

// Destroy releases the FTS query payload resources.
// Must not be called on FTS instances returned by SearchQuery.GetFTS().
func (f *FTS) Destroy() {
	if f.handle != nil && f.owned {
		C.zvec_fts_destroy(f.handle)
		f.handle = nil
	}
}

// SetQueryString sets the FTS boolean / advanced query expression.
func (f *FTS) SetQueryString(query string) error {
	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))
	return toError(C.zvec_fts_set_query_string(f.handle, cQuery))
}

// GetQueryString returns the FTS query expression.
func (f *FTS) GetQueryString() string {
	return C.GoString(C.zvec_fts_get_query_string(f.handle))
}

// SetMatchString sets the FTS natural-language match string.
func (f *FTS) SetMatchString(match string) error {
	cMatch := C.CString(match)
	defer C.free(unsafe.Pointer(cMatch))
	return toError(C.zvec_fts_set_match_string(f.handle, cMatch))
}

// GetMatchString returns the FTS match string.
func (f *FTS) GetMatchString() string {
	return C.GoString(C.zvec_fts_get_match_string(f.handle))
}
