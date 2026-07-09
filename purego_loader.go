//go:build purego

package zvec

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const zvecLibraryPathEnv = "ZVEC_LIBRARY_PATH"

var (
	puregoLoadOnce sync.Once
	puregoLoadErr  error
	puregoHandle   uintptr
	puregoFns      zvecPuregoAPI
)

type zvecPuregoAPI struct {
	getVersion      func() string
	checkVersion    func(int32, int32, int32) bool
	getVersionMajor func() int32
	getVersionMinor func() int32
	getVersionPatch func() int32

	initialize    func(unsafe.Pointer) int32
	shutdown      func() int32
	isInitialized func() bool
	clearError    func()
	getLastError  func(*unsafe.Pointer) int32
	free          func(unsafe.Pointer)

	indexParamsCreate                func(uint32) unsafe.Pointer
	indexParamsDestroy               func(unsafe.Pointer)
	indexParamsGetType               func(unsafe.Pointer) uint32
	indexParamsSetMetricType         func(unsafe.Pointer, uint32) int32
	indexParamsGetMetricType         func(unsafe.Pointer) uint32
	indexParamsSetQuantizeType       func(unsafe.Pointer, uint32) int32
	indexParamsGetQuantizeType       func(unsafe.Pointer) uint32
	indexParamsSetHNSWParams         func(unsafe.Pointer, int32, int32) int32
	indexParamsGetHNSWM              func(unsafe.Pointer) int32
	indexParamsGetHNSWEfConstruction func(unsafe.Pointer) int32
	indexParamsSetIVFParams          func(unsafe.Pointer, int32, int32, bool) int32
	indexParamsSetInvertParams       func(unsafe.Pointer, bool, bool) int32
	indexParamsSetFTSParams          func(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer, unsafe.Pointer) int32
	indexParamsGetFTSParams          func(unsafe.Pointer, *unsafe.Pointer, *unsafe.Pointer, *unsafe.Pointer) int32
	stringArrayCreate                func(uintptr) unsafe.Pointer
	stringArrayAdd                   func(unsafe.Pointer, uintptr, string)
	stringArrayDestroy               func(unsafe.Pointer)

	fieldSchemaCreate         func(string, uint32, bool, uint32) unsafe.Pointer
	fieldSchemaDestroy        func(unsafe.Pointer)
	fieldSchemaSetName        func(unsafe.Pointer, string) int32
	fieldSchemaGetName        func(unsafe.Pointer) string
	fieldSchemaGetDataType    func(unsafe.Pointer) uint32
	fieldSchemaSetDataType    func(unsafe.Pointer, uint32) int32
	fieldSchemaIsNullable     func(unsafe.Pointer) bool
	fieldSchemaSetNullable    func(unsafe.Pointer, bool) int32
	fieldSchemaGetDimension   func(unsafe.Pointer) uint32
	fieldSchemaSetDimension   func(unsafe.Pointer, uint32) int32
	fieldSchemaIsVectorField  func(unsafe.Pointer) bool
	fieldSchemaIsDenseVector  func(unsafe.Pointer) bool
	fieldSchemaIsSparseVector func(unsafe.Pointer) bool
	fieldSchemaHasIndex       func(unsafe.Pointer) bool
	fieldSchemaGetIndexType   func(unsafe.Pointer) uint32
	fieldSchemaSetIndexParams func(unsafe.Pointer, unsafe.Pointer) int32

	collectionSchemaCreate               func(string) unsafe.Pointer
	collectionSchemaDestroy              func(unsafe.Pointer)
	collectionSchemaGetName              func(unsafe.Pointer) string
	collectionSchemaSetName              func(unsafe.Pointer, string) int32
	collectionSchemaAddField             func(unsafe.Pointer, unsafe.Pointer) int32
	collectionSchemaHasField             func(unsafe.Pointer, string) bool
	collectionSchemaGetField             func(unsafe.Pointer, string) unsafe.Pointer
	collectionSchemaDropField            func(unsafe.Pointer, string) int32
	collectionSchemaAddIndex             func(unsafe.Pointer, string, unsafe.Pointer) int32
	collectionSchemaDropIndex            func(unsafe.Pointer, string) int32
	collectionSchemaHasIndex             func(unsafe.Pointer, string) bool
	collectionSchemaSetMaxDocCountPerSeg func(unsafe.Pointer, uint64) int32
	collectionSchemaGetMaxDocCountPerSeg func(unsafe.Pointer) uint64

	collectionOptionsCreate           func() unsafe.Pointer
	collectionOptionsDestroy          func(unsafe.Pointer)
	collectionOptionsSetEnableMmap    func(unsafe.Pointer, bool) int32
	collectionOptionsGetEnableMmap    func(unsafe.Pointer) bool
	collectionOptionsSetMaxBufferSize func(unsafe.Pointer, uintptr) int32
	collectionOptionsGetMaxBufferSize func(unsafe.Pointer) uintptr
	collectionOptionsSetReadOnly      func(unsafe.Pointer, bool) int32
	collectionOptionsGetReadOnly      func(unsafe.Pointer) bool

	collectionCreateAndOpen func(string, unsafe.Pointer, unsafe.Pointer, *unsafe.Pointer) int32
	collectionOpen          func(string, unsafe.Pointer, *unsafe.Pointer) int32
	collectionClose         func(unsafe.Pointer) int32
	collectionDestroy       func(unsafe.Pointer) int32
	collectionFlush         func(unsafe.Pointer) int32
	collectionInsert        func(unsafe.Pointer, unsafe.Pointer, uintptr, *uintptr, *uintptr) int32
	collectionUpsert        func(unsafe.Pointer, unsafe.Pointer, uintptr, *uintptr, *uintptr) int32
	collectionQuery         func(unsafe.Pointer, unsafe.Pointer, *unsafe.Pointer, *uintptr) int32
	collectionMultiQuery    func(unsafe.Pointer, unsafe.Pointer, *unsafe.Pointer, *uintptr) int32

	docCreate               func() unsafe.Pointer
	docDestroy              func(unsafe.Pointer)
	docClear                func(unsafe.Pointer)
	docAddFieldByValue      func(unsafe.Pointer, string, uint32, unsafe.Pointer, uintptr) int32
	docRemoveField          func(unsafe.Pointer, string) int32
	docSetPK                func(unsafe.Pointer, string)
	docSetDocID             func(unsafe.Pointer, uint64)
	docSetScore             func(unsafe.Pointer, float32)
	docSetOperator          func(unsafe.Pointer, int32)
	docSetFieldNull         func(unsafe.Pointer, string) int32
	docGetDocID             func(unsafe.Pointer) uint64
	docGetScore             func(unsafe.Pointer) float32
	docGetOperator          func(unsafe.Pointer) int32
	docGetFieldCount        func(unsafe.Pointer) uintptr
	docGetPKCopy            func(unsafe.Pointer) unsafe.Pointer
	docGetFieldValueBasic   func(unsafe.Pointer, string, uint32, unsafe.Pointer, uintptr) int32
	docGetFieldValuePointer func(unsafe.Pointer, string, uint32, *unsafe.Pointer, *uintptr) int32
	docIsEmpty              func(unsafe.Pointer) bool
	docHasField             func(unsafe.Pointer, string) bool
	docHasFieldValue        func(unsafe.Pointer, string) bool
	docIsFieldNull          func(unsafe.Pointer, string) bool
	docGetFieldNames        func(unsafe.Pointer, *unsafe.Pointer, *uintptr) int32
	freeStrArray            func(unsafe.Pointer, uintptr)
	docsFree                func(unsafe.Pointer, uintptr)

	hnswQueryParamsCreate   func(int32, float32, bool, bool) unsafe.Pointer
	hnswQueryParamsDestroy  func(unsafe.Pointer)
	hnswQueryParamsSetEf    func(unsafe.Pointer, int32) int32
	hnswQueryParamsGetEf    func(unsafe.Pointer) int32
	ivfQueryParamsCreate    func(int32, bool, float32) unsafe.Pointer
	ivfQueryParamsDestroy   func(unsafe.Pointer)
	ivfQueryParamsSetNprobe func(unsafe.Pointer, int32) int32
	flatQueryParamsCreate   func(bool, float32) unsafe.Pointer
	flatQueryParamsDestroy  func(unsafe.Pointer)
	ftsQueryParamsCreate    func(unsafe.Pointer) unsafe.Pointer
	ftsQueryParamsDestroy   func(unsafe.Pointer)
	ftsQueryParamsSetOp     func(unsafe.Pointer, string) int32
	ftsQueryParamsGetOp     func(unsafe.Pointer) string

	vectorQueryCreate           func() unsafe.Pointer
	vectorQueryDestroy          func(unsafe.Pointer)
	vectorQuerySetTopK          func(unsafe.Pointer, int32) int32
	vectorQueryGetTopK          func(unsafe.Pointer) int32
	vectorQuerySetFieldName     func(unsafe.Pointer, string) int32
	vectorQueryGetFieldName     func(unsafe.Pointer) string
	vectorQuerySetQueryVector   func(unsafe.Pointer, unsafe.Pointer, uintptr) int32
	vectorQuerySetFilter        func(unsafe.Pointer, string) int32
	vectorQueryGetFilter        func(unsafe.Pointer) string
	vectorQuerySetIncludeVector func(unsafe.Pointer, bool) int32
	vectorQueryGetIncludeVector func(unsafe.Pointer) bool
	vectorQuerySetIncludeDocID  func(unsafe.Pointer, bool) int32
	vectorQueryGetIncludeDocID  func(unsafe.Pointer) bool
	vectorQuerySetOutputFields  func(unsafe.Pointer, unsafe.Pointer, uintptr) int32
	vectorQuerySetHNSWParams    func(unsafe.Pointer, unsafe.Pointer) int32
	vectorQuerySetIVFParams     func(unsafe.Pointer, unsafe.Pointer) int32
	vectorQuerySetFlatParams    func(unsafe.Pointer, unsafe.Pointer) int32
	vectorQuerySetFTSParams     func(unsafe.Pointer, unsafe.Pointer) int32
	ftsCreate                   func() unsafe.Pointer
	ftsDestroy                  func(unsafe.Pointer)
	ftsSetQueryString           func(unsafe.Pointer, string) int32
	ftsSetMatchString           func(unsafe.Pointer, string) int32
	ftsGetQueryString           func(unsafe.Pointer) string
	ftsGetMatchString           func(unsafe.Pointer) string
	vectorQuerySetFTS           func(unsafe.Pointer, unsafe.Pointer) int32
	vectorQueryGetFTS           func(unsafe.Pointer) unsafe.Pointer

	multiQueryCreate            func() unsafe.Pointer
	multiQueryDestroy           func(unsafe.Pointer)
	multiQueryAddSubQuery       func(unsafe.Pointer, unsafe.Pointer) int32
	multiQueryGetSubQueryCount  func(unsafe.Pointer) uintptr
	multiQuerySetTopK           func(unsafe.Pointer, int32) int32
	multiQueryGetTopK           func(unsafe.Pointer) int32
	multiQuerySetFilter         func(unsafe.Pointer, string) int32
	multiQueryGetFilter         func(unsafe.Pointer) string
	multiQuerySetIncludeVector  func(unsafe.Pointer, bool) int32
	multiQueryGetIncludeVector  func(unsafe.Pointer) bool
	multiQuerySetOutputFields   func(unsafe.Pointer, unsafe.Pointer, uintptr) int32
	multiQuerySetRerankRRF      func(unsafe.Pointer, int32) int32
	multiQuerySetRerankWeighted func(unsafe.Pointer, unsafe.Pointer, uintptr) int32

	subQueryCreate           func() unsafe.Pointer
	subQueryDestroy          func(unsafe.Pointer)
	subQuerySetNumCandidates func(unsafe.Pointer, int32) int32
	subQueryGetNumCandidates func(unsafe.Pointer) int32
	subQuerySetFieldName     func(unsafe.Pointer, string) int32
	subQueryGetFieldName     func(unsafe.Pointer) string
	subQuerySetQueryVector   func(unsafe.Pointer, unsafe.Pointer, uintptr) int32
	subQuerySetHNSWParams    func(unsafe.Pointer, unsafe.Pointer) int32
	subQuerySetIVFParams     func(unsafe.Pointer, unsafe.Pointer) int32
	subQuerySetFlatParams    func(unsafe.Pointer, unsafe.Pointer) int32
	subQuerySetFTSParams     func(unsafe.Pointer, unsafe.Pointer) int32
	subQuerySetFTS           func(unsafe.Pointer, unsafe.Pointer) int32
}

