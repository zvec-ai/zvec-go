package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// MultiQuery represents a multi-query operation combining multiple sub-queries.
type MultiQuery struct {
	handle *C.zvec_multi_query_t
}

// NewMultiQuery creates a new multi-query instance.
func NewMultiQuery() *MultiQuery {
	handle := C.zvec_multi_query_create()
	if handle == nil {
		return nil
	}
	return &MultiQuery{handle: handle}
}

// Destroy releases the multi-query resources.
func (q *MultiQuery) Destroy() {
	if q.handle != nil {
		C.zvec_multi_query_destroy(q.handle)
		q.handle = nil
	}
}

// AddSubQuery adds a sub-query to the multi-query.
// The sub-query is copied; the caller retains ownership.
func (q *MultiQuery) AddSubQuery(sub *SubQuery) error {
	return toError(C.zvec_multi_query_add_sub_query(q.handle, sub.handle))
}

// GetSubQueryCount returns the number of sub-queries.
func (q *MultiQuery) GetSubQueryCount() int {
	return int(C.zvec_multi_query_get_sub_query_count(q.handle))
}

// SetTopK sets the top-k parameter for the multi-query.
func (q *MultiQuery) SetTopK(topk int) error {
	return toError(C.zvec_multi_query_set_topk(q.handle, C.int(topk)))
}

// GetTopK returns the top-k parameter.
func (q *MultiQuery) GetTopK() int {
	return int(C.zvec_multi_query_get_topk(q.handle))
}

// SetFilter sets the filter expression for the multi-query.
func (q *MultiQuery) SetFilter(filter string) error {
	cFilter := C.CString(filter)
	defer C.free(unsafe.Pointer(cFilter))
	return toError(C.zvec_multi_query_set_filter(q.handle, cFilter))
}

// GetFilter returns the filter expression.
func (q *MultiQuery) GetFilter() string {
	return C.GoString(C.zvec_multi_query_get_filter(q.handle))
}

// SetIncludeVector sets whether to include vector data in results.
func (q *MultiQuery) SetIncludeVector(include bool) error {
	return toError(C.zvec_multi_query_set_include_vector(q.handle, C.bool(include)))
}

// GetIncludeVector returns whether vector data is included in results.
func (q *MultiQuery) GetIncludeVector() bool {
	return bool(C.zvec_multi_query_get_include_vector(q.handle))
}

// SetOutputFields sets the output fields for the multi-query.
func (q *MultiQuery) SetOutputFields(fields []string) error {
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
	return toError(C.zvec_multi_query_set_output_fields(
		q.handle,
		(**C.char)(unsafe.Pointer(&cFields[0])),
		C.size_t(len(fields)),
	))
}

// SetRerankRRF sets the RRF (Reciprocal Rank Fusion) rerank strategy.
func (q *MultiQuery) SetRerankRRF(rankConstant int) error {
	return toError(C.zvec_multi_query_set_rerank_rrf(q.handle, C.int(rankConstant)))
}

// SetRerankWeighted sets the Weighted rerank strategy with the given per-sub-query weights.
func (q *MultiQuery) SetRerankWeighted(weights []float64) error {
	if len(weights) == 0 {
		return &Error{Code: InvalidArgument, Message: "weights cannot be empty"}
	}
	return toError(C.zvec_multi_query_set_rerank_weighted(
		q.handle,
		(*C.double)(unsafe.Pointer(&weights[0])),
		C.size_t(len(weights)),
	))
}

// SubQuery represents a sub-query within a multi-query.
type SubQuery struct {
	handle *C.zvec_sub_query_t
}

// NewSubQuery creates a new sub-query instance.
func NewSubQuery() *SubQuery {
	handle := C.zvec_sub_query_create()
	if handle == nil {
		return nil
	}
	return &SubQuery{handle: handle}
}

