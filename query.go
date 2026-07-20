//go:build cgo && !purego

package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// HNSWQueryParams represents query parameters for HNSW index.
type HNSWQueryParams struct {
	handle *C.zvec_hnsw_query_params_t
}

// NewHNSWQueryParams creates a new HNSW query parameters instance.
func NewHNSWQueryParams(ef int, radius float32, isLinear, isUsingRefiner bool) *HNSWQueryParams {
	handle := C.zvec_query_params_hnsw_create(
		C.int(ef),
		C.float(radius),
		C.bool(isLinear),
		C.bool(isUsingRefiner),
	)
	if handle == nil {
		return nil
	}
	return &HNSWQueryParams{handle: handle}
}

// Destroy releases the HNSW query parameters resources.
func (p *HNSWQueryParams) Destroy() {
	if p.handle != nil {
		C.zvec_query_params_hnsw_destroy(p.handle)
		p.handle = nil
	}
}

// SetEf sets the ef parameter for HNSW query.
func (p *HNSWQueryParams) SetEf(ef int) error {
	defer lockErrorThread()()
	return toError(C.zvec_query_params_hnsw_set_ef(p.handle, C.int(ef)))
}

// GetEf returns the ef parameter.
func (p *HNSWQueryParams) GetEf() int {
	return int(C.zvec_query_params_hnsw_get_ef(p.handle))
}

// IVFQueryParams represents query parameters for IVF index.
type IVFQueryParams struct {
	handle *C.zvec_ivf_query_params_t
}

// NewIVFQueryParams creates a new IVF query parameters instance.
func NewIVFQueryParams(nprobe int, isUsingRefiner bool, scaleFactor float32) *IVFQueryParams {
	handle := C.zvec_query_params_ivf_create(
		C.int(nprobe),
		C.bool(isUsingRefiner),
		C.float(scaleFactor),
	)
	if handle == nil {
		return nil
	}
	return &IVFQueryParams{handle: handle}
}

// Destroy releases the IVF query parameters resources.
func (p *IVFQueryParams) Destroy() {
	if p.handle != nil {
		C.zvec_query_params_ivf_destroy(p.handle)
		p.handle = nil
	}
}

// SetNprobe sets the nprobe parameter for IVF query.
func (p *IVFQueryParams) SetNprobe(nprobe int) error {
	defer lockErrorThread()()
	return toError(C.zvec_query_params_ivf_set_nprobe(p.handle, C.int(nprobe)))
}

// FlatQueryParams represents query parameters for Flat index.
type FlatQueryParams struct {
	handle *C.zvec_flat_query_params_t
}

// NewFlatQueryParams creates a new Flat query parameters instance.
func NewFlatQueryParams(isUsingRefiner bool, scaleFactor float32) *FlatQueryParams {
	handle := C.zvec_query_params_flat_create(
		C.bool(isUsingRefiner),
		C.float(scaleFactor),
	)
	if handle == nil {
		return nil
	}
	return &FlatQueryParams{handle: handle}
}

// DiskANNQueryParams represents query parameters for DiskANN index.
type DiskANNQueryParams struct {
	handle *C.zvec_diskann_query_params_t
}

// NewDiskANNQueryParams creates a new DiskANN query parameters instance.
func NewDiskANNQueryParams(listSize int) *DiskANNQueryParams {
	handle := C.zvec_query_params_diskann_create(C.int(listSize))
	if handle == nil {
		return nil
	}
	return &DiskANNQueryParams{handle: handle}
}

// Destroy releases the Flat query parameters resources.
func (p *FlatQueryParams) Destroy() {
	if p.handle != nil {
		C.zvec_query_params_flat_destroy(p.handle)
		p.handle = nil
	}
}

// Destroy releases the DiskANN query parameters resources.
func (p *DiskANNQueryParams) Destroy() {
	if p.handle != nil {
		C.zvec_query_params_diskann_destroy(p.handle)
		p.handle = nil
	}
}