func puregoAPI() (*zvecPuregoAPI, error) {
	puregoLoadOnce.Do(func() {
		puregoLoadErr = loadPuregoBackend()
	})
	if puregoLoadErr != nil {
		return nil, puregoLoadErr
	}
	return &puregoFns, nil
}

func loadPuregoBackend() error {
	var attempts []string
	for _, candidate := range zvecLibraryCandidates() {
		handle, err := openZvecLibrary(candidate)
		if err != nil {
			attempts = append(attempts, fmt.Sprintf("%s: %v", candidate, err))
			continue
		}
		if err := registerPuregoSymbols(handle); err != nil {
			_ = closeZvecLibrary(handle)
			attempts = append(attempts, fmt.Sprintf("%s: %v", candidate, err))
			continue
		}
		puregoHandle = handle
		return nil
	}
	if len(attempts) == 0 {
		return fmt.Errorf("zvec purego backend: no library candidates for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Errorf("zvec purego backend: failed to load C API library; set %s to the library path; attempts: %s",
		zvecLibraryPathEnv, strings.Join(attempts, "; "))
}

func registerPuregoSymbols(handle uintptr) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("register symbols: %v", r)
		}
	}()

	puregoFns = zvecPuregoAPI{}
	register := func(fn any, name string) {
		purego.RegisterLibFunc(fn, handle, name)
	}

	register(&puregoFns.getVersion, "zvec_get_version")
	register(&puregoFns.checkVersion, "zvec_check_version")
	register(&puregoFns.getVersionMajor, "zvec_get_version_major")
	register(&puregoFns.getVersionMinor, "zvec_get_version_minor")
	register(&puregoFns.getVersionPatch, "zvec_get_version_patch")
	register(&puregoFns.initialize, "zvec_initialize")
	register(&puregoFns.shutdown, "zvec_shutdown")
	register(&puregoFns.isInitialized, "zvec_is_initialized")
	register(&puregoFns.clearError, "zvec_clear_error")
	register(&puregoFns.getLastError, "zvec_get_last_error")
	register(&puregoFns.free, "zvec_free")

	register(&puregoFns.indexParamsCreate, "zvec_index_params_create")
	register(&puregoFns.indexParamsDestroy, "zvec_index_params_destroy")
	register(&puregoFns.indexParamsGetType, "zvec_index_params_get_type")
	register(&puregoFns.indexParamsSetMetricType, "zvec_index_params_set_metric_type")
	register(&puregoFns.indexParamsGetMetricType, "zvec_index_params_get_metric_type")
	register(&puregoFns.indexParamsSetQuantizeType, "zvec_index_params_set_quantize_type")
	register(&puregoFns.indexParamsGetQuantizeType, "zvec_index_params_get_quantize_type")
	register(&puregoFns.indexParamsSetHNSWParams, "zvec_index_params_set_hnsw_params")
	register(&puregoFns.indexParamsGetHNSWM, "zvec_index_params_get_hnsw_m")
	register(&puregoFns.indexParamsGetHNSWEfConstruction, "zvec_index_params_get_hnsw_ef_construction")
	register(&puregoFns.indexParamsSetIVFParams, "zvec_index_params_set_ivf_params")
	register(&puregoFns.indexParamsSetInvertParams, "zvec_index_params_set_invert_params")
	register(&puregoFns.indexParamsSetFTSParams, "zvec_index_params_set_fts_params")
	register(&puregoFns.indexParamsGetFTSParams, "zvec_index_params_get_fts_params")
	register(&puregoFns.stringArrayCreate, "zvec_string_array_create")
	register(&puregoFns.stringArrayAdd, "zvec_string_array_add")
	register(&puregoFns.stringArrayDestroy, "zvec_string_array_destroy")

	register(&puregoFns.fieldSchemaCreate, "zvec_field_schema_create")
	register(&puregoFns.fieldSchemaDestroy, "zvec_field_schema_destroy")
	register(&puregoFns.fieldSchemaSetName, "zvec_field_schema_set_name")
	register(&puregoFns.fieldSchemaGetName, "zvec_field_schema_get_name")
	register(&puregoFns.fieldSchemaGetDataType, "zvec_field_schema_get_data_type")
	register(&puregoFns.fieldSchemaSetDataType, "zvec_field_schema_set_data_type")
	register(&puregoFns.fieldSchemaIsNullable, "zvec_field_schema_is_nullable")
	register(&puregoFns.fieldSchemaSetNullable, "zvec_field_schema_set_nullable")
	register(&puregoFns.fieldSchemaGetDimension, "zvec_field_schema_get_dimension")
	register(&puregoFns.fieldSchemaSetDimension, "zvec_field_schema_set_dimension")
	register(&puregoFns.fieldSchemaIsVectorField, "zvec_field_schema_is_vector_field")
	register(&puregoFns.fieldSchemaIsDenseVector, "zvec_field_schema_is_dense_vector")
	register(&puregoFns.fieldSchemaIsSparseVector, "zvec_field_schema_is_sparse_vector")
	register(&puregoFns.fieldSchemaHasIndex, "zvec_field_schema_has_index")
	register(&puregoFns.fieldSchemaGetIndexType, "zvec_field_schema_get_index_type")
	register(&puregoFns.fieldSchemaSetIndexParams, "zvec_field_schema_set_index_params")

	register(&puregoFns.collectionSchemaCreate, "zvec_collection_schema_create")
	register(&puregoFns.collectionSchemaDestroy, "zvec_collection_schema_destroy")
	register(&puregoFns.collectionSchemaGetName, "zvec_collection_schema_get_name")
	register(&puregoFns.collectionSchemaSetName, "zvec_collection_schema_set_name")
	register(&puregoFns.collectionSchemaAddField, "zvec_collection_schema_add_field")
	register(&puregoFns.collectionSchemaHasField, "zvec_collection_schema_has_field")
	register(&puregoFns.collectionSchemaGetField, "zvec_collection_schema_get_field")
	register(&puregoFns.collectionSchemaDropField, "zvec_collection_schema_drop_field")
	register(&puregoFns.collectionSchemaAddIndex, "zvec_collection_schema_add_index")
	register(&puregoFns.collectionSchemaDropIndex, "zvec_collection_schema_drop_index")
	register(&puregoFns.collectionSchemaHasIndex, "zvec_collection_schema_has_index")
	register(&puregoFns.collectionSchemaSetMaxDocCountPerSeg, "zvec_collection_schema_set_max_doc_count_per_segment")
	register(&puregoFns.collectionSchemaGetMaxDocCountPerSeg, "zvec_collection_schema_get_max_doc_count_per_segment")

	register(&puregoFns.collectionOptionsCreate, "zvec_collection_options_create")
	register(&puregoFns.collectionOptionsDestroy, "zvec_collection_options_destroy")
	register(&puregoFns.collectionOptionsSetEnableMmap, "zvec_collection_options_set_enable_mmap")
	register(&puregoFns.collectionOptionsGetEnableMmap, "zvec_collection_options_get_enable_mmap")
	register(&puregoFns.collectionOptionsSetMaxBufferSize, "zvec_collection_options_set_max_buffer_size")
	register(&puregoFns.collectionOptionsGetMaxBufferSize, "zvec_collection_options_get_max_buffer_size")
	register(&puregoFns.collectionOptionsSetReadOnly, "zvec_collection_options_set_read_only")
	register(&puregoFns.collectionOptionsGetReadOnly, "zvec_collection_options_get_read_only")

	register(&puregoFns.collectionCreateAndOpen, "zvec_collection_create_and_open")
	register(&puregoFns.collectionOpen, "zvec_collection_open")
	register(&puregoFns.collectionClose, "zvec_collection_close")
	register(&puregoFns.collectionDestroy, "zvec_collection_destroy")
	register(&puregoFns.collectionFlush, "zvec_collection_flush")
	register(&puregoFns.collectionInsert, "zvec_collection_insert")
	register(&puregoFns.collectionUpsert, "zvec_collection_upsert")
	register(&puregoFns.collectionQuery, "zvec_collection_query")
	register(&puregoFns.collectionMultiQuery, "zvec_collection_multi_query")

	register(&puregoFns.docCreate, "zvec_doc_create")
	register(&puregoFns.docDestroy, "zvec_doc_destroy")
	register(&puregoFns.docClear, "zvec_doc_clear")
	register(&puregoFns.docAddFieldByValue, "zvec_doc_add_field_by_value")
	register(&puregoFns.docRemoveField, "zvec_doc_remove_field")
	register(&puregoFns.docSetPK, "zvec_doc_set_pk")
	register(&puregoFns.docSetDocID, "zvec_doc_set_doc_id")
	register(&puregoFns.docSetScore, "zvec_doc_set_score")
	register(&puregoFns.docSetOperator, "zvec_doc_set_operator")
	register(&puregoFns.docSetFieldNull, "zvec_doc_set_field_null")
	register(&puregoFns.docGetDocID, "zvec_doc_get_doc_id")
	register(&puregoFns.docGetScore, "zvec_doc_get_score")
	register(&puregoFns.docGetOperator, "zvec_doc_get_operator")
	register(&puregoFns.docGetFieldCount, "zvec_doc_get_field_count")
	register(&puregoFns.docGetPKCopy, "zvec_doc_get_pk_copy")
	register(&puregoFns.docGetFieldValueBasic, "zvec_doc_get_field_value_basic")
	register(&puregoFns.docGetFieldValuePointer, "zvec_doc_get_field_value_pointer")
	register(&puregoFns.docIsEmpty, "zvec_doc_is_empty")
	register(&puregoFns.docHasField, "zvec_doc_has_field")
	register(&puregoFns.docHasFieldValue, "zvec_doc_has_field_value")
	register(&puregoFns.docIsFieldNull, "zvec_doc_is_field_null")
	register(&puregoFns.docGetFieldNames, "zvec_doc_get_field_names")
	register(&puregoFns.freeStrArray, "zvec_free_str_array")
	register(&puregoFns.docsFree, "zvec_docs_free")

	register(&puregoFns.hnswQueryParamsCreate, "zvec_query_params_hnsw_create")
	register(&puregoFns.hnswQueryParamsDestroy, "zvec_query_params_hnsw_destroy")
	register(&puregoFns.hnswQueryParamsSetEf, "zvec_query_params_hnsw_set_ef")
	register(&puregoFns.hnswQueryParamsGetEf, "zvec_query_params_hnsw_get_ef")
	register(&puregoFns.ivfQueryParamsCreate, "zvec_query_params_ivf_create")
	register(&puregoFns.ivfQueryParamsDestroy, "zvec_query_params_ivf_destroy")
	register(&puregoFns.ivfQueryParamsSetNprobe, "zvec_query_params_ivf_set_nprobe")
	register(&puregoFns.flatQueryParamsCreate, "zvec_query_params_flat_create")
	register(&puregoFns.flatQueryParamsDestroy, "zvec_query_params_flat_destroy")
	register(&puregoFns.ftsQueryParamsCreate, "zvec_query_params_fts_create")
	register(&puregoFns.ftsQueryParamsDestroy, "zvec_query_params_fts_destroy")
	register(&puregoFns.ftsQueryParamsSetOp, "zvec_query_params_fts_set_default_operator")
	register(&puregoFns.ftsQueryParamsGetOp, "zvec_query_params_fts_get_default_operator")

	register(&puregoFns.vectorQueryCreate, "zvec_vector_query_create")
	register(&puregoFns.vectorQueryDestroy, "zvec_vector_query_destroy")
	register(&puregoFns.vectorQuerySetTopK, "zvec_vector_query_set_topk")
	register(&puregoFns.vectorQueryGetTopK, "zvec_vector_query_get_topk")
	register(&puregoFns.vectorQuerySetFieldName, "zvec_vector_query_set_field_name")
	register(&puregoFns.vectorQueryGetFieldName, "zvec_vector_query_get_field_name")
	register(&puregoFns.vectorQuerySetQueryVector, "zvec_vector_query_set_query_vector")
	register(&puregoFns.vectorQuerySetFilter, "zvec_vector_query_set_filter")
	register(&puregoFns.vectorQueryGetFilter, "zvec_vector_query_get_filter")
	register(&puregoFns.vectorQuerySetIncludeVector, "zvec_vector_query_set_include_vector")
	register(&puregoFns.vectorQueryGetIncludeVector, "zvec_vector_query_get_include_vector")
	register(&puregoFns.vectorQuerySetIncludeDocID, "zvec_vector_query_set_include_doc_id")
	register(&puregoFns.vectorQueryGetIncludeDocID, "zvec_vector_query_get_include_doc_id")
	register(&puregoFns.vectorQuerySetOutputFields, "zvec_vector_query_set_output_fields")
	register(&puregoFns.vectorQuerySetHNSWParams, "zvec_vector_query_set_hnsw_params")
	register(&puregoFns.vectorQuerySetIVFParams, "zvec_vector_query_set_ivf_params")
	register(&puregoFns.vectorQuerySetFlatParams, "zvec_vector_query_set_flat_params")
	register(&puregoFns.vectorQuerySetFTSParams, "zvec_vector_query_set_fts_params")
	register(&puregoFns.ftsCreate, "zvec_fts_create")
	register(&puregoFns.ftsDestroy, "zvec_fts_destroy")
	register(&puregoFns.ftsSetQueryString, "zvec_fts_set_query_string")
	register(&puregoFns.ftsSetMatchString, "zvec_fts_set_match_string")
	register(&puregoFns.ftsGetQueryString, "zvec_fts_get_query_string")
	register(&puregoFns.ftsGetMatchString, "zvec_fts_get_match_string")
	register(&puregoFns.vectorQuerySetFTS, "zvec_vector_query_set_fts")
	register(&puregoFns.vectorQueryGetFTS, "zvec_vector_query_get_fts")

	register(&puregoFns.multiQueryCreate, "zvec_multi_query_create")
	register(&puregoFns.multiQueryDestroy, "zvec_multi_query_destroy")
	register(&puregoFns.multiQueryAddSubQuery, "zvec_multi_query_add_sub_query")
	register(&puregoFns.multiQueryGetSubQueryCount, "zvec_multi_query_get_sub_query_count")
	register(&puregoFns.multiQuerySetTopK, "zvec_multi_query_set_topk")
	register(&puregoFns.multiQueryGetTopK, "zvec_multi_query_get_topk")
	register(&puregoFns.multiQuerySetFilter, "zvec_multi_query_set_filter")
	register(&puregoFns.multiQueryGetFilter, "zvec_multi_query_get_filter")
	register(&puregoFns.multiQuerySetIncludeVector, "zvec_multi_query_set_include_vector")
	register(&puregoFns.multiQueryGetIncludeVector, "zvec_multi_query_get_include_vector")
	register(&puregoFns.multiQuerySetOutputFields, "zvec_multi_query_set_output_fields")
	register(&puregoFns.multiQuerySetRerankRRF, "zvec_multi_query_set_rerank_rrf")
	register(&puregoFns.multiQuerySetRerankWeighted, "zvec_multi_query_set_rerank_weighted")

	register(&puregoFns.subQueryCreate, "zvec_sub_query_create")
	register(&puregoFns.subQueryDestroy, "zvec_sub_query_destroy")
	register(&puregoFns.subQuerySetNumCandidates, "zvec_sub_query_set_num_candidates")
	register(&puregoFns.subQueryGetNumCandidates, "zvec_sub_query_get_num_candidates")
	register(&puregoFns.subQuerySetFieldName, "zvec_sub_query_set_field_name")
	register(&puregoFns.subQueryGetFieldName, "zvec_sub_query_get_field_name")
	register(&puregoFns.subQuerySetQueryVector, "zvec_sub_query_set_query_vector")
	register(&puregoFns.subQuerySetHNSWParams, "zvec_sub_query_set_hnsw_params")
	register(&puregoFns.subQuerySetIVFParams, "zvec_sub_query_set_ivf_params")
	register(&puregoFns.subQuerySetFlatParams, "zvec_sub_query_set_flat_params")
	register(&puregoFns.subQuerySetFTSParams, "zvec_sub_query_set_fts_params")
	register(&puregoFns.subQuerySetFTS, "zvec_sub_query_set_fts")

	return nil
}

