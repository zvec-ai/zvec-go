//go:build purego || !cgo

package zvec

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

func bindDirectPuregoSymbols(api *zvecPuregoAPI, handle uintptr) error {
	lookup := func(name string) (uintptr, error) {
		address, err := lookupZvecSymbol(handle, name)
		if err != nil {
			return 0, fmt.Errorf("lookup %s: %w", name, err)
		}
		return address, nil
	}

	var err error
	if api.indexParamsCreate, err = bindPointerUint32(lookup, "zvec_index_params_create"); err != nil {
		return err
	}
	if api.indexParamsDestroy, err = bindVoidPointer(lookup, "zvec_index_params_destroy"); err != nil {
		return err
	}
	if api.indexParamsSetMetricType, err = bindInt32PointerUint32(lookup, "zvec_index_params_set_metric_type"); err != nil {
		return err
	}
	if api.indexParamsSetHNSWParams, err = bindInt32PointerInt32Int32(lookup, "zvec_index_params_set_hnsw_params"); err != nil {
		return err
	}

	if api.fieldSchemaCreate, err = bindPointerStringUint32BoolUint32(lookup, "zvec_field_schema_create"); err != nil {
		return err
	}
	if api.fieldSchemaDestroy, err = bindVoidPointer(lookup, "zvec_field_schema_destroy"); err != nil {
		return err
	}
	if api.collectionSchemaCreate, err = bindPointerString(lookup, "zvec_collection_schema_create"); err != nil {
		return err
	}
	if api.collectionSchemaDestroy, err = bindVoidPointer(lookup, "zvec_collection_schema_destroy"); err != nil {
		return err
	}
	if api.collectionInsert, err = bindInt32PointerPointerUintptrOutUintptrOutUintptr(lookup, "zvec_collection_insert"); err != nil {
		return err
	}
	if api.collectionUpdate, err = bindInt32PointerPointerUintptrOutUintptrOutUintptr(lookup, "zvec_collection_update"); err != nil {
		return err
	}
	if api.collectionUpsert, err = bindInt32PointerPointerUintptrOutUintptrOutUintptr(lookup, "zvec_collection_upsert"); err != nil {
		return err
	}
	if api.collectionDelete, err = bindInt32PointerPointerUintptrOutUintptrOutUintptr(lookup, "zvec_collection_delete"); err != nil {
		return err
	}
	if api.collectionQuery, err = bindInt32PointerPointerOutPointerOutUintptr(lookup, "zvec_collection_query"); err != nil {
		return err
	}
	if api.collectionMultiQuery, err = bindInt32PointerPointerOutPointerOutUintptr(lookup, "zvec_collection_multi_query"); err != nil {
		return err
	}
	if api.collectionFetch, err = bindInt32CollectionFetch(lookup, "zvec_collection_fetch"); err != nil {
		return err
	}

	if api.docCreate, err = bindPointer(lookup, "zvec_doc_create"); err != nil {
		return err
	}
	if api.docDestroy, err = bindVoidPointer(lookup, "zvec_doc_destroy"); err != nil {
		return err
	}
	if api.docSetPK, err = bindVoidPointerString(lookup, "zvec_doc_set_pk"); err != nil {
		return err
	}
	if api.docAddFieldByValue, err = bindInt32PointerStringUint32PointerUintptr(lookup, "zvec_doc_add_field_by_value"); err != nil {
		return err
	}
	if api.docGetFieldValuePointer, err = bindInt32PointerStringUint32OutPointerOutUintptr(lookup, "zvec_doc_get_field_value_pointer"); err != nil {
		return err
	}

	if api.vectorQueryCreate, err = bindPointer(lookup, "zvec_vector_query_create"); err != nil {
		return err
	}
	if api.vectorQueryDestroy, err = bindVoidPointer(lookup, "zvec_vector_query_destroy"); err != nil {
		return err
	}
	if api.vectorQuerySetTopK, err = bindInt32PointerInt32(lookup, "zvec_vector_query_set_topk"); err != nil {
		return err
	}
	if api.vectorQuerySetFieldName, err = bindInt32PointerString(lookup, "zvec_vector_query_set_field_name"); err != nil {
		return err
	}
	if api.vectorQuerySetQueryVector, err = bindInt32PointerPointerUintptr(lookup, "zvec_vector_query_set_query_vector"); err != nil {
		return err
	}
	if api.vectorQuerySetFilter, err = bindInt32PointerString(lookup, "zvec_vector_query_set_filter"); err != nil {
		return err
	}

	return nil
}

