//go:build cgo && !purego

package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

// IndexParams wraps zvec_index_params_t (opaque pointer).
type IndexParams struct {
	handle *C.zvec_index_params_t
}

// NewIndexParams creates index parameters with the specified index type.
func NewIndexParams(indexType IndexType) *IndexParams {
	handle := C.zvec_index_params_create(C.zvec_index_type_t(indexType))
	if handle == nil {
		return nil
	}
	return &IndexParams{handle: handle}
}

// NewHNSWIndexParams creates HNSW index parameters with the specified metric type and parameters.
func NewHNSWIndexParams(metric MetricType, m, efConstruction int) (*IndexParams, error) {
	params := NewIndexParams(IndexTypeHNSW)
	if params == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create HNSW index params"}
	}
	if err := params.SetMetricType(metric); err != nil {
		params.Destroy()
		return nil, err
	}
	if err := params.SetHNSWParams(m, efConstruction); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

// NewInvertIndexParams creates invert index parameters with the specified options.
func NewInvertIndexParams(enableRangeOpt, enableWildcard bool) (*IndexParams, error) {
	params := NewIndexParams(IndexTypeInvert)
	if params == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create invert index params"}
	}
	if err := params.SetInvertParams(enableRangeOpt, enableWildcard); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

// NewIVFIndexParams creates IVF index parameters with the specified metric type and parameters.
func NewIVFIndexParams(metric MetricType, nList, nIters int, useSoar bool) (*IndexParams, error) {
	params := NewIndexParams(IndexTypeIVF)
	if params == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create IVF index params"}
	}
	if err := params.SetMetricType(metric); err != nil {
		params.Destroy()
		return nil, err
	}
	if err := params.SetIVFParams(nList, nIters, useSoar); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

// NewFlatIndexParams creates Flat index parameters with the specified metric type.
func NewFlatIndexParams(metric MetricType) (*IndexParams, error) {
	params := NewIndexParams(IndexTypeFlat)
	if params == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create flat index params"}
	}
	if err := params.SetMetricType(metric); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

// NewDiskANNIndexParams creates DiskANN index parameters with the specified metric type.
func NewDiskANNIndexParams(metric MetricType, maxDegree, listSize, pqChunkNum int) (*IndexParams, error) {
	params := NewIndexParams(IndexTypeDiskANN)
	if params == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create DiskANN index params"}
	}
	if err := params.SetMetricType(metric); err != nil {
		params.Destroy()
		return nil, err
	}
	if err := params.SetDiskANNParams(maxDegree, listSize, pqChunkNum); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

// NewFTSIndexParams creates FTS index parameters with the specified tokenizer and filters.
func NewFTSIndexParams(tokenizerName string, filters []string, extraParams string) (*IndexParams, error) {
	params := NewIndexParams(IndexTypeFTS)
	if params == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create FTS index params"}
	}
	if err := params.SetFTSParams(tokenizerName, filters, extraParams); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

// Destroy releases the index parameters resources.
func (p *IndexParams) Destroy() {
	if p.handle != nil {
		C.zvec_index_params_destroy(p.handle)
		p.handle = nil
	}
}

// GetType returns the index type.
func (p *IndexParams) GetType() IndexType {
	return IndexType(C.zvec_index_params_get_type(p.handle))
}

// SetMetricType sets the metric type for vector indexes.
func (p *IndexParams) SetMetricType(metric MetricType) error {
	defer lockErrorThread()()
	return toError(C.zvec_index_params_set_metric_type(p.handle, C.zvec_metric_type_t(metric)))
}

// GetMetricType returns the metric type.
func (p *IndexParams) GetMetricType() MetricType {
	return MetricType(C.zvec_index_params_get_metric_type(p.handle))
}

// SetQuantizeType sets the quantize type for vector indexes.
func (p *IndexParams) SetQuantizeType(quantize QuantizeType) error {
	defer lockErrorThread()()
	return toError(C.zvec_index_params_set_quantize_type(p.handle, C.zvec_quantize_type_t(quantize)))
}

// GetQuantizeType returns the quantize type.
func (p *IndexParams) GetQuantizeType() QuantizeType {
	return QuantizeType(C.zvec_index_params_get_quantize_type(p.handle))
}