func zvecLibraryCandidates() []string {
	var candidates []string
	add := func(path string) {
		if path == "" {
			return
		}
		for _, existing := range candidates {
			if existing == path {
				return
			}
		}
		candidates = append(candidates, path)
	}
	addPathOrDir := func(path string) {
		if path == "" {
			return
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			for _, name := range zvecLibraryNames() {
				add(filepath.Join(path, name))
			}
			return
		}
		add(path)
	}

	addPathOrDir(os.Getenv(zvecLibraryPathEnv))

	if cwd, err := os.Getwd(); err == nil {
		for _, name := range zvecLibraryNames() {
			add(filepath.Join(cwd, name))
			add(filepath.Join(cwd, "lib", zvecPlatformDir(), name))
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, name := range zvecLibraryNames() {
			add(filepath.Join(exeDir, name))
			add(filepath.Join(exeDir, "lib", zvecPlatformDir(), name))
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Dir(file)
		for _, name := range zvecLibraryNames() {
			add(filepath.Join(root, "lib", zvecPlatformDir(), name))
		}
	}
	for _, name := range zvecLibraryNames() {
		add(name)
	}
	return candidates
}

func zvecPlatformDir() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

func cStringFromPointer(ptr unsafe.Pointer) string {
	if ptr == nil {
		return ""
	}
	n := 0
	for {
		if *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(n))) == 0 {
			break
		}
		n++
	}
	return string(unsafe.Slice((*byte)(ptr), n))
}
