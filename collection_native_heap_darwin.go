//go:build integration && darwin && cgo && !purego

package zvec

/*
#include <malloc/malloc.h>

static size_t zvec_native_heap_in_use(void) {
    malloc_statistics_t stats = {0};
    malloc_zone_statistics(malloc_default_zone(), &stats);
    return stats.size_in_use;
}
*/
import "C"

func nativeHeapInUse() uint64 {
	return uint64(C.zvec_native_heap_in_use())
}