// SetHNSWParams sets HNSW specific parameters.
func (p *IndexParams) SetHNSWParams(m, efConstruction int) error {
	defer lockErrorThread()()
	return toError(C.zvec_index_params_set_hnsw_params(p.handle, C.int(m), C.int(efConstruction)))
}

// GetHNSWM returns the HNSW m parameter.
func (p *IndexParams) GetHNSWM() int {
	return int(C.zvec_index_params_get_hnsw_m(p.handle))
}

// GetHNSWEfConstruction returns the HNSW ef_construction parameter.
func (p *IndexParams) GetHNSWEfConstruction() int {
	return int(C.zvec_index_params_get_hnsw_ef_construction(p.handle))
}

// SetDiskANNParams sets DiskANN specific parameters.
func (p *IndexParams) SetDiskANNParams(maxDegree, listSize, pqChunkNum int) error {
	return toError(C.zvec_index_params_set_diskann_params(p.handle, C.int(maxDegree), C.int(listSize), C.int(pqChunkNum)))
}

// GetDiskANNMaxDegree returns the DiskANN max_degree parameter.
func (p *IndexParams) GetDiskANNMaxDegree() int {
	return int(C.zvec_index_params_get_diskann_max_degree(p.handle))
}

// GetDiskANNListSize returns the DiskANN list_size parameter.
func (p *IndexParams) GetDiskANNListSize() int {
	return int(C.zvec_index_params_get_diskann_list_size(p.handle))
}

// GetDiskANNPQChunkNum returns the DiskANN pq_chunk_num parameter.
func (p *IndexParams) GetDiskANNPQChunkNum() int {
	return int(C.zvec_index_params_get_diskann_pq_chunk_num(p.handle))
}

// SetIVFParams sets IVF specific parameters.
func (p *IndexParams) SetIVFParams(nList, nIters int, useSoar bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_index_params_set_ivf_params(p.handle, C.int(nList), C.int(nIters), C.bool(useSoar)))
}

// SetInvertParams sets invert index specific parameters.
func (p *IndexParams) SetInvertParams(enableRangeOpt, enableWildcard bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_index_params_set_invert_params(p.handle, C.bool(enableRangeOpt), C.bool(enableWildcard)))
}

// SetFTSParams sets FTS index specific parameters.
func (p *IndexParams) SetFTSParams(tokenizerName string, filters []string, extraParams string) error {
	var cTokenizer *C.char
	if tokenizerName != "" {
		cTokenizer = C.CString(tokenizerName)
		defer C.free(unsafe.Pointer(cTokenizer))
	}
	var cExtra *C.char
	if extraParams != "" {
		cExtra = C.CString(extraParams)
		defer C.free(unsafe.Pointer(cExtra))
	}
	var cFilters *C.zvec_string_array_t
	if len(filters) > 0 {
		cFilters = C.zvec_string_array_create(C.size_t(len(filters)))
		for i, f := range filters {
			cs := C.CString(f)
			C.zvec_string_array_add(cFilters, C.size_t(i), cs)
			C.free(unsafe.Pointer(cs))
		}
		defer C.zvec_string_array_destroy(cFilters)
	}
	defer lockErrorThread()()
	return toError(C.zvec_index_params_set_fts_params(p.handle, cTokenizer, cFilters, cExtra))
}

// GetFTSParams returns FTS index parameters.
func (p *IndexParams) GetFTSParams() (tokenizerName string, filters []string, extraParams string, err error) {
	var cTokenizer *C.char
	var cFilters *C.zvec_string_array_t
	var cExtra *C.char
	defer lockErrorThread()()
	err = toError(C.zvec_index_params_get_fts_params(p.handle, &cTokenizer, &cFilters, &cExtra))
	if err != nil {
		return
	}
	if cTokenizer != nil {
		tokenizerName = C.GoString(cTokenizer)
	}
	if cExtra != nil {
		extraParams = C.GoString(cExtra)
	}
	if cFilters != nil {
		defer C.zvec_string_array_destroy(cFilters)
		count := int(cFilters.count)
		filters = make([]string, count)
		for i := 0; i < count; i++ {
			s := (*C.zvec_string_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cFilters.strings)) + uintptr(i)*unsafe.Sizeof(*cFilters.strings)))
			if s.data != nil {
				filters[i] = C.GoStringN(s.data, C.int(s.length))
			}
		}
	}
	return
}

// FieldSchema wraps zvec_field_schema_t (opaque pointer).
type FieldSchema struct {
	handle          *C.zvec_field_schema_t
	owned           bool
	owner           *CollectionSchema
	ownerGeneration uint64
}