// SetListSize sets the search frontier size.
func (p *DiskANNQueryParams) SetListSize(listSize int) error {
	return toError(C.zvec_query_params_diskann_set_list_size(p.handle, C.int(listSize)))
}

// GetListSize returns the search frontier size.
func (p *DiskANNQueryParams) GetListSize() int {
	return int(C.zvec_query_params_diskann_get_list_size(p.handle))
}

// SetRadius sets the search radius.
func (p *DiskANNQueryParams) SetRadius(radius float32) error {
	return toError(C.zvec_query_params_diskann_set_radius(p.handle, C.float(radius)))
}

// GetRadius returns the search radius.
func (p *DiskANNQueryParams) GetRadius() float32 {
	return float32(C.zvec_query_params_diskann_get_radius(p.handle))
}

// SetIsLinear sets the linear search mode.
func (p *DiskANNQueryParams) SetIsLinear(isLinear bool) error {
	return toError(C.zvec_query_params_diskann_set_is_linear(p.handle, C.bool(isLinear)))
}

// GetIsLinear returns the linear search mode.
func (p *DiskANNQueryParams) GetIsLinear() bool {
	return bool(C.zvec_query_params_diskann_get_is_linear(p.handle))
}

// SetIsUsingRefiner sets whether to use refiner.
func (p *DiskANNQueryParams) SetIsUsingRefiner(isUsingRefiner bool) error {
	return toError(C.zvec_query_params_diskann_set_is_using_refiner(p.handle, C.bool(isUsingRefiner)))
}

// GetIsUsingRefiner returns whether to use refiner.
func (p *DiskANNQueryParams) GetIsUsingRefiner() bool {
	return bool(C.zvec_query_params_diskann_get_is_using_refiner(p.handle))
}

// FTSQueryParams represents query parameters for FTS index.
type FTSQueryParams struct {
	handle *C.zvec_fts_query_params_t
}

// NewFTSQueryParams creates a new FTS query parameters instance.
// defaultOperator is the boolean operator for adjacent bare terms ("OR" or "AND").
// Pass empty string to use the built-in default.
func NewFTSQueryParams(defaultOperator string) *FTSQueryParams {
	var cOp *C.char
	if defaultOperator != "" {
		cOp = C.CString(defaultOperator)
		defer C.free(unsafe.Pointer(cOp))
	}
	handle := C.zvec_query_params_fts_create(cOp)
	if handle == nil {
		return nil
	}
	return &FTSQueryParams{handle: handle}
}

// Destroy releases the FTS query parameters resources.
func (p *FTSQueryParams) Destroy() {
	if p.handle != nil {
		C.zvec_query_params_fts_destroy(p.handle)
		p.handle = nil
	}
}

// SetDefaultOperator sets the default boolean operator.
func (p *FTSQueryParams) SetDefaultOperator(op string) error {
	cOp := C.CString(op)
	defer C.free(unsafe.Pointer(cOp))
	defer lockErrorThread()()
	return toError(C.zvec_query_params_fts_set_default_operator(p.handle, cOp))
}

// GetDefaultOperator returns the default boolean operator.
func (p *FTSQueryParams) GetDefaultOperator() string {
	return C.GoString(C.zvec_query_params_fts_get_default_operator(p.handle))
}

// SearchQuery represents a vector query operation.
type SearchQuery struct {
	handle *C.zvec_vector_query_t
}

// NewSearchQuery creates a new vector query instance.
func NewSearchQuery() *SearchQuery {
	handle := C.zvec_vector_query_create()
	if handle == nil {
		return nil
	}
	return &SearchQuery{handle: handle}
}

// Destroy releases the vector query resources.
func (q *SearchQuery) Destroy() {
	if q.handle != nil {
		C.zvec_vector_query_destroy(q.handle)
		q.handle = nil
	}
}

// SetFieldName sets the field name for the vector query.
func (q *SearchQuery) SetFieldName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_field_name(q.handle, cName))
}

// GetFieldName returns the field name.
func (q *SearchQuery) GetFieldName() string {
	return C.GoString(C.zvec_vector_query_get_field_name(q.handle))
}