// Destroy releases the sub-query resources.
func (q *SubQuery) Destroy() {
	if q.handle != nil {
		C.zvec_sub_query_destroy(q.handle)
		q.handle = nil
	}
}

// SetNumCandidates sets the number of candidates to retrieve per field.
func (q *SubQuery) SetNumCandidates(n int) error {
	return toError(C.zvec_sub_query_set_num_candidates(q.handle, C.int(n)))
}

// GetNumCandidates returns the number of candidates.
func (q *SubQuery) GetNumCandidates() int {
	return int(C.zvec_sub_query_get_num_candidates(q.handle))
}

// SetFieldName sets the field name for the sub-query.
func (q *SubQuery) SetFieldName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return toError(C.zvec_sub_query_set_field_name(q.handle, cName))
}

// GetFieldName returns the field name.
func (q *SubQuery) GetFieldName() string {
	return C.GoString(C.zvec_sub_query_get_field_name(q.handle))
}

// SetQueryVector sets the query vector data.
func (q *SubQuery) SetQueryVector(data []float32) error {
	if len(data) == 0 {
		return &Error{Code: InvalidArgument, Message: "query vector cannot be empty"}
	}
	return toError(C.zvec_sub_query_set_query_vector(
		q.handle,
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)*4),
	))
}

// SetSparseVector sets the sparse vector indices and values.
func (q *SubQuery) SetSparseVector(indices []uint32, values []float32) error {
	if len(indices) != len(values) {
		return &Error{Code: InvalidArgument, Message: "indices and values must have the same length"}
	}
	if len(indices) == 0 {
		return &Error{Code: InvalidArgument, Message: "sparse vector cannot be empty"}
	}
	return toError(C.zvec_sub_query_set_sparse_vector(
		q.handle,
		(*C.uint32_t)(unsafe.Pointer(&indices[0])),
		(*C.float)(unsafe.Pointer(&values[0])),
		C.size_t(len(indices)),
	))
}

// SetHNSWParams sets the HNSW query parameters.
// Ownership of params is transferred to the sub-query on success.
func (q *SubQuery) SetHNSWParams(params *HNSWQueryParams) error {
	err := toError(C.zvec_sub_query_set_hnsw_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

// SetIVFParams sets the IVF query parameters.
// Ownership of params is transferred to the sub-query on success.
func (q *SubQuery) SetIVFParams(params *IVFQueryParams) error {
	err := toError(C.zvec_sub_query_set_ivf_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

// SetFlatParams sets the Flat query parameters.
// Ownership of params is transferred to the sub-query on success.
func (q *SubQuery) SetFlatParams(params *FlatQueryParams) error {
	err := toError(C.zvec_sub_query_set_flat_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

// SetFTSParams sets the FTS query parameters on the sub-query (takes ownership).
//
// Available since zvec v0.5.1 (c_api: zvec_sub_query_set_fts_params).
// This enables FTS as a sub-query inside a MultiQuery, allowing
// combinations like FTS + Vector rerank via RRF/Weighted.
// Ownership of params is transferred to the sub-query on success.
func (q *SubQuery) SetFTSParams(params *FTSQueryParams) error {
	err := toError(C.zvec_sub_query_set_fts_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

// SetFTS attaches an FTS clause to the sub-query. The clause is copied;
// the caller retains ownership of fts.
//
// Available since zvec v0.5.1 (c_api: zvec_sub_query_set_fts).
func (q *SubQuery) SetFTS(fts *FTS) error {
	return toError(C.zvec_sub_query_set_fts(q.handle, fts.handle))
}

// SetDiskANNParams sets the DiskANN query parameters on the sub-query (takes ownership).
//
// Available since zvec v0.6.0 (c_api: zvec_sub_query_set_diskann_params).
// Ownership of params is transferred to the sub-query on success.
func (q *SubQuery) SetDiskANNParams(params *DiskANNQueryParams) error {
	err := toError(C.zvec_sub_query_set_diskann_params(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}