func (f *FieldSchema) validHandle() *C.zvec_field_schema_t {
	if f == nil || f.handle == nil {
		return nil
	}
	if f.owner != nil && (f.owner.handle == nil || f.owner.generation != f.ownerGeneration) {
		f.handle = nil
		f.owner = nil
		return nil
	}
	return f.handle
}

// NewFieldSchema creates a new field schema with the specified parameters.
func NewFieldSchema(name string, dataType DataType, nullable bool, dimension uint32) *FieldSchema {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	handle := C.zvec_field_schema_create(cName, C.zvec_data_type_t(dataType), C.bool(nullable), C.uint32_t(dimension))
	if handle == nil {
		return nil
	}
	return &FieldSchema{handle: handle, owned: true}
}

// Destroy releases the field schema resources if owned.
func (f *FieldSchema) Destroy() {
	if f == nil {
		return
	}
	handle := f.validHandle()
	if handle != nil && f.owned {
		C.zvec_field_schema_destroy(handle)
	}
	f.handle = nil
	f.owner = nil
}

// GetName returns the field name.
func (f *FieldSchema) GetName() string {
	handle := f.validHandle()
	if handle == nil {
		return ""
	}
	return C.GoString(C.zvec_field_schema_get_name(handle))
}

// SetName sets the field name.
func (f *FieldSchema) SetName(name string) error {
	handle := f.validHandle()
	if handle == nil {
		return invalidArgumentError("field schema is no longer valid")
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	defer lockErrorThread()()
	return toError(C.zvec_field_schema_set_name(handle, cName))
}

// GetDataType returns the field data type.
func (f *FieldSchema) GetDataType() DataType {
	handle := f.validHandle()
	if handle == nil {
		return DataTypeUndefined
	}
	return DataType(C.zvec_field_schema_get_data_type(handle))
}

// SetDataType sets the field data type.
func (f *FieldSchema) SetDataType(dataType DataType) error {
	handle := f.validHandle()
	if handle == nil {
		return invalidArgumentError("field schema is no longer valid")
	}
	defer lockErrorThread()()
	return toError(C.zvec_field_schema_set_data_type(handle, C.zvec_data_type_t(dataType)))
}

// IsNullable returns whether the field is nullable.
func (f *FieldSchema) IsNullable() bool {
	handle := f.validHandle()
	return handle != nil && bool(C.zvec_field_schema_is_nullable(handle))
}

// SetNullable sets whether the field is nullable.
func (f *FieldSchema) SetNullable(nullable bool) error {
	handle := f.validHandle()
	if handle == nil {
		return invalidArgumentError("field schema is no longer valid")
	}
	defer lockErrorThread()()
	return toError(C.zvec_field_schema_set_nullable(handle, C.bool(nullable)))
}

// GetDimension returns the field dimension (for vector fields).
func (f *FieldSchema) GetDimension() uint32 {
	handle := f.validHandle()
	if handle == nil {
		return 0
	}
	return uint32(C.zvec_field_schema_get_dimension(handle))
}

// SetDimension sets the field dimension (for vector fields).
func (f *FieldSchema) SetDimension(dimension uint32) error {
	handle := f.validHandle()
	if handle == nil {
		return invalidArgumentError("field schema is no longer valid")
	}
	defer lockErrorThread()()
	return toError(C.zvec_field_schema_set_dimension(handle, C.uint32_t(dimension)))
}

// IsVectorField returns whether the field is a vector field (dense or sparse).
func (f *FieldSchema) IsVectorField() bool {
	handle := f.validHandle()
	return handle != nil && bool(C.zvec_field_schema_is_vector_field(handle))
}

// IsDenseVector returns whether the field is a dense vector field.
func (f *FieldSchema) IsDenseVector() bool {
	handle := f.validHandle()
	return handle != nil && bool(C.zvec_field_schema_is_dense_vector(handle))
}

// IsSparseVector returns whether the field is a sparse vector field.
func (f *FieldSchema) IsSparseVector() bool {
	handle := f.validHandle()
	return handle != nil && bool(C.zvec_field_schema_is_sparse_vector(handle))
}

// HasIndex returns whether the field has an index.
func (f *FieldSchema) HasIndex() bool {
	handle := f.validHandle()
	return handle != nil && bool(C.zvec_field_schema_has_index(handle))
}

// GetIndexType returns the index type of the field.
func (f *FieldSchema) GetIndexType() IndexType {
	handle := f.validHandle()
	if handle == nil {
		return IndexTypeUndefined
	}
	return IndexType(C.zvec_field_schema_get_index_type(handle))
}

// SetIndexParams sets the index parameters for the field.
func (f *FieldSchema) SetIndexParams(params *IndexParams) error {
	handle := f.validHandle()
	if handle == nil {
		return invalidArgumentError("field schema is no longer valid")
	}
	if params == nil || params.handle == nil {
		return invalidArgumentError("index params is nil")
	}
	defer lockErrorThread()()
	return toError(C.zvec_field_schema_set_index_params(handle, params.handle))
}

// CollectionSchema wraps zvec_collection_schema_t (opaque pointer).
type CollectionSchema struct {
	handle     *C.zvec_collection_schema_t
	generation uint64
}

// NewCollectionSchema creates a new collection schema with the specified name.
func NewCollectionSchema(name string) *CollectionSchema {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	handle := C.zvec_collection_schema_create(cName)
	if handle == nil {
		return nil
	}
	return &CollectionSchema{handle: handle}
}

// Destroy releases the collection schema resources.
func (s *CollectionSchema) Destroy() {
	if s != nil && s.handle != nil {
		C.zvec_collection_schema_destroy(s.handle)
		s.handle = nil
		s.generation++
	}
}

// GetName returns the collection name.
func (s *CollectionSchema) GetName() string {
	return C.GoString(C.zvec_collection_schema_get_name(s.handle))
}

// SetName sets the collection name.
func (s *CollectionSchema) SetName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	defer lockErrorThread()()
	return toError(C.zvec_collection_schema_set_name(s.handle, cName))
}