// SetTopK sets the top-k parameter for the query.
func (q *SearchQuery) SetTopK(topk int) error {
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_topk(q.handle, C.int(topk)))
}

// GetTopK returns the top-k parameter.
func (q *SearchQuery) GetTopK() int {
	return int(C.zvec_vector_query_get_topk(q.handle))
}

// SetQueryVector sets the query vector data.
func (q *SearchQuery) SetQueryVector(data []float32) error {
	if len(data) == 0 {
		return &Error{Code: InvalidArgument, Message: "query vector cannot be empty"}
	}
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_query_vector(
		q.handle,
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)*4),
	))
}

// SetFilter sets the filter expression for the query.
func (q *SearchQuery) SetFilter(filter string) error {
	cFilter := C.CString(filter)
	defer C.free(unsafe.Pointer(cFilter))
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_filter(q.handle, cFilter))
}

// GetFilter returns the filter expression.
func (q *SearchQuery) GetFilter() string {
	return C.GoString(C.zvec_vector_query_get_filter(q.handle))
}

// SetIncludeVector sets whether to include vector data in results.
func (q *SearchQuery) SetIncludeVector(include bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_include_vector(q.handle, C.bool(include)))
}

// GetIncludeVector returns whether vector data is included in results.
func (q *SearchQuery) GetIncludeVector() bool {
	return bool(C.zvec_vector_query_get_include_vector(q.handle))
}

// SetIncludeDocID sets whether to include document ID in results.
func (q *SearchQuery) SetIncludeDocID(include bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_include_doc_id(q.handle, C.bool(include)))
}

// GetIncludeDocID returns whether document ID is included in results.
func (q *SearchQuery) GetIncludeDocID() bool {
	return bool(C.zvec_vector_query_get_include_doc_id(q.handle))
}

// SetOutputFields sets the output fields for the query.
func (q *SearchQuery) SetOutputFields(fields []string) error {
	if len(fields) == 0 {
		return nil
	}
	cFields := make([]*C.char, len(fields))
	for i, f := range fields {
		cFields[i] = C.CString(f)
	}
	defer func() {
		for _, cf := range cFields {
			C.free(unsafe.Pointer(cf))
		}
	}()
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_output_fields(
		q.handle,
		(**C.char)(unsafe.Pointer(&cFields[0])),
		C.size_t(len(fields)),
	))
}