type symbolLookup func(string) (uintptr, error)

func bindPointer(lookup symbolLookup, name string) (func() unsafe.Pointer, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func() unsafe.Pointer {
		result, _, _ := purego.SyscallN(address)
		return pointerResult(result)
	}, nil
}

func bindPointerUint32(lookup symbolLookup, name string) (func(uint32) unsafe.Pointer, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(value uint32) unsafe.Pointer {
		result, _, _ := purego.SyscallN(address, uintptr(value))
		return pointerResult(result)
	}, nil
}

func bindVoidPointer(lookup symbolLookup, name string) (func(unsafe.Pointer), error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(value unsafe.Pointer) {
		_, _, _ = purego.SyscallN(address, uintptr(value))
		runtime.KeepAlive(value)
	}, nil
}

func bindInt32PointerUint32(lookup symbolLookup, name string) (func(unsafe.Pointer, uint32) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(pointer unsafe.Pointer, value uint32) int32 {
		result, _, _ := purego.SyscallN(address, uintptr(pointer), uintptr(value))
		runtime.KeepAlive(pointer)
		return int32(result)
	}, nil
}

func bindInt32PointerInt32(lookup symbolLookup, name string) (func(unsafe.Pointer, int32) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(pointer unsafe.Pointer, value int32) int32 {
		result, _, _ := purego.SyscallN(address, uintptr(pointer), uintptr(uint32(value)))
		runtime.KeepAlive(pointer)
		return int32(result)
	}, nil
}

func bindInt32PointerInt32Int32(lookup symbolLookup, name string) (func(unsafe.Pointer, int32, int32) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(pointer unsafe.Pointer, first, second int32) int32 {
		result, _, _ := purego.SyscallN(address, uintptr(pointer), uintptr(uint32(first)), uintptr(uint32(second)))
		runtime.KeepAlive(pointer)
		return int32(result)
	}, nil
}

func bindPointerString(lookup symbolLookup, name string) (func(string) unsafe.Pointer, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(value string) unsafe.Pointer {
		cValue := nullTerminatedBytes(value)
		result, _, _ := purego.SyscallN(address, uintptr(unsafe.Pointer(&cValue[0])))
		runtime.KeepAlive(cValue)
		return pointerResult(result)
	}, nil
}

func bindPointerStringUint32BoolUint32(lookup symbolLookup, name string) (func(string, uint32, bool, uint32) unsafe.Pointer, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(value string, first uint32, flag bool, second uint32) unsafe.Pointer {
		cValue := nullTerminatedBytes(value)
		result, _, _ := purego.SyscallN(
			address,
			uintptr(unsafe.Pointer(&cValue[0])),
			uintptr(first),
			boolUintptr(flag),
			uintptr(second),
		)
		runtime.KeepAlive(cValue)
		return pointerResult(result)
	}, nil
}

func bindVoidPointerString(lookup symbolLookup, name string) (func(unsafe.Pointer, string), error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(pointer unsafe.Pointer, value string) {
		cValue := nullTerminatedBytes(value)
		_, _, _ = purego.SyscallN(address, uintptr(pointer), uintptr(unsafe.Pointer(&cValue[0])))
		runtime.KeepAlive(pointer)
		runtime.KeepAlive(cValue)
	}, nil
}

func bindInt32PointerString(lookup symbolLookup, name string) (func(unsafe.Pointer, string) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(pointer unsafe.Pointer, value string) int32 {
		cValue := nullTerminatedBytes(value)
		result, _, _ := purego.SyscallN(address, uintptr(pointer), uintptr(unsafe.Pointer(&cValue[0])))
		runtime.KeepAlive(pointer)
		runtime.KeepAlive(cValue)
		return int32(result)
	}, nil
}

func bindInt32PointerPointerUintptr(lookup symbolLookup, name string) (func(unsafe.Pointer, unsafe.Pointer, uintptr) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(first, second unsafe.Pointer, size uintptr) int32 {
		result, _, _ := purego.SyscallN(address, uintptr(first), uintptr(second), size)
		runtime.KeepAlive(first)
		runtime.KeepAlive(second)
		return int32(result)
	}, nil
}