// AddField adds a field to the collection schema.
func (s *CollectionSchema) AddField(field *FieldSchema) error {
	fieldHandle := field.validHandle()
	if fieldHandle == nil {
		return invalidArgumentError("field schema is nil")
	}
	defer lockErrorThread()()
	err := toError(C.zvec_collection_schema_add_field(s.handle, fieldHandle))
	if err == nil {
		s.generation++
	}
	return err
}

// HasField returns whether the collection has a field with the specified name.
func (s *CollectionSchema) HasField(name string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return bool(C.zvec_collection_schema_has_field(s.handle, cName))
}

// GetField returns the field with the specified name (non-owning).
func (s *CollectionSchema) GetField(name string) *FieldSchema {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	handle := C.zvec_collection_schema_get_field(s.handle, cName)
	if handle == nil {
		return nil
	}
	return &FieldSchema{handle: handle, owned: false, owner: s, ownerGeneration: s.generation}
}

// DropField drops a field from the collection schema.
func (s *CollectionSchema) DropField(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	defer lockErrorThread()()
	err := toError(C.zvec_collection_schema_drop_field(s.handle, cName))
	if err == nil {
		s.generation++
	}
	return err
}

// AddIndex adds an index to a field.
func (s *CollectionSchema) AddIndex(fieldName string, params *IndexParams) error {
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))
	defer lockErrorThread()()
	err := toError(C.zvec_collection_schema_add_index(s.handle, cFieldName, params.handle))
	if err == nil {
		s.generation++
	}
	return err
}

// DropIndex drops an index from a field.
func (s *CollectionSchema) DropIndex(fieldName string) error {
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))
	defer lockErrorThread()()
	err := toError(C.zvec_collection_schema_drop_index(s.handle, cFieldName))
	if err == nil {
		s.generation++
	}
	return err
}

// HasIndex returns whether a field has an index.
func (s *CollectionSchema) HasIndex(fieldName string) bool {
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))
	return bool(C.zvec_collection_schema_has_index(s.handle, cFieldName))
}

// SetMaxDocCountPerSegment sets the maximum document count per segment.
func (s *CollectionSchema) SetMaxDocCountPerSegment(count uint64) error {
	defer lockErrorThread()()
	return toError(C.zvec_collection_schema_set_max_doc_count_per_segment(s.handle, C.uint64_t(count)))
}

// GetMaxDocCountPerSegment returns the maximum document count per segment.
func (s *CollectionSchema) GetMaxDocCountPerSegment() uint64 {
	return uint64(C.zvec_collection_schema_get_max_doc_count_per_segment(s.handle))
}