// SetHNSWParams sets the HNSW query parameters.
// Note: ownership of params is transferred to the query.
func (q *SearchQuery) SetHNSWParams(params *HNSWQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_vector_query_set_hnsw_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetIVFParams sets the IVF query parameters.
func (q *SearchQuery) SetIVFParams(params *IVFQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_vector_query_set_ivf_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetFlatParams sets the Flat query parameters.
func (q *SearchQuery) SetFlatParams(params *FlatQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_vector_query_set_flat_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetDiskANNParams sets the DiskANN query parameters.
// Ownership of params is transferred to the query on success.
func (q *SearchQuery) SetDiskANNParams(params *DiskANNQueryParams) error {
	err := toError(C.zvec_vector_query_set_diskann_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetFTSParams sets the FTS query parameters.
// Ownership of params is transferred to the query on success.
func (q *SearchQuery) SetFTSParams(params *FTSQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_vector_query_set_fts_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetFTS sets the FTS query payload on this query. The payload is copied;
// the caller retains ownership of fts.
func (q *SearchQuery) SetFTS(fts *FTS) error {
	defer lockErrorThread()()
	return toError(C.zvec_vector_query_set_fts(q.handle, fts.handle))
}

// GetFTS returns an independent copy of the FTS query payload attached to this
// query. The caller owns the result and must destroy it.
func (q *SearchQuery) GetFTS() *FTS {
	source := C.zvec_vector_query_get_fts(q.handle)
	if source == nil {
		return nil
	}
	clone := NewFTS()
	if clone == nil {
		return nil
	}
	if err := clone.SetQueryString(C.GoString(C.zvec_fts_get_query_string(source))); err != nil {
		clone.Destroy()
		return nil
	}
	if err := clone.SetMatchString(C.GoString(C.zvec_fts_get_match_string(source))); err != nil {
		clone.Destroy()
		return nil
	}
	return clone
}

// GroupBySearchQuery represents a group-by vector query operation.
type GroupBySearchQuery struct {
	handle *C.zvec_group_by_vector_query_t
}

// NewGroupBySearchQuery creates a new group-by vector query instance.
func NewGroupBySearchQuery() *GroupBySearchQuery {
	handle := C.zvec_group_by_vector_query_create()
	if handle == nil {
		return nil
	}
	return &GroupBySearchQuery{handle: handle}
}

// Destroy releases the group-by vector query resources.
func (q *GroupBySearchQuery) Destroy() {
	if q.handle != nil {
		C.zvec_group_by_vector_query_destroy(q.handle)
		q.handle = nil
	}
}

// SetFieldName sets the field name for the vector query.
func (q *GroupBySearchQuery) SetFieldName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_field_name(q.handle, cName))
}

// SetGroupByFieldName sets the group-by field name.
func (q *GroupBySearchQuery) SetGroupByFieldName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_group_by_field_name(q.handle, cName))
}

// SetGroupCount sets the group count parameter.
func (q *GroupBySearchQuery) SetGroupCount(count uint32) error {
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_group_count(q.handle, C.uint32_t(count)))
}

// SetTopkPerGroup sets the maximum number of results per group.
func (q *GroupBySearchQuery) SetTopkPerGroup(topk uint32) error {
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_topk_per_group(q.handle, C.uint32_t(topk)))
}

// SetQueryVector sets the query vector data.
func (q *GroupBySearchQuery) SetQueryVector(data []float32) error {
	if len(data) == 0 {
		return &Error{Code: InvalidArgument, Message: "query vector cannot be empty"}
	}
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_query_vector(
		q.handle,
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)*4),
	))
}

// SetFilter sets the filter expression for the query.
func (q *GroupBySearchQuery) SetFilter(filter string) error {
	cFilter := C.CString(filter)
	defer C.free(unsafe.Pointer(cFilter))
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_filter(q.handle, cFilter))
}

// SetIncludeVector sets whether to include vector data in results.
func (q *GroupBySearchQuery) SetIncludeVector(include bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_include_vector(q.handle, C.bool(include)))
}

// SetOutputFields sets the output fields for the query.
func (q *GroupBySearchQuery) SetOutputFields(fields []string) error {
	if len(fields) == 0 {
		return nil
	}
	cFields := make([]*C.char, len(fields))
	for i, f := range fields {
		cFields[i] = C.CString(f)
	}
	defer func() {
		for _, cf := range cFields {
			C.free(unsafe.Pointer(cf))
		}
	}()
	defer lockErrorThread()()
	return toError(C.zvec_group_by_vector_query_set_output_fields(
		q.handle,
		(**C.char)(unsafe.Pointer(&cFields[0])),
		C.size_t(len(fields)),
	))
}

// SetHNSWParams sets the HNSW query parameters.
func (q *GroupBySearchQuery) SetHNSWParams(params *HNSWQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_group_by_vector_query_set_hnsw_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetIVFParams sets the IVF query parameters.
func (q *GroupBySearchQuery) SetIVFParams(params *IVFQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_group_by_vector_query_set_ivf_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetFlatParams sets the Flat query parameters.
func (q *GroupBySearchQuery) SetFlatParams(params *FlatQueryParams) error {
	defer lockErrorThread()()
	err := toError(C.zvec_group_by_vector_query_set_flat_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}

// SetDiskANNParams sets the DiskANN query parameters.
// Ownership of params is transferred to the query on success.
func (q *GroupBySearchQuery) SetDiskANNParams(params *DiskANNQueryParams) error {
	err := toError(C.zvec_group_by_vector_query_set_diskann_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil // ownership transferred
	}
	return err
}
