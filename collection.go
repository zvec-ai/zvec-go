//go:build cgo && !purego

package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>

// zvec_go_status_t must be repeated in every preamble that uses it:
// cgo compiles each file's C helpers against that file's preamble only.

typedef struct zvec_go_status {
  zvec_error_code_t code;
  char *err_msg; // malloc'd copy when non-NULL; free with zvec_free (any thread)
} zvec_go_status_t;
// fold each collection operation and its thread-local
// last-error fetch into ONE native call: the Go runtime does not pin a
// goroutine to an OS thread across cgo calls, so a separate zvec_get_last_
// error transition could otherwise observe another call's error state.
// Happy path: err_msg stays NULL and the wrapper costs nothing.

static zvec_go_status_t zvec_go_collection_insert_ex(
    zvec_collection_t *c, const zvec_doc_t **docs, size_t doc_count,
    size_t *success_count, size_t *error_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_insert(c, docs, doc_count, success_count, error_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_update_ex(
    zvec_collection_t *c, const zvec_doc_t **docs, size_t doc_count,
    size_t *success_count, size_t *error_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_update(c, docs, doc_count, success_count, error_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_upsert_ex(
    zvec_collection_t *c, const zvec_doc_t **docs, size_t doc_count,
    size_t *success_count, size_t *error_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_upsert(c, docs, doc_count, success_count, error_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_delete_ex(
    zvec_collection_t *c, const char *const *pks, size_t pk_count,
    size_t *success_count, size_t *error_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_delete(c, pks, pk_count, success_count, error_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_delete_by_filter_ex(
    zvec_collection_t *c, const char *filter) {
  zvec_go_status_t s;
  s.code = zvec_collection_delete_by_filter(c, filter);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_query_ex(
    const zvec_collection_t *c, const zvec_vector_query_t *query,
    zvec_doc_t ***results, size_t *result_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_query(c, query, results, result_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_multi_query_ex(
    const zvec_collection_t *c, const zvec_multi_query_t *query,
    zvec_doc_t ***results, size_t *result_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_multi_query(c, query, results, result_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_fetch_ex(
    zvec_collection_t *c, const char *const *primary_keys, size_t count,
    const char *const *output_fields, size_t output_field_count,
    bool include_vector, zvec_doc_t ***documents, size_t *found_count) {
  zvec_go_status_t s;
  s.code = zvec_collection_fetch(c, primary_keys, count, output_fields,
      output_field_count, include_vector, documents, found_count);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_flush_ex(zvec_collection_t *c) {
  zvec_go_status_t s;
  s.code = zvec_collection_flush(c);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_close_ex(zvec_collection_t *c) {
  zvec_go_status_t s;
  s.code = zvec_collection_close(c);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}
static zvec_go_status_t zvec_go_collection_destroy_ex(zvec_collection_t *c) {
  zvec_go_status_t s;
  s.code = zvec_collection_destroy(c);
  s.err_msg = NULL;
  if (s.code != ZVEC_OK) zvec_get_last_error(&s.err_msg);
  return s;
}

// Batch-destroy result documents in ONE cgo transition instead of N. The
// array itself is NOT freed here (and must not be): FreeDocs passes a
// Go-slice-backed handle array, and the real zvec_docs_free calls free() on
// the array — freeing Go-allocated memory from C would corrupt the heap.
// Passing the Go slice for the duration of the call is legal because it
// contains only C pointers, no Go pointers.
static void zvec_go_docs_destroy(zvec_doc_t **docs, size_t count) {
  for (size_t i = 0; i < count; i++) {
    if (docs[i]) zvec_doc_destroy(docs[i]);
  }
}
*/
import "C"
import (
	"sync/atomic"
	"unsafe"
)

const maxCollectionBufferSize uint64 = 1<<32 - 1

// CollectionOptions represents options for creating or opening a collection.
type CollectionOptions struct {
	handle *C.zvec_collection_options_t
}

// NewCollectionOptions creates a new collection options instance.
func NewCollectionOptions() *CollectionOptions {
	handle := C.zvec_collection_options_create()
	if handle == nil {
		return nil
	}
	return &CollectionOptions{handle: handle}
}

// Destroy releases the collection options resources.
func (o *CollectionOptions) Destroy() {
	if o.handle != nil {
		C.zvec_collection_options_destroy(o.handle)
		o.handle = nil
	}
}

// SetEnableMmap sets whether to enable memory mapping.
func (o *CollectionOptions) SetEnableMmap(enable bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_collection_options_set_enable_mmap(o.handle, C.bool(enable)))
}

// GetEnableMmap returns whether memory mapping is enabled.
func (o *CollectionOptions) GetEnableMmap() bool {
	return bool(C.zvec_collection_options_get_enable_mmap(o.handle))
}

// SetMaxBufferSize sets the maximum buffer size in bytes.
func (o *CollectionOptions) SetMaxBufferSize(size uint64) error {
	if size > maxCollectionBufferSize {
		return invalidArgumentError("max buffer size exceeds the supported uint32 range")
	}
	defer lockErrorThread()()
	return toError(C.zvec_collection_options_set_max_buffer_size(o.handle, C.size_t(size)))
}

// GetMaxBufferSize returns the maximum buffer size in bytes.
func (o *CollectionOptions) GetMaxBufferSize() uint64 {
	return uint64(C.zvec_collection_options_get_max_buffer_size(o.handle))
}

// SetReadOnly sets whether the collection is read-only.
func (o *CollectionOptions) SetReadOnly(readOnly bool) error {
	defer lockErrorThread()()
	return toError(C.zvec_collection_options_set_read_only(o.handle, C.bool(readOnly)))
}

// GetReadOnly returns whether the collection is read-only.
func (o *CollectionOptions) GetReadOnly() bool {
	return bool(C.zvec_collection_options_get_read_only(o.handle))
}

// CollectionStats holds statistics about a collection.
type CollectionStats struct {
	DocCount          uint64
	IndexCount        int
	IndexNames        []string
	IndexCompleteness []float32
}

// WriteResult holds the result of a write operation (insert/update/upsert/delete).
type WriteResult struct {
	SuccessCount uint64
	ErrorCount   uint64
}

// Collection represents a zvec collection.
//
// A Collection is safe for concurrent use of its data-plane operations
// (Insert/Update/Upsert/Delete/Query/Fetch/...) from multiple goroutines.
// Close and Destroy are idempotent and safe to call concurrently with each
// other. However, closing does not wait for in-flight operations, so
// callers must ensure no goroutine is still using the Collection when
// Close or Destroy runs — releasing the native handle underneath a running
// call is undefined behavior that no cheap atomic gate can prevent (a full
// fix would need a reader epoch).
type Collection struct {
	handle *C.zvec_collection_t
	closed atomic.Bool
}

// live returns the native handle for use in an operation, or ErrClosed if
// Close or Destroy has already claimed the collection.
func (c *Collection) live() (*C.zvec_collection_t, error) {
	if c.closed.Load() {
		return nil, ErrClosed
	}
	return c.handle, nil
}

// CreateAndOpen creates a new collection and opens it.
// The caller is responsible for calling Close() when done.
func CreateAndOpen(path string, schema *CollectionSchema, options *CollectionOptions) (*Collection, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cOptions *C.zvec_collection_options_t
	if options != nil {
		cOptions = options.handle
	}

	var cCollection *C.zvec_collection_t
	defer lockErrorThread()()
	err := toError(C.zvec_collection_create_and_open(cPath, schema.handle, cOptions, &cCollection))
	if err != nil {
		return nil, err
	}
	return &Collection{handle: cCollection}, nil
}

// Open opens an existing collection.
// The caller is responsible for calling Close() when done.
func Open(path string, options *CollectionOptions) (*Collection, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cOptions *C.zvec_collection_options_t
	if options != nil {
		cOptions = options.handle
	}

	var cCollection *C.zvec_collection_t
	defer lockErrorThread()()
	err := toError(C.zvec_collection_open(cPath, cOptions, &cCollection))
	if err != nil {
		return nil, err
	}
	return &Collection{handle: cCollection}, nil
}

// Close closes the collection and releases the handle.
// The collection data on disk is preserved and can be reopened with Open().
// Close is idempotent and safe to call concurrently: an atomic
// compare-and-swap gate guarantees exactly one caller performs the
// teardown, eliminating the double-free crash two racing Close calls would
// otherwise cause. Later (or racing loser) calls return nil immediately.
// Operations invoked after close return ErrClosed.
func (c *Collection) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	status := C.zvec_go_collection_close_ex(c.handle)
	c.handle = nil
	return statusError(status)
}

// Destroy destroys the collection data on disk and releases the handle.
// After calling Destroy, the collection data is permanently deleted.
// Note: zvec_collection_destroy deletes data but does not free the handle;
// zvec_collection_close frees the handle (deletes the shared_ptr).
// Destroy is idempotent and safe to call concurrently, for the same
// reason as Close.
func (c *Collection) Destroy() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	destroyStatus := C.zvec_go_collection_destroy_ex(c.handle)
	// close frees the handle memory regardless of destroy result
	C.zvec_collection_close(c.handle)
	c.handle = nil
	return statusError(destroyStatus)
}

// Flush flushes collection data to disk.
func (c *Collection) Flush() error {
	h, err := c.live()
	if err != nil {
		return err
	}
	return statusError(C.zvec_go_collection_flush_ex(h))
}

// GetSchema returns the collection schema.
// The caller is responsible for calling Destroy() on the returned schema.
func (c *Collection) GetSchema() (*CollectionSchema, error) {
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	var cSchema *C.zvec_collection_schema_t
	defer lockErrorThread()()
	err = toError(C.zvec_collection_get_schema(h, &cSchema))
	if err != nil {
		return nil, err
	}
	return &CollectionSchema{handle: cSchema}, nil
}

// GetOptions returns the collection options.
// The caller is responsible for calling Destroy() on the returned options.
func (c *Collection) GetOptions() (*CollectionOptions, error) {
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	var cOptions *C.zvec_collection_options_t
	defer lockErrorThread()()
	err = toError(C.zvec_collection_get_options(h, &cOptions))
	if err != nil {
		return nil, err
	}
	return &CollectionOptions{handle: cOptions}, nil
}

// GetStats returns collection statistics.
func (c *Collection) GetStats() (*CollectionStats, error) {
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	var cStats *C.zvec_collection_stats_t
	defer lockErrorThread()()
	err = toError(C.zvec_collection_get_stats(h, &cStats))
	if err != nil {
		return nil, err
	}
	defer C.zvec_collection_stats_destroy(cStats)

	indexCount := int(C.zvec_collection_stats_get_index_count(cStats))
	indexNames := make([]string, indexCount)
	indexCompleteness := make([]float32, indexCount)
	for i := 0; i < indexCount; i++ {
		cName := C.zvec_collection_stats_get_index_name(cStats, C.size_t(i))
		if cName != nil {
			indexNames[i] = C.GoString(cName)
		}
		indexCompleteness[i] = float32(C.zvec_collection_stats_get_index_completeness(cStats, C.size_t(i)))
	}

	return &CollectionStats{
		DocCount:          uint64(C.zvec_collection_stats_get_doc_count(cStats)),
		IndexCount:        indexCount,
		IndexNames:        indexNames,
		IndexCompleteness: indexCompleteness,
	}, nil
}

// Optimize optimizes the collection (rebuild indexes, merge segments, etc.).
func (c *Collection) Optimize() error {
	h, err := c.live()
	if err != nil {
		return err
	}
	defer lockErrorThread()()
	return toError(C.zvec_collection_optimize(h))
}

// CreateIndex creates an index for a collection field.
func (c *Collection) CreateIndex(fieldName string, params *IndexParams) error {
	h, err := c.live()
	if err != nil {
		return err
	}
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))
	defer lockErrorThread()()
	return toError(C.zvec_collection_create_index(h, cFieldName, params.handle))
}

// DropIndex drops an index from a collection field.
func (c *Collection) DropIndex(fieldName string) error {
	h, err := c.live()
	if err != nil {
		return err
	}
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))
	defer lockErrorThread()()
	return toError(C.zvec_collection_drop_index(h, cFieldName))
}

// AddColumn adds a new column to the collection.
func (c *Collection) AddColumn(fieldSchema *FieldSchema, defaultExpr string) error {
	h, err := c.live()
	if err != nil {
		return err
	}
	fieldHandle := fieldSchema.validHandle()
	if fieldHandle == nil {
		return invalidArgumentError("field schema is no longer valid")
	}
	var cExpr *C.char
	if defaultExpr != "" {
		cExpr = C.CString(defaultExpr)
		defer C.free(unsafe.Pointer(cExpr))
	}
	defer lockErrorThread()()
	return toError(C.zvec_collection_add_column(h, fieldHandle, cExpr))
}

// DropColumn drops a column from the collection.
func (c *Collection) DropColumn(columnName string) error {
	h, err := c.live()
	if err != nil {
		return err
	}
	cColumnName := C.CString(columnName)
	defer C.free(unsafe.Pointer(cColumnName))
	defer lockErrorThread()()
	return toError(C.zvec_collection_drop_column(h, cColumnName))
}

// AlterColumn alters a column in the collection.
// Pass empty string for newName to skip renaming.
// Pass nil for newSchema to skip schema modification.
func (c *Collection) AlterColumn(columnName, newName string, newSchema *FieldSchema) error {
	h, err := c.live()
	if err != nil {
		return err
	}
	cColumnName := C.CString(columnName)
	defer C.free(unsafe.Pointer(cColumnName))
	var cNewName *C.char
	if newName != "" {
		cNewName = C.CString(newName)
		defer C.free(unsafe.Pointer(cNewName))
	}

	var cNewSchema *C.zvec_field_schema_t
	if newSchema != nil {
		cNewSchema = newSchema.validHandle()
		if cNewSchema == nil {
			return invalidArgumentError("new field schema is no longer valid")
		}
	}
	defer lockErrorThread()()

	return toError(C.zvec_collection_alter_column(h, cColumnName, cNewName, cNewSchema))
}

// Insert inserts documents into the collection.
func (c *Collection) Insert(docs []*Doc) (*WriteResult, error) {
	if len(docs) == 0 {
		return &WriteResult{}, nil
	}
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	cDocs := make([]*C.zvec_doc_t, len(docs))
	for i, doc := range docs {
		cDocs[i] = doc.handle
	}
	var successCount, errorCount C.size_t
	status := C.zvec_go_collection_insert_ex(
		h,
		(**C.zvec_doc_t)(unsafe.Pointer(&cDocs[0])),
		C.size_t(len(docs)),
		&successCount,
		&errorCount,
	)
	if err := statusError(status); err != nil {
		return nil, err
	}
	return &WriteResult{
		SuccessCount: uint64(successCount),
		ErrorCount:   uint64(errorCount),
	}, nil
}

// Update updates documents in the collection.
func (c *Collection) Update(docs []*Doc) (*WriteResult, error) {
	if len(docs) == 0 {
		return &WriteResult{}, nil
	}
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	cDocs := make([]*C.zvec_doc_t, len(docs))
	for i, doc := range docs {
		cDocs[i] = doc.handle
	}
	var successCount, errorCount C.size_t
	status := C.zvec_go_collection_update_ex(
		h,
		(**C.zvec_doc_t)(unsafe.Pointer(&cDocs[0])),
		C.size_t(len(docs)),
		&successCount,
		&errorCount,
	)
	if err := statusError(status); err != nil {
		return nil, err
	}
	return &WriteResult{
		SuccessCount: uint64(successCount),
		ErrorCount:   uint64(errorCount),
	}, nil
}

// Upsert inserts or updates documents in the collection.
func (c *Collection) Upsert(docs []*Doc) (*WriteResult, error) {
	if len(docs) == 0 {
		return &WriteResult{}, nil
	}
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	cDocs := make([]*C.zvec_doc_t, len(docs))
	for i, doc := range docs {
		cDocs[i] = doc.handle
	}
	var successCount, errorCount C.size_t
	status := C.zvec_go_collection_upsert_ex(
		h,
		(**C.zvec_doc_t)(unsafe.Pointer(&cDocs[0])),
		C.size_t(len(docs)),
		&successCount,
		&errorCount,
	)
	if err := statusError(status); err != nil {
		return nil, err
	}
	return &WriteResult{
		SuccessCount: uint64(successCount),
		ErrorCount:   uint64(errorCount),
	}, nil
}

// Delete deletes documents by primary keys.
func (c *Collection) Delete(pks []string) (*WriteResult, error) {
	if len(pks) == 0 {
		return &WriteResult{}, nil
	}
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	cPKs, buffer, err := cPrimaryKeyArray(pks)
	if err != nil {
		return nil, err
	}
	defer C.free(buffer)

	var successCount, errorCount C.size_t
	status := C.zvec_go_collection_delete_ex(
		h,
		(**C.char)(unsafe.Pointer(&cPKs[0])),
		C.size_t(len(pks)),
		&successCount,
		&errorCount,
	)
	if err := statusError(status); err != nil {
		return nil, err
	}
	return &WriteResult{
		SuccessCount: uint64(successCount),
		ErrorCount:   uint64(errorCount),
	}, nil
}

// DeleteByFilter deletes documents matching the filter expression.
func (c *Collection) DeleteByFilter(filter string) error {
	h, err := c.live()
	if err != nil {
		return err
	}
	cFilter := C.CString(filter)
	defer C.free(unsafe.Pointer(cFilter))
	return statusError(C.zvec_go_collection_delete_by_filter_ex(h, cFilter))
}

// Query performs a vector similarity search.
// The caller is responsible for calling Destroy() on each returned Doc,
// or using FreeDocs() to free all at once.
func (c *Collection) Query(query *SearchQuery) ([]*Doc, error) {
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	var cResults **C.zvec_doc_t
	var resultCount C.size_t
	status := C.zvec_go_collection_query_ex(h, query.handle, &cResults, &resultCount)
	defer C.zvec_free(unsafe.Pointer(cResults))
	if err := statusError(status); err != nil {
		return nil, err
	}
	return wrapCResultDocs(cResults, resultCount), nil
}

// MultiQuery performs a multi-query search combining multiple sub-queries.
// The caller is responsible for calling Destroy() on each returned Doc,
// or using FreeDocs() to free all at once.
func (c *Collection) MultiQuery(query *MultiQuery) ([]*Doc, error) {
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	var cResults **C.zvec_doc_t
	var resultCount C.size_t
	status := C.zvec_go_collection_multi_query_ex(h, query.handle, &cResults, &resultCount)
	defer C.zvec_free(unsafe.Pointer(cResults))
	if err := statusError(status); err != nil {
		return nil, err
	}
	return wrapCResultDocs(cResults, resultCount), nil
}

func wrapCResultDocs(results **C.zvec_doc_t, resultCount C.size_t) []*Doc {
	count := int(resultCount)
	if count == 0 {
		return nil
	}
	resultSlice := unsafe.Slice(results, count)
	docs := make([]*Doc, count)
	storage := make([]Doc, count)
	for index := range docs {
		storage[index] = Doc{handle: resultSlice[index]}
		docs[index] = &storage[index]
	}
	return docs
}

// FetchOptions controls optional parameters for Fetch.
type FetchOptions struct {
	OutputFields  []string
	IncludeVector bool
}

// Fetch retrieves documents by primary keys.
// Pass nil for opts to use defaults (all fields, no vectors).
// The caller is responsible for calling Destroy() on each returned Doc,
// or using FreeDocs() to free all at once.
func (c *Collection) Fetch(primaryKeys []string, opts *FetchOptions) ([]*Doc, error) {
	if len(primaryKeys) == 0 {
		return nil, nil
	}
	h, err := c.live()
	if err != nil {
		return nil, err
	}
	cPKs, buffer, err := cPrimaryKeyArray(primaryKeys)
	if err != nil {
		return nil, err
	}
	defer C.free(buffer)

	var cOutputFields **C.char
	var outputFieldCount C.size_t
	var includeVector C.bool
	if opts != nil {
		includeVector = C.bool(opts.IncludeVector)
		if len(opts.OutputFields) > 0 {
			fields := make([]*C.char, len(opts.OutputFields))
			for i, f := range opts.OutputFields {
				fields[i] = C.CString(f)
			}
			defer func() {
				for _, cf := range fields {
					C.free(unsafe.Pointer(cf))
				}
			}()
			cOutputFields = (**C.char)(unsafe.Pointer(&fields[0]))
			outputFieldCount = C.size_t(len(opts.OutputFields))
		}
	}

	var cDocs **C.zvec_doc_t
	var foundCount C.size_t
	status := C.zvec_go_collection_fetch_ex(
		h,
		(**C.char)(unsafe.Pointer(&cPKs[0])),
		C.size_t(len(primaryKeys)),
		cOutputFields,
		outputFieldCount,
		includeVector,
		&cDocs,
		&foundCount,
	)
	defer C.zvec_free(unsafe.Pointer(cDocs))
	if err := statusError(status); err != nil {
		return nil, err
	}
	return wrapCResultDocs(cDocs, foundCount), nil
}

func cPrimaryKeyArray(keys []string) ([]*C.char, unsafe.Pointer, error) {
	ptrs := make([]*C.char, len(keys))
	if len(keys) == 1 {
		ptrs[0] = C.CString(keys[0])
		return ptrs, unsafe.Pointer(ptrs[0]), nil
	}
	totalBytes := len(keys)
	maxInt := int(^uint(0) >> 1)
	for _, key := range keys {
		if len(key) > maxInt-totalBytes {
			return nil, nil, &Error{Code: InvalidArgument, Message: "primary keys are too large"}
		}
		totalBytes += len(key)
	}
	buffer := C.malloc(C.size_t(totalBytes))
	if buffer == nil {
		return nil, nil, &Error{Code: ResourceExhausted, Message: "failed to allocate primary key buffer"}
	}
	bytes := unsafe.Slice((*byte)(buffer), totalBytes)
	offset := 0
	for index, key := range keys {
		ptrs[index] = (*C.char)(unsafe.Add(buffer, offset))
		offset += copy(bytes[offset:], key)
		bytes[offset] = 0
		offset++
	}
	return ptrs, buffer, nil
}

// FreeDocs is a convenience function to destroy multiple documents at once.
// All surviving native handles are destroyed in ONE cgo transition via the
// zvec_go_docs_destroy shim instead of one transition per doc. Each Doc's
// handle is atomically claimed first, so calling FreeDocs concurrently or
// repeatedly (or mixing it with per-Doc Destroy) still releases every
// underlying document exactly once.
func FreeDocs(docs []*Doc) {
	if len(docs) == 0 {
		return
	}
	handles := make([]*C.zvec_doc_t, 0, len(docs))
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		h := (*unsafe.Pointer)(unsafe.Pointer(&doc.handle))
		if old := atomic.SwapPointer(h, nil); old != nil {
			handles = append(handles, (*C.zvec_doc_t)(old))
		}
	}
	if len(handles) == 0 {
		return
	}
	C.zvec_go_docs_destroy(&handles[0], C.size_t(len(handles)))
}