func bindInt32PointerPointerUintptrOutUintptrOutUintptr(lookup symbolLookup, name string) (func(unsafe.Pointer, unsafe.Pointer, uintptr, *uintptr, *uintptr) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(first, second unsafe.Pointer, count uintptr, successCount, failedCount *uintptr) int32 {
		result, _, _ := purego.SyscallN(
			address,
			uintptr(first),
			uintptr(second),
			count,
			uintptr(unsafe.Pointer(successCount)),
			uintptr(unsafe.Pointer(failedCount)),
		)
		runtime.KeepAlive(first)
		runtime.KeepAlive(second)
		runtime.KeepAlive(successCount)
		runtime.KeepAlive(failedCount)
		return int32(result)
	}, nil
}

func bindInt32PointerPointerOutPointerOutUintptr(lookup symbolLookup, name string) (func(unsafe.Pointer, unsafe.Pointer, *unsafe.Pointer, *uintptr) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(first, second unsafe.Pointer, outPointer *unsafe.Pointer, outCount *uintptr) int32 {
		result, _, _ := purego.SyscallN(
			address,
			uintptr(first),
			uintptr(second),
			uintptr(unsafe.Pointer(outPointer)),
			uintptr(unsafe.Pointer(outCount)),
		)
		runtime.KeepAlive(first)
		runtime.KeepAlive(second)
		runtime.KeepAlive(outPointer)
		runtime.KeepAlive(outCount)
		return int32(result)
	}, nil
}

func bindInt32CollectionFetch(lookup symbolLookup, name string) (func(unsafe.Pointer, unsafe.Pointer, uintptr, unsafe.Pointer, uintptr, bool, *unsafe.Pointer, *uintptr) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(collection, primaryKeys unsafe.Pointer, primaryKeyCount uintptr, fields unsafe.Pointer, fieldCount uintptr, includeVector bool, outPointer *unsafe.Pointer, outCount *uintptr) int32 {
		result, _, _ := purego.SyscallN(
			address,
			uintptr(collection),
			uintptr(primaryKeys),
			primaryKeyCount,
			uintptr(fields),
			fieldCount,
			boolUintptr(includeVector),
			uintptr(unsafe.Pointer(outPointer)),
			uintptr(unsafe.Pointer(outCount)),
		)
		runtime.KeepAlive(collection)
		runtime.KeepAlive(primaryKeys)
		runtime.KeepAlive(fields)
		runtime.KeepAlive(outPointer)
		runtime.KeepAlive(outCount)
		return int32(result)
	}, nil
}

func bindInt32PointerStringUint32PointerUintptr(lookup symbolLookup, name string) (func(unsafe.Pointer, string, uint32, unsafe.Pointer, uintptr) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(first unsafe.Pointer, value string, dataType uint32, second unsafe.Pointer, size uintptr) int32 {
		cValue := nullTerminatedBytes(value)
		result, _, _ := purego.SyscallN(
			address,
			uintptr(first),
			uintptr(unsafe.Pointer(&cValue[0])),
			uintptr(dataType),
			uintptr(second),
			size,
		)
		runtime.KeepAlive(first)
		runtime.KeepAlive(second)
		runtime.KeepAlive(cValue)
		return int32(result)
	}, nil
}

func bindInt32PointerStringUint32OutPointerOutUintptr(lookup symbolLookup, name string) (func(unsafe.Pointer, string, uint32, *unsafe.Pointer, *uintptr) int32, error) {
	address, err := lookup(name)
	if err != nil {
		return nil, err
	}
	return func(pointer unsafe.Pointer, value string, dataType uint32, outPointer *unsafe.Pointer, outSize *uintptr) int32 {
		cValue := nullTerminatedBytes(value)
		result, _, _ := purego.SyscallN(
			address,
			uintptr(pointer),
			uintptr(unsafe.Pointer(&cValue[0])),
			uintptr(dataType),
			uintptr(unsafe.Pointer(outPointer)),
			uintptr(unsafe.Pointer(outSize)),
		)
		runtime.KeepAlive(pointer)
		runtime.KeepAlive(outPointer)
		runtime.KeepAlive(outSize)
		runtime.KeepAlive(cValue)
		return int32(result)
	}, nil
}

func boolUintptr(value bool) uintptr {
	if value {
		return 1
	}
	return 0
}

func pointerResult(value uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&value))
}
