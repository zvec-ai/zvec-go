//go:build purego

package zvec

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// ErrorCode represents a zvec error code.
type ErrorCode int

const (
	OK                 ErrorCode = 0
	NotFound           ErrorCode = 1
	AlreadyExists      ErrorCode = 2
	InvalidArgument    ErrorCode = 3
	PermissionDenied   ErrorCode = 4
	FailedPrecondition ErrorCode = 5
	ResourceExhausted  ErrorCode = 6
	Unavailable        ErrorCode = 7
	InternalError      ErrorCode = 8
	NotSupported       ErrorCode = 9
	Unknown            ErrorCode = 10
)

func (c ErrorCode) String() string {
	switch c {
	case OK:
		return "OK"
	case NotFound:
		return "NotFound"
	case AlreadyExists:
		return "AlreadyExists"
	case InvalidArgument:
		return "InvalidArgument"
	case PermissionDenied:
		return "PermissionDenied"
	case FailedPrecondition:
		return "FailedPrecondition"
	case ResourceExhausted:
		return "ResourceExhausted"
	case Unavailable:
		return "Unavailable"
	case InternalError:
		return "InternalError"
	case NotSupported:
		return "NotSupported"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// Error represents a zvec error with code and message.
type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("zvec error [%s]: %s", e.Code, e.Message)
}

var (
	ErrNotFound           = &Error{Code: NotFound, Message: "resource not found"}
	ErrAlreadyExists      = &Error{Code: AlreadyExists, Message: "resource already exists"}
	ErrInvalidArgument    = &Error{Code: InvalidArgument, Message: "invalid argument"}
	ErrPermissionDenied   = &Error{Code: PermissionDenied, Message: "permission denied"}
	ErrFailedPrecondition = &Error{Code: FailedPrecondition, Message: "failed precondition"}
	ErrResourceExhausted  = &Error{Code: ResourceExhausted, Message: "resource exhausted"}
	ErrUnavailable        = &Error{Code: Unavailable, Message: "unavailable"}
	ErrInternalError      = &Error{Code: InternalError, Message: "internal error"}
	ErrNotSupported       = &Error{Code: NotSupported, Message: "not supported"}
	ErrUnknown            = &Error{Code: Unknown, Message: "unknown error"}
)

func IsNotFound(err error) bool {
	var zvecErr *Error
	return errors.As(err, &zvecErr) && zvecErr.Code == NotFound
}

func IsAlreadyExists(err error) bool {
	var zvecErr *Error
	return errors.As(err, &zvecErr) && zvecErr.Code == AlreadyExists
}

func IsInvalidArgument(err error) bool {
	var zvecErr *Error
	return errors.As(err, &zvecErr) && zvecErr.Code == InvalidArgument
}

func toError(code int32) error {
	if code == int32(OK) {
		return nil
	}

	message := ErrorCode(code).String()
	if puregoFns.getLastError != nil && puregoFns.free != nil {
		var cMsg unsafe.Pointer
		if puregoFns.getLastError(&cMsg) == int32(OK) && cMsg != nil {
			message = cStringFromPointer(cMsg)
			puregoFns.free(cMsg)
		}
	}

	return &Error{Code: ErrorCode(code), Message: message}
}

func unsupportedError(feature string) error {
	return &Error{Code: NotSupported, Message: feature + " is not bound in the purego backend POC"}
}

func invalidArgumentError(message string) error {
	return &Error{Code: InvalidArgument, Message: message}
}

// DataType represents the data type of a field.
type DataType uint32

const (
	DataTypeUndefined        DataType = 0
	DataTypeBinary           DataType = 1
	DataTypeString           DataType = 2
	DataTypeBool             DataType = 3
	DataTypeInt32            DataType = 4
	DataTypeInt64            DataType = 5
	DataTypeUint32           DataType = 6
	DataTypeUint64           DataType = 7
	DataTypeFloat            DataType = 8
	DataTypeDouble           DataType = 9
	DataTypeVectorBinary32   DataType = 20
	DataTypeVectorBinary64   DataType = 21
	DataTypeVectorFP16       DataType = 22
	DataTypeVectorFP32       DataType = 23
	DataTypeVectorFP64       DataType = 24
	DataTypeVectorInt4       DataType = 25
	DataTypeVectorInt8       DataType = 26
	DataTypeVectorInt16      DataType = 27
	DataTypeSparseVectorFP16 DataType = 30
	DataTypeSparseVectorFP32 DataType = 31
	DataTypeArrayBinary      DataType = 40
	DataTypeArrayString      DataType = 41
	DataTypeArrayBool        DataType = 42
	DataTypeArrayInt32       DataType = 43
	DataTypeArrayInt64       DataType = 44
	DataTypeArrayUint32      DataType = 45
	DataTypeArrayUint64      DataType = 46
	DataTypeArrayFloat       DataType = 47
	DataTypeArrayDouble      DataType = 48
)

func (d DataType) String() string {
	switch d {
	case DataTypeUndefined:
		return "Undefined"
	case DataTypeBinary:
		return "Binary"
	case DataTypeString:
		return "String"
	case DataTypeBool:
		return "Bool"
	case DataTypeInt32:
		return "Int32"
	case DataTypeInt64:
		return "Int64"
	case DataTypeUint32:
		return "Uint32"
	case DataTypeUint64:
		return "Uint64"
	case DataTypeFloat:
		return "Float"
	case DataTypeDouble:
		return "Double"
	case DataTypeVectorFP32:
		return "VectorFP32"
	default:
		return "Unknown"
	}
}

// IndexType represents the type of index.
type IndexType uint32

const (
	IndexTypeUndefined IndexType = 0
	IndexTypeHNSW      IndexType = 1
	IndexTypeIVF       IndexType = 2
	IndexTypeFlat      IndexType = 3
	IndexTypeVamana    IndexType = 6
	IndexTypeInvert    IndexType = 10
	IndexTypeFTS       IndexType = 11
)

func (i IndexType) String() string {
	switch i {
	case IndexTypeUndefined:
		return "Undefined"
	case IndexTypeHNSW:
		return "HNSW"
	case IndexTypeIVF:
		return "IVF"
	case IndexTypeFlat:
		return "Flat"
	case IndexTypeVamana:
		return "Vamana"
	case IndexTypeInvert:
		return "Invert"
	case IndexTypeFTS:
		return "FTS"
	default:
		return "Unknown"
	}
}

// MetricType represents the distance metric type.
type MetricType uint32

const (
	MetricTypeUndefined MetricType = 0
	MetricTypeL2        MetricType = 1
	MetricTypeIP        MetricType = 2
	MetricTypeCosine    MetricType = 3
	MetricTypeMIPSL2    MetricType = 4
)

func (m MetricType) String() string {
	switch m {
	case MetricTypeUndefined:
		return "Undefined"
	case MetricTypeL2:
		return "L2"
	case MetricTypeIP:
		return "IP"
	case MetricTypeCosine:
		return "Cosine"
	case MetricTypeMIPSL2:
		return "MIPSL2"
	default:
		return "Unknown"
	}
}

// QuantizeType represents the quantization type.
type QuantizeType uint32

const (
	QuantizeTypeUndefined QuantizeType = 0
	QuantizeTypeFP16      QuantizeType = 1
	QuantizeTypeInt8      QuantizeType = 2
	QuantizeTypeInt4      QuantizeType = 3
)

func (q QuantizeType) String() string {
	switch q {
	case QuantizeTypeUndefined:
		return "Undefined"
	case QuantizeTypeFP16:
		return "FP16"
	case QuantizeTypeInt8:
		return "Int8"
	case QuantizeTypeInt4:
		return "Int4"
	default:
		return "Unknown"
	}
}

// LogLevel represents the log level.
type LogLevel int

const (
	LogLevelDebug LogLevel = 0
	LogLevelInfo  LogLevel = 1
	LogLevelWarn  LogLevel = 2
	LogLevelError LogLevel = 3
	LogLevelFatal LogLevel = 4
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "Debug"
	case LogLevelInfo:
		return "Info"
	case LogLevelWarn:
		return "Warn"
	case LogLevelError:
		return "Error"
	case LogLevelFatal:
		return "Fatal"
	default:
		return "Unknown"
	}
}

// DocOperator represents the document operation type.
type DocOperator int

const (
	DocOpInsert DocOperator = 0
	DocOpUpdate DocOperator = 1
	DocOpUpsert DocOperator = 2
	DocOpDelete DocOperator = 3
)

func (d DocOperator) String() string {
	switch d {
	case DocOpInsert:
		return "Insert"
	case DocOpUpdate:
		return "Update"
	case DocOpUpsert:
		return "Upsert"
	case DocOpDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}

// Initialize initializes the zvec library with optional configuration.
func Initialize(config *ConfigData) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	var cConfig unsafe.Pointer
	if config != nil {
		cConfig = config.handle
	}
	return toError(api.initialize(cConfig))
}

// Shutdown cleans up zvec library resources.
func Shutdown() error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.shutdown())
}

// IsInitialized checks if the library has been initialized.
func IsInitialized() bool {
	api, err := puregoAPI()
	return err == nil && api.isInitialized()
}

// GetVersion returns the library version string.
func GetVersion() string {
	api, err := puregoAPI()
	if err != nil {
		return ""
	}
	return api.getVersion()
}

// CheckVersion checks if the current library version meets the minimum requirements.
func CheckVersion(major, minor, patch int) bool {
	api, err := puregoAPI()
	return err == nil && api.checkVersion(int32(major), int32(minor), int32(patch))
}

func GetVersionMajor() int {
	api, err := puregoAPI()
	if err != nil {
		return 0
	}
	return int(api.getVersionMajor())
}

func GetVersionMinor() int {
	api, err := puregoAPI()
	if err != nil {
		return 0
	}
	return int(api.getVersionMinor())
}

func GetVersionPatch() int {
	api, err := puregoAPI()
	if err != nil {
		return 0
	}
	return int(api.getVersionPatch())
}

func ClearError() {
	api, err := puregoAPI()
	if err == nil {
		api.clearError()
	}
}

// ConfigData represents global zvec configuration.
type ConfigData struct {
	handle unsafe.Pointer
}

func NewConfigData() *ConfigData {
	return nil
}

func (c *ConfigData) Destroy() {
	if c != nil {
		c.handle = nil
	}
}

func (c *ConfigData) SetMemoryLimit(bytes uint64) error {
	return unsupportedError("ConfigData.SetMemoryLimit")
}

func (c *ConfigData) GetMemoryLimit() uint64 {
	return 0
}

func (c *ConfigData) SetQueryThreadCount(count uint32) error {
	return unsupportedError("ConfigData.SetQueryThreadCount")
}

func (c *ConfigData) GetQueryThreadCount() uint32 {
	return 0
}

func (c *ConfigData) SetOptimizeThreadCount(count uint32) error {
	return unsupportedError("ConfigData.SetOptimizeThreadCount")
}

func (c *ConfigData) GetOptimizeThreadCount() uint32 {
	return 0
}

func (c *ConfigData) SetFTSBruteForceByKeysRatio(ratio float32) error {
	return unsupportedError("ConfigData.SetFTSBruteForceByKeysRatio")
}

func (c *ConfigData) GetFTSBruteForceByKeysRatio() float32 {
	return 0
}

func (c *ConfigData) SetJiebaDictDir(dir string) error {
	return unsupportedError("ConfigData.SetJiebaDictDir")
}

func (c *ConfigData) GetJiebaDictDir() string {
	return ""
}

func (c *ConfigData) SetConsoleLog(level LogLevel) error {
	return unsupportedError("ConfigData.SetConsoleLog")
}

func (c *ConfigData) SetFileLog(level LogLevel, dir, basename string, fileSizeMB, overdueDays uint32) error {
	return unsupportedError("ConfigData.SetFileLog")
}

func SetDefaultJiebaDictDir(dir string) {}

func GetDefaultJiebaDictDir() string {
	return ""
}

// IndexParams wraps zvec_index_params_t.
type IndexParams struct {
	handle unsafe.Pointer
}

func newIndexParams(indexType IndexType) (*IndexParams, error) {
	api, err := puregoAPI()
	if err != nil {
		return nil, err
	}
	handle := api.indexParamsCreate(uint32(indexType))
	if handle == nil {
		return nil, &Error{Code: InternalError, Message: "failed to create index params"}
	}
	return &IndexParams{handle: handle}, nil
}

func NewIndexParams(indexType IndexType) *IndexParams {
	params, _ := newIndexParams(indexType)
	return params
}

func NewHNSWIndexParams(metric MetricType, m, efConstruction int) (*IndexParams, error) {
	params, err := newIndexParams(IndexTypeHNSW)
	if err != nil {
		return nil, err
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

func NewInvertIndexParams(enableRangeOpt, enableWildcard bool) (*IndexParams, error) {
	params, err := newIndexParams(IndexTypeInvert)
	if err != nil {
		return nil, err
	}
	if err := params.SetInvertParams(enableRangeOpt, enableWildcard); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

func NewIVFIndexParams(metric MetricType, nList, nIters int, useSoar bool) (*IndexParams, error) {
	params, err := newIndexParams(IndexTypeIVF)
	if err != nil {
		return nil, err
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

func NewFlatIndexParams(metric MetricType) (*IndexParams, error) {
	params, err := newIndexParams(IndexTypeFlat)
	if err != nil {
		return nil, err
	}
	if err := params.SetMetricType(metric); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

func NewFTSIndexParams(tokenizerName string, filters []string, extraParams string) (*IndexParams, error) {
	params, err := newIndexParams(IndexTypeFTS)
	if err != nil {
		return nil, err
	}
	if err := params.SetFTSParams(tokenizerName, filters, extraParams); err != nil {
		params.Destroy()
		return nil, err
	}
	return params, nil
}

func (p *IndexParams) Destroy() {
	if p == nil || p.handle == nil {
		return
	}
	if api, err := puregoAPI(); err == nil {
		api.indexParamsDestroy(p.handle)
	}
	p.handle = nil
}

func (p *IndexParams) GetType() IndexType {
	if p == nil || p.handle == nil {
		return IndexTypeUndefined
	}
	api, err := puregoAPI()
	if err != nil {
		return IndexTypeUndefined
	}
	return IndexType(api.indexParamsGetType(p.handle))
}

func (p *IndexParams) SetMetricType(metric MetricType) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.indexParamsSetMetricType(p.handle, uint32(metric)))
}

func (p *IndexParams) GetMetricType() MetricType {
	api, err := puregoAPI()
	if err != nil || p == nil || p.handle == nil {
		return MetricTypeUndefined
	}
	return MetricType(api.indexParamsGetMetricType(p.handle))
}

func (p *IndexParams) SetQuantizeType(quantize QuantizeType) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.indexParamsSetQuantizeType(p.handle, uint32(quantize)))
}

func (p *IndexParams) GetQuantizeType() QuantizeType {
	api, err := puregoAPI()
	if err != nil || p == nil || p.handle == nil {
		return QuantizeTypeUndefined
	}
	return QuantizeType(api.indexParamsGetQuantizeType(p.handle))
}

func (p *IndexParams) SetHNSWParams(m, efConstruction int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.indexParamsSetHNSWParams(p.handle, int32(m), int32(efConstruction)))
}

func (p *IndexParams) GetHNSWM() int {
	api, err := puregoAPI()
	if err != nil || p == nil || p.handle == nil {
		return 0
	}
	return int(api.indexParamsGetHNSWM(p.handle))
}

func (p *IndexParams) GetHNSWEfConstruction() int {
	api, err := puregoAPI()
	if err != nil || p == nil || p.handle == nil {
		return 0
	}
	return int(api.indexParamsGetHNSWEfConstruction(p.handle))
}

func (p *IndexParams) SetIVFParams(nList, nIters int, useSoar bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.indexParamsSetIVFParams(p.handle, int32(nList), int32(nIters), useSoar))
}

func (p *IndexParams) SetInvertParams(enableRangeOpt, enableWildcard bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.indexParamsSetInvertParams(p.handle, enableRangeOpt, enableWildcard))
}

func (p *IndexParams) SetFTSParams(tokenizerName string, filters []string, extraParams string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	if p == nil || p.handle == nil {
		return invalidArgumentError("index params is nil")
	}

	tokenizerBuf, cTokenizer := optionalCString(tokenizerName)
	extraBuf, cExtra := optionalCString(extraParams)
	var cFilters unsafe.Pointer
	if len(filters) > 0 {
		cFilters = api.stringArrayCreate(uintptr(len(filters)))
		if cFilters == nil {
			return &Error{Code: InternalError, Message: "failed to create FTS filter array"}
		}
		defer api.stringArrayDestroy(cFilters)
		for i, filter := range filters {
			api.stringArrayAdd(cFilters, uintptr(i), filter)
		}
	}

	err = toError(api.indexParamsSetFTSParams(p.handle, cTokenizer, cFilters, cExtra))
	runtime.KeepAlive(tokenizerBuf)
	runtime.KeepAlive(extraBuf)
	return err
}

func (p *IndexParams) GetFTSParams() (tokenizerName string, filters []string, extraParams string, err error) {
	api, apiErr := puregoAPI()
	if apiErr != nil {
		err = apiErr
		return
	}
	if p == nil || p.handle == nil {
		err = invalidArgumentError("index params is nil")
		return
	}

	var cTokenizer, cFilters, cExtra unsafe.Pointer
	err = toError(api.indexParamsGetFTSParams(p.handle, &cTokenizer, &cFilters, &cExtra))
	if err != nil {
		return
	}
	tokenizerName = cStringFromPointer(cTokenizer)
	extraParams = cStringFromPointer(cExtra)
	if cFilters != nil {
		defer api.stringArrayDestroy(cFilters)
		filters = readZvecStringArray(cFilters)
	}
	return
}

// FieldSchema wraps zvec_field_schema_t.
type FieldSchema struct {
	handle unsafe.Pointer
	owned  bool
}

func NewFieldSchema(name string, dataType DataType, nullable bool, dimension uint32) *FieldSchema {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.fieldSchemaCreate(name, uint32(dataType), nullable, dimension)
	if handle == nil {
		return nil
	}
	return &FieldSchema{handle: handle, owned: true}
}

func (f *FieldSchema) Destroy() {
	if f != nil && f.handle != nil && f.owned {
		if api, err := puregoAPI(); err == nil {
			api.fieldSchemaDestroy(f.handle)
		}
		f.handle = nil
	}
}

func (f *FieldSchema) GetName() string {
	api, err := puregoAPI()
	if err != nil || f == nil || f.handle == nil {
		return ""
	}
	return api.fieldSchemaGetName(f.handle)
}

func (f *FieldSchema) SetName(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.fieldSchemaSetName(f.handle, name))
}

func (f *FieldSchema) GetDataType() DataType {
	api, err := puregoAPI()
	if err != nil || f == nil || f.handle == nil {
		return DataTypeUndefined
	}
	return DataType(api.fieldSchemaGetDataType(f.handle))
}

func (f *FieldSchema) SetDataType(dataType DataType) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.fieldSchemaSetDataType(f.handle, uint32(dataType)))
}

func (f *FieldSchema) IsNullable() bool {
	api, err := puregoAPI()
	return err == nil && f != nil && f.handle != nil && api.fieldSchemaIsNullable(f.handle)
}

func (f *FieldSchema) SetNullable(nullable bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.fieldSchemaSetNullable(f.handle, nullable))
}

func (f *FieldSchema) GetDimension() uint32 {
	api, err := puregoAPI()
	if err != nil || f == nil || f.handle == nil {
		return 0
	}
	return api.fieldSchemaGetDimension(f.handle)
}

func (f *FieldSchema) SetDimension(dimension uint32) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.fieldSchemaSetDimension(f.handle, dimension))
}

func (f *FieldSchema) IsVectorField() bool {
	api, err := puregoAPI()
	return err == nil && f != nil && f.handle != nil && api.fieldSchemaIsVectorField(f.handle)
}

func (f *FieldSchema) IsDenseVector() bool {
	api, err := puregoAPI()
	return err == nil && f != nil && f.handle != nil && api.fieldSchemaIsDenseVector(f.handle)
}

func (f *FieldSchema) IsSparseVector() bool {
	api, err := puregoAPI()
	return err == nil && f != nil && f.handle != nil && api.fieldSchemaIsSparseVector(f.handle)
}

func (f *FieldSchema) HasIndex() bool {
	api, err := puregoAPI()
	return err == nil && f != nil && f.handle != nil && api.fieldSchemaHasIndex(f.handle)
}

func (f *FieldSchema) GetIndexType() IndexType {
	api, err := puregoAPI()
	if err != nil || f == nil || f.handle == nil {
		return IndexTypeUndefined
	}
	return IndexType(api.fieldSchemaGetIndexType(f.handle))
}

func (f *FieldSchema) SetIndexParams(params *IndexParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("index params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.fieldSchemaSetIndexParams(f.handle, params.handle))
}

// CollectionSchema wraps zvec_collection_schema_t.
type CollectionSchema struct {
	handle unsafe.Pointer
}

func NewCollectionSchema(name string) *CollectionSchema {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.collectionSchemaCreate(name)
	if handle == nil {
		return nil
	}
	return &CollectionSchema{handle: handle}
}

func (s *CollectionSchema) Destroy() {
	if s != nil && s.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.collectionSchemaDestroy(s.handle)
		}
		s.handle = nil
	}
}

func (s *CollectionSchema) GetName() string {
	api, err := puregoAPI()
	if err != nil || s == nil || s.handle == nil {
		return ""
	}
	return api.collectionSchemaGetName(s.handle)
}

func (s *CollectionSchema) SetName(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionSchemaSetName(s.handle, name))
}

func (s *CollectionSchema) AddField(field *FieldSchema) error {
	if field == nil || field.handle == nil {
		return invalidArgumentError("field schema is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionSchemaAddField(s.handle, field.handle))
}

func (s *CollectionSchema) HasField(name string) bool {
	api, err := puregoAPI()
	return err == nil && s != nil && s.handle != nil && api.collectionSchemaHasField(s.handle, name)
}

func (s *CollectionSchema) GetField(name string) *FieldSchema {
	api, err := puregoAPI()
	if err != nil || s == nil || s.handle == nil {
		return nil
	}
	handle := api.collectionSchemaGetField(s.handle, name)
	if handle == nil {
		return nil
	}
	return &FieldSchema{handle: handle, owned: false}
}

func (s *CollectionSchema) DropField(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionSchemaDropField(s.handle, name))
}

func (s *CollectionSchema) AddIndex(fieldName string, params *IndexParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("index params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionSchemaAddIndex(s.handle, fieldName, params.handle))
}

func (s *CollectionSchema) DropIndex(fieldName string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionSchemaDropIndex(s.handle, fieldName))
}

func (s *CollectionSchema) HasIndex(fieldName string) bool {
	api, err := puregoAPI()
	return err == nil && s != nil && s.handle != nil && api.collectionSchemaHasIndex(s.handle, fieldName)
}

func (s *CollectionSchema) SetMaxDocCountPerSegment(count uint64) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionSchemaSetMaxDocCountPerSeg(s.handle, count))
}

func (s *CollectionSchema) GetMaxDocCountPerSegment() uint64 {
	api, err := puregoAPI()
	if err != nil || s == nil || s.handle == nil {
		return 0
	}
	return api.collectionSchemaGetMaxDocCountPerSeg(s.handle)
}

// CollectionOptions represents options for creating or opening a collection.
type CollectionOptions struct {
	handle unsafe.Pointer
}

func NewCollectionOptions() *CollectionOptions {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.collectionOptionsCreate()
	if handle == nil {
		return nil
	}
	return &CollectionOptions{handle: handle}
}

func (o *CollectionOptions) Destroy() {
	if o != nil && o.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.collectionOptionsDestroy(o.handle)
		}
		o.handle = nil
	}
}

func (o *CollectionOptions) SetEnableMmap(enable bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionOptionsSetEnableMmap(o.handle, enable))
}

func (o *CollectionOptions) GetEnableMmap() bool {
	api, err := puregoAPI()
	return err == nil && o != nil && o.handle != nil && api.collectionOptionsGetEnableMmap(o.handle)
}

func (o *CollectionOptions) SetMaxBufferSize(size uint64) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionOptionsSetMaxBufferSize(o.handle, uintptr(size)))
}

func (o *CollectionOptions) GetMaxBufferSize() uint64 {
	api, err := puregoAPI()
	if err != nil || o == nil || o.handle == nil {
		return 0
	}
	return uint64(api.collectionOptionsGetMaxBufferSize(o.handle))
}

func (o *CollectionOptions) SetReadOnly(readOnly bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionOptionsSetReadOnly(o.handle, readOnly))
}

func (o *CollectionOptions) GetReadOnly() bool {
	api, err := puregoAPI()
	return err == nil && o != nil && o.handle != nil && api.collectionOptionsGetReadOnly(o.handle)
}

// CollectionStats holds statistics about a collection.
type CollectionStats struct {
	DocCount          uint64
	IndexCount        int
	IndexNames        []string
	IndexCompleteness []float32
}

// WriteResult holds the result of a write operation.
type WriteResult struct {
	SuccessCount uint64
	ErrorCount   uint64
}

// Collection represents a zvec collection.
type Collection struct {
	handle unsafe.Pointer
}

func CreateAndOpen(path string, schema *CollectionSchema, options *CollectionOptions) (*Collection, error) {
	if schema == nil || schema.handle == nil {
		return nil, invalidArgumentError("collection schema is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return nil, err
	}
	var cOptions unsafe.Pointer
	if options != nil {
		cOptions = options.handle
	}
	var cCollection unsafe.Pointer
	if err := toError(api.collectionCreateAndOpen(path, schema.handle, cOptions, &cCollection)); err != nil {
		return nil, err
	}
	return &Collection{handle: cCollection}, nil
}

func Open(path string, options *CollectionOptions) (*Collection, error) {
	api, err := puregoAPI()
	if err != nil {
		return nil, err
	}
	var cOptions unsafe.Pointer
	if options != nil {
		cOptions = options.handle
	}
	var cCollection unsafe.Pointer
	if err := toError(api.collectionOpen(path, cOptions, &cCollection)); err != nil {
		return nil, err
	}
	return &Collection{handle: cCollection}, nil
}

func (c *Collection) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.collectionClose(c.handle))
	c.handle = nil
	return err
}

func (c *Collection) Destroy() error {
	if c == nil || c.handle == nil {
		return nil
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	destroyErr := toError(api.collectionDestroy(c.handle))
	_ = api.collectionClose(c.handle)
	c.handle = nil
	return destroyErr
}

func (c *Collection) Flush() error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.collectionFlush(c.handle))
}

func (c *Collection) GetSchema() (*CollectionSchema, error) {
	return nil, unsupportedError("Collection.GetSchema")
}

func (c *Collection) GetOptions() (*CollectionOptions, error) {
	return nil, unsupportedError("Collection.GetOptions")
}

func (c *Collection) GetStats() (*CollectionStats, error) {
	return nil, unsupportedError("Collection.GetStats")
}

func (c *Collection) Optimize() error {
	return unsupportedError("Collection.Optimize")
}

func (c *Collection) CreateIndex(fieldName string, params *IndexParams) error {
	return unsupportedError("Collection.CreateIndex")
}

func (c *Collection) DropIndex(fieldName string) error {
	return unsupportedError("Collection.DropIndex")
}

func (c *Collection) AddColumn(fieldSchema *FieldSchema, defaultExpr string) error {
	return unsupportedError("Collection.AddColumn")
}

func (c *Collection) DropColumn(columnName string) error {
	return unsupportedError("Collection.DropColumn")
}

func (c *Collection) AlterColumn(columnName, newName string, newSchema *FieldSchema) error {
	return unsupportedError("Collection.AlterColumn")
}

func (c *Collection) Insert(docs []*Doc) (*WriteResult, error) {
	return c.writeDocs(docs, puregoFns.collectionInsert)
}

func (c *Collection) Update(docs []*Doc) (*WriteResult, error) {
	if len(docs) == 0 {
		return &WriteResult{}, nil
	}
	return nil, unsupportedError("Collection.Update")
}

func (c *Collection) Upsert(docs []*Doc) (*WriteResult, error) {
	return c.writeDocs(docs, puregoFns.collectionUpsert)
}

func (c *Collection) writeDocs(docs []*Doc, fn func(unsafe.Pointer, unsafe.Pointer, uintptr, *uintptr, *uintptr) int32) (*WriteResult, error) {
	if len(docs) == 0 {
		return &WriteResult{}, nil
	}
	if _, err := puregoAPI(); err != nil {
		return nil, err
	}
	handles := make([]unsafe.Pointer, len(docs))
	for i, doc := range docs {
		if doc == nil || doc.handle == nil {
			return nil, invalidArgumentError("document is nil")
		}
		handles[i] = doc.handle
	}
	var successCount, errorCount uintptr
	if err := toError(fn(c.handle, unsafe.Pointer(&handles[0]), uintptr(len(handles)), &successCount, &errorCount)); err != nil {
		return nil, err
	}
	runtime.KeepAlive(handles)
	return &WriteResult{SuccessCount: uint64(successCount), ErrorCount: uint64(errorCount)}, nil
}

func (c *Collection) Delete(pks []string) (*WriteResult, error) {
	if len(pks) == 0 {
		return &WriteResult{}, nil
	}
	return nil, unsupportedError("Collection.Delete")
}

func (c *Collection) DeleteByFilter(filter string) error {
	return unsupportedError("Collection.DeleteByFilter")
}

func (c *Collection) Query(query *SearchQuery) ([]*Doc, error) {
	if query == nil || query.handle == nil {
		return nil, invalidArgumentError("query is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return nil, err
	}
	var cResults unsafe.Pointer
	var resultCount uintptr
	if err := toError(api.collectionQuery(c.handle, query.handle, &cResults, &resultCount)); err != nil {
		return nil, err
	}
	return wrapDocResults(cResults, resultCount), nil
}

func (c *Collection) MultiQuery(query *MultiQuery) ([]*Doc, error) {
	if query == nil || query.handle == nil {
		return nil, invalidArgumentError("multi query is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return nil, err
	}
	var cResults unsafe.Pointer
	var resultCount uintptr
	if err := toError(api.collectionMultiQuery(c.handle, query.handle, &cResults, &resultCount)); err != nil {
		return nil, err
	}
	return wrapDocResults(cResults, resultCount), nil
}

type FetchOptions struct {
	OutputFields  []string
	IncludeVector bool
}

func (c *Collection) Fetch(primaryKeys []string, opts *FetchOptions) ([]*Doc, error) {
	if len(primaryKeys) == 0 {
		return nil, nil
	}
	return nil, unsupportedError("Collection.Fetch")
}

func wrapDocResults(cResults unsafe.Pointer, count uintptr) []*Doc {
	if cResults == nil || count == 0 {
		return nil
	}
	resultSlice := unsafe.Slice((*unsafe.Pointer)(cResults), int(count))
	docs := make([]*Doc, int(count))
	for i := 0; i < int(count); i++ {
		docs[i] = &Doc{handle: resultSlice[i]}
	}
	if puregoFns.free != nil {
		puregoFns.free(cResults)
	}
	return docs
}

func FreeDocs(docs []*Doc) {
	for _, doc := range docs {
		if doc != nil {
			doc.Destroy()
		}
	}
}

// Doc represents a document in zvec.
type Doc struct {
	handle unsafe.Pointer
}

func NewDoc() *Doc {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.docCreate()
	if handle == nil {
		return nil
	}
	return &Doc{handle: handle}
}

func (d *Doc) Destroy() {
	if d != nil && d.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.docDestroy(d.handle)
		}
		d.handle = nil
	}
}

func (d *Doc) Clear() {
	if d != nil && d.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.docClear(d.handle)
		}
	}
}

func (d *Doc) SetPK(pk string) {
	if d != nil && d.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.docSetPK(d.handle, pk)
		}
	}
}

func (d *Doc) SetDocID(docID uint64) {
	if d != nil && d.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.docSetDocID(d.handle, docID)
		}
	}
}

func (d *Doc) SetScore(score float32) {
	if d != nil && d.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.docSetScore(d.handle, score)
		}
	}
}

func (d *Doc) SetOperator(op DocOperator) {
	if d != nil && d.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.docSetOperator(d.handle, int32(op))
		}
	}
}

func (d *Doc) AddStringField(name, value string) error {
	buf := nullTerminatedBytes(value)
	err := d.addField(name, DataTypeString, unsafe.Pointer(&buf[0]), uintptr(len(value)))
	runtime.KeepAlive(buf)
	return err
}

func (d *Doc) AddBoolField(name string, value bool) error {
	return d.addField(name, DataTypeBool, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddInt32Field(name string, value int32) error {
	return d.addField(name, DataTypeInt32, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddInt64Field(name string, value int64) error {
	return d.addField(name, DataTypeInt64, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddUint32Field(name string, value uint32) error {
	return d.addField(name, DataTypeUint32, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddUint64Field(name string, value uint64) error {
	return d.addField(name, DataTypeUint64, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddFloatField(name string, value float32) error {
	return d.addField(name, DataTypeFloat, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddDoubleField(name string, value float64) error {
	return d.addField(name, DataTypeDouble, unsafe.Pointer(&value), unsafe.Sizeof(value))
}

func (d *Doc) AddVectorFP32Field(name string, vector []float32) error {
	if len(vector) == 0 {
		return invalidArgumentError("vector cannot be empty")
	}
	err := d.addField(name, DataTypeVectorFP32, unsafe.Pointer(&vector[0]), uintptr(len(vector)*4))
	runtime.KeepAlive(vector)
	return err
}

func (d *Doc) AddBinaryField(name string, data []byte) error {
	if len(data) == 0 {
		return invalidArgumentError("binary data cannot be empty")
	}
	err := d.addField(name, DataTypeBinary, unsafe.Pointer(&data[0]), uintptr(len(data)))
	runtime.KeepAlive(data)
	return err
}

func (d *Doc) addField(name string, dataType DataType, value unsafe.Pointer, size uintptr) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	if d == nil || d.handle == nil {
		return invalidArgumentError("document is nil")
	}
	return toError(api.docAddFieldByValue(d.handle, name, uint32(dataType), value, size))
}

func (d *Doc) SetFieldNull(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.docSetFieldNull(d.handle, name))
}

func (d *Doc) RemoveField(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.docRemoveField(d.handle, name))
}

func (d *Doc) GetPK() string {
	api, err := puregoAPI()
	if err != nil || d == nil || d.handle == nil {
		return ""
	}
	cPK := api.docGetPKCopy(d.handle)
	if cPK == nil {
		return ""
	}
	defer api.free(cPK)
	return cStringFromPointer(cPK)
}

func (d *Doc) GetDocID() uint64 {
	api, err := puregoAPI()
	if err != nil || d == nil || d.handle == nil {
		return 0
	}
	return api.docGetDocID(d.handle)
}

func (d *Doc) GetScore() float32 {
	api, err := puregoAPI()
	if err != nil || d == nil || d.handle == nil {
		return 0
	}
	return api.docGetScore(d.handle)
}

func (d *Doc) GetOperator() DocOperator {
	api, err := puregoAPI()
	if err != nil || d == nil || d.handle == nil {
		return DocOpInsert
	}
	return DocOperator(api.docGetOperator(d.handle))
}

func (d *Doc) GetFieldCount() int {
	api, err := puregoAPI()
	if err != nil || d == nil || d.handle == nil {
		return 0
	}
	return int(api.docGetFieldCount(d.handle))
}

func (d *Doc) IsEmpty() bool {
	api, err := puregoAPI()
	return err == nil && d != nil && d.handle != nil && api.docIsEmpty(d.handle)
}

func (d *Doc) GetStringField(name string) (string, error) {
	ptr, size, err := d.getFieldPointer(name, DataTypeString)
	if err != nil {
		return "", err
	}
	if ptr == nil || size == 0 {
		return "", nil
	}
	return string(unsafe.Slice((*byte)(ptr), int(size))), nil
}

func (d *Doc) GetBoolField(name string) (bool, error) {
	var value bool
	err := d.getFieldBasic(name, DataTypeBool, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetInt32Field(name string) (int32, error) {
	var value int32
	err := d.getFieldBasic(name, DataTypeInt32, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetInt64Field(name string) (int64, error) {
	var value int64
	err := d.getFieldBasic(name, DataTypeInt64, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetUint32Field(name string) (uint32, error) {
	var value uint32
	err := d.getFieldBasic(name, DataTypeUint32, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetUint64Field(name string) (uint64, error) {
	var value uint64
	err := d.getFieldBasic(name, DataTypeUint64, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetFloatField(name string) (float32, error) {
	var value float32
	err := d.getFieldBasic(name, DataTypeFloat, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetDoubleField(name string) (float64, error) {
	var value float64
	err := d.getFieldBasic(name, DataTypeDouble, unsafe.Pointer(&value), unsafe.Sizeof(value))
	return value, err
}

func (d *Doc) GetVectorFP32Field(name string) ([]float32, error) {
	ptr, size, err := d.getFieldPointer(name, DataTypeVectorFP32)
	if err != nil {
		return nil, err
	}
	count := int(size) / 4
	values := unsafe.Slice((*float32)(ptr), count)
	out := make([]float32, count)
	copy(out, values)
	return out, nil
}

func (d *Doc) getFieldBasic(name string, fieldType DataType, value unsafe.Pointer, size uintptr) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.docGetFieldValueBasic(d.handle, name, uint32(fieldType), value, size))
}

func (d *Doc) getFieldPointer(name string, fieldType DataType) (unsafe.Pointer, uintptr, error) {
	api, err := puregoAPI()
	if err != nil {
		return nil, 0, err
	}
	var value unsafe.Pointer
	var size uintptr
	if err := toError(api.docGetFieldValuePointer(d.handle, name, uint32(fieldType), &value, &size)); err != nil {
		return nil, 0, err
	}
	return value, size, nil
}

func (d *Doc) HasField(name string) bool {
	api, err := puregoAPI()
	return err == nil && d != nil && d.handle != nil && api.docHasField(d.handle, name)
}

func (d *Doc) HasFieldValue(name string) bool {
	api, err := puregoAPI()
	return err == nil && d != nil && d.handle != nil && api.docHasFieldValue(d.handle, name)
}

func (d *Doc) IsFieldNull(name string) bool {
	api, err := puregoAPI()
	return err == nil && d != nil && d.handle != nil && api.docIsFieldNull(d.handle, name)
}

func (d *Doc) GetFieldNames() ([]string, error) {
	api, err := puregoAPI()
	if err != nil {
		return nil, err
	}
	var cNames unsafe.Pointer
	var count uintptr
	if err := toError(api.docGetFieldNames(d.handle, &cNames, &count)); err != nil {
		return nil, err
	}
	defer api.freeStrArray(cNames, count)
	nameSlice := unsafe.Slice((*unsafe.Pointer)(cNames), int(count))
	names := make([]string, int(count))
	for i := range names {
		names[i] = cStringFromPointer(nameSlice[i])
	}
	return names, nil
}

// HNSWQueryParams represents query parameters for HNSW index.
type HNSWQueryParams struct {
	handle unsafe.Pointer
}

func NewHNSWQueryParams(ef int, radius float32, isLinear, isUsingRefiner bool) *HNSWQueryParams {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.hnswQueryParamsCreate(int32(ef), radius, isLinear, isUsingRefiner)
	if handle == nil {
		return nil
	}
	return &HNSWQueryParams{handle: handle}
}

func (p *HNSWQueryParams) Destroy() {
	if p != nil && p.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.hnswQueryParamsDestroy(p.handle)
		}
		p.handle = nil
	}
}

func (p *HNSWQueryParams) SetEf(ef int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.hnswQueryParamsSetEf(p.handle, int32(ef)))
}

func (p *HNSWQueryParams) GetEf() int {
	api, err := puregoAPI()
	if err != nil || p == nil || p.handle == nil {
		return 0
	}
	return int(api.hnswQueryParamsGetEf(p.handle))
}

// IVFQueryParams represents query parameters for IVF index.
type IVFQueryParams struct {
	handle unsafe.Pointer
}

func NewIVFQueryParams(nprobe int, isUsingRefiner bool, scaleFactor float32) *IVFQueryParams {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.ivfQueryParamsCreate(int32(nprobe), isUsingRefiner, scaleFactor)
	if handle == nil {
		return nil
	}
	return &IVFQueryParams{handle: handle}
}

func (p *IVFQueryParams) Destroy() {
	if p != nil && p.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.ivfQueryParamsDestroy(p.handle)
		}
		p.handle = nil
	}
}

func (p *IVFQueryParams) SetNprobe(nprobe int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.ivfQueryParamsSetNprobe(p.handle, int32(nprobe)))
}

// FlatQueryParams represents query parameters for Flat index.
type FlatQueryParams struct {
	handle unsafe.Pointer
}

func NewFlatQueryParams(isUsingRefiner bool, scaleFactor float32) *FlatQueryParams {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.flatQueryParamsCreate(isUsingRefiner, scaleFactor)
	if handle == nil {
		return nil
	}
	return &FlatQueryParams{handle: handle}
}

func (p *FlatQueryParams) Destroy() {
	if p != nil && p.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.flatQueryParamsDestroy(p.handle)
		}
		p.handle = nil
	}
}

type FTSQueryParams struct {
	handle unsafe.Pointer
}

func NewFTSQueryParams(defaultOperator string) *FTSQueryParams {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	opBuf, cOp := optionalCString(defaultOperator)
	handle := api.ftsQueryParamsCreate(cOp)
	runtime.KeepAlive(opBuf)
	if handle == nil {
		return nil
	}
	return &FTSQueryParams{handle: handle}
}

func (p *FTSQueryParams) Destroy() {
	if p != nil && p.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.ftsQueryParamsDestroy(p.handle)
		}
		p.handle = nil
	}
}

func (p *FTSQueryParams) SetDefaultOperator(op string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	if p == nil || p.handle == nil {
		return invalidArgumentError("FTS query params is nil")
	}
	return toError(api.ftsQueryParamsSetOp(p.handle, op))
}

func (p *FTSQueryParams) GetDefaultOperator() string {
	api, err := puregoAPI()
	if err != nil || p == nil || p.handle == nil {
		return ""
	}
	return api.ftsQueryParamsGetOp(p.handle)
}

// SearchQuery represents a vector query operation.
type SearchQuery struct {
	handle unsafe.Pointer
	fts    *FTS
}

func NewSearchQuery() *SearchQuery {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.vectorQueryCreate()
	if handle == nil {
		return nil
	}
	return &SearchQuery{handle: handle}
}

func (q *SearchQuery) Destroy() {
	if q != nil && q.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.vectorQueryDestroy(q.handle)
		}
		q.handle = nil
	}
}

func (q *SearchQuery) SetFieldName(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.vectorQuerySetFieldName(q.handle, name))
}

func (q *SearchQuery) GetFieldName() string {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return ""
	}
	return api.vectorQueryGetFieldName(q.handle)
}

func (q *SearchQuery) SetTopK(topk int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.vectorQuerySetTopK(q.handle, int32(topk)))
}

func (q *SearchQuery) GetTopK() int {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return 0
	}
	return int(api.vectorQueryGetTopK(q.handle))
}

func (q *SearchQuery) SetQueryVector(data []float32) error {
	if len(data) == 0 {
		return invalidArgumentError("query vector cannot be empty")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.vectorQuerySetQueryVector(q.handle, unsafe.Pointer(&data[0]), uintptr(len(data)*4)))
	runtime.KeepAlive(data)
	return err
}

func (q *SearchQuery) SetFilter(filter string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.vectorQuerySetFilter(q.handle, filter))
}

func (q *SearchQuery) GetFilter() string {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return ""
	}
	return api.vectorQueryGetFilter(q.handle)
}

func (q *SearchQuery) SetIncludeVector(include bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.vectorQuerySetIncludeVector(q.handle, include))
}

func (q *SearchQuery) GetIncludeVector() bool {
	api, err := puregoAPI()
	return err == nil && q != nil && q.handle != nil && api.vectorQueryGetIncludeVector(q.handle)
}

func (q *SearchQuery) SetIncludeDocID(include bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.vectorQuerySetIncludeDocID(q.handle, include))
}

func (q *SearchQuery) GetIncludeDocID() bool {
	api, err := puregoAPI()
	return err == nil && q != nil && q.handle != nil && api.vectorQueryGetIncludeDocID(q.handle)
}

func (q *SearchQuery) SetOutputFields(fields []string) error {
	if len(fields) == 0 {
		return nil
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	ptrs, keep := cStringArray(fields)
	err = toError(api.vectorQuerySetOutputFields(q.handle, unsafe.Pointer(&ptrs[0]), uintptr(len(ptrs))))
	runtime.KeepAlive(ptrs)
	runtime.KeepAlive(keep)
	return err
}

func (q *SearchQuery) SetHNSWParams(params *HNSWQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("HNSW query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.vectorQuerySetHNSWParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SearchQuery) SetIVFParams(params *IVFQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("IVF query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.vectorQuerySetIVFParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SearchQuery) SetFlatParams(params *FlatQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("Flat query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.vectorQuerySetFlatParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SearchQuery) SetFTSParams(params *FTSQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("FTS query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.vectorQuerySetFTSParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SearchQuery) SetFTS(fts *FTS) error {
	if fts == nil || fts.handle == nil {
		return invalidArgumentError("FTS payload is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.vectorQuerySetFTS(q.handle, fts.handle))
}

func (q *SearchQuery) GetFTS() *FTS {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return nil
	}
	handle := api.vectorQueryGetFTS(q.handle)
	if handle == nil {
		return nil
	}
	return &FTS{handle: handle, owned: false}
}

// GroupBySearchQuery is not part of the initial purego POC binding.
type GroupBySearchQuery struct {
	handle unsafe.Pointer
}

func NewGroupBySearchQuery() *GroupBySearchQuery {
	return &GroupBySearchQuery{}
}

func (q *GroupBySearchQuery) Destroy() { q.handle = nil }
func (q *GroupBySearchQuery) SetFieldName(name string) error {
	return unsupportedError("GroupBySearchQuery.SetFieldName")
}
func (q *GroupBySearchQuery) SetGroupByFieldName(name string) error {
	return unsupportedError("GroupBySearchQuery.SetGroupByFieldName")
}
func (q *GroupBySearchQuery) SetGroupCount(count uint32) error {
	return unsupportedError("GroupBySearchQuery.SetGroupCount")
}
func (q *GroupBySearchQuery) SetGroupTopK(topk uint32) error {
	return unsupportedError("GroupBySearchQuery.SetGroupTopK")
}
func (q *GroupBySearchQuery) SetQueryVector(data []float32) error {
	if len(data) == 0 {
		return invalidArgumentError("query vector cannot be empty")
	}
	return unsupportedError("GroupBySearchQuery.SetQueryVector")
}
func (q *GroupBySearchQuery) SetFilter(filter string) error {
	return unsupportedError("GroupBySearchQuery.SetFilter")
}
func (q *GroupBySearchQuery) SetIncludeVector(include bool) error {
	return unsupportedError("GroupBySearchQuery.SetIncludeVector")
}
func (q *GroupBySearchQuery) SetOutputFields(fields []string) error {
	return unsupportedError("GroupBySearchQuery.SetOutputFields")
}
func (q *GroupBySearchQuery) SetHNSWParams(params *HNSWQueryParams) error {
	return unsupportedError("GroupBySearchQuery.SetHNSWParams")
}
func (q *GroupBySearchQuery) SetIVFParams(params *IVFQueryParams) error {
	return unsupportedError("GroupBySearchQuery.SetIVFParams")
}
func (q *GroupBySearchQuery) SetFlatParams(params *FlatQueryParams) error {
	return unsupportedError("GroupBySearchQuery.SetFlatParams")
}

type MultiQuery struct {
	handle unsafe.Pointer
}

func NewMultiQuery() *MultiQuery {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.multiQueryCreate()
	if handle == nil {
		return nil
	}
	return &MultiQuery{handle: handle}
}

func (q *MultiQuery) Destroy() {
	if q != nil && q.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.multiQueryDestroy(q.handle)
		}
		q.handle = nil
	}
}

func (q *MultiQuery) AddSubQuery(sub *SubQuery) error {
	if sub == nil || sub.handle == nil {
		return invalidArgumentError("sub query is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.multiQueryAddSubQuery(q.handle, sub.handle))
}

func (q *MultiQuery) GetSubQueryCount() int {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return 0
	}
	return int(api.multiQueryGetSubQueryCount(q.handle))
}

func (q *MultiQuery) SetTopK(topk int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.multiQuerySetTopK(q.handle, int32(topk)))
}

func (q *MultiQuery) GetTopK() int {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return 0
	}
	return int(api.multiQueryGetTopK(q.handle))
}

func (q *MultiQuery) SetFilter(filter string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.multiQuerySetFilter(q.handle, filter))
}

func (q *MultiQuery) GetFilter() string {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return ""
	}
	return api.multiQueryGetFilter(q.handle)
}

func (q *MultiQuery) SetIncludeVector(include bool) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.multiQuerySetIncludeVector(q.handle, include))
}

func (q *MultiQuery) GetIncludeVector() bool {
	api, err := puregoAPI()
	return err == nil && q != nil && q.handle != nil && api.multiQueryGetIncludeVector(q.handle)
}

func (q *MultiQuery) SetOutputFields(fields []string) error {
	if len(fields) == 0 {
		return nil
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	ptrs, keep := cStringArray(fields)
	err = toError(api.multiQuerySetOutputFields(q.handle, unsafe.Pointer(&ptrs[0]), uintptr(len(ptrs))))
	runtime.KeepAlive(ptrs)
	runtime.KeepAlive(keep)
	return err
}

func (q *MultiQuery) SetRerankRRF(rankConstant int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.multiQuerySetRerankRRF(q.handle, int32(rankConstant)))
}

func (q *MultiQuery) SetRerankWeighted(weights []float64) error {
	if len(weights) == 0 {
		return invalidArgumentError("weights cannot be empty")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.multiQuerySetRerankWeighted(q.handle, unsafe.Pointer(&weights[0]), uintptr(len(weights))))
	runtime.KeepAlive(weights)
	return err
}

type SubQuery struct {
	handle unsafe.Pointer
}

func NewSubQuery() *SubQuery {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.subQueryCreate()
	if handle == nil {
		return nil
	}
	return &SubQuery{handle: handle}
}

func (q *SubQuery) Destroy() {
	if q != nil && q.handle != nil {
		if api, err := puregoAPI(); err == nil {
			api.subQueryDestroy(q.handle)
		}
		q.handle = nil
	}
}

func (q *SubQuery) SetNumCandidates(n int) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.subQuerySetNumCandidates(q.handle, int32(n)))
}

func (q *SubQuery) GetNumCandidates() int {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return 0
	}
	return int(api.subQueryGetNumCandidates(q.handle))
}

func (q *SubQuery) SetFieldName(name string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.subQuerySetFieldName(q.handle, name))
}

func (q *SubQuery) GetFieldName() string {
	api, err := puregoAPI()
	if err != nil || q == nil || q.handle == nil {
		return ""
	}
	return api.subQueryGetFieldName(q.handle)
}

func (q *SubQuery) SetQueryVector(data []float32) error {
	if len(data) == 0 {
		return invalidArgumentError("query vector cannot be empty")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.subQuerySetQueryVector(q.handle, unsafe.Pointer(&data[0]), uintptr(len(data)*4)))
	runtime.KeepAlive(data)
	return err
}

func (q *SubQuery) SetSparseVector(indices []uint32, values []float32) error {
	if len(indices) != len(values) {
		return invalidArgumentError("indices and values must have the same length")
	}
	if len(indices) == 0 {
		return invalidArgumentError("sparse vector cannot be empty")
	}
	return unsupportedError("SubQuery.SetSparseVector")
}

func (q *SubQuery) SetHNSWParams(params *HNSWQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("HNSW query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.subQuerySetHNSWParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SubQuery) SetIVFParams(params *IVFQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("IVF query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.subQuerySetIVFParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SubQuery) SetFlatParams(params *FlatQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("Flat query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.subQuerySetFlatParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SubQuery) SetFTSParams(params *FTSQueryParams) error {
	if params == nil || params.handle == nil {
		return invalidArgumentError("FTS query params is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	err = toError(api.subQuerySetFTSParams(q.handle, params.handle))
	if err == nil {
		params.handle = nil
	}
	return err
}

func (q *SubQuery) SetFTS(fts *FTS) error {
	if fts == nil || fts.handle == nil {
		return invalidArgumentError("FTS payload is nil")
	}
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	return toError(api.subQuerySetFTS(q.handle, fts.handle))
}

type FTS struct {
	handle unsafe.Pointer
	owned  bool
}

func NewFTS() *FTS {
	api, err := puregoAPI()
	if err != nil {
		return nil
	}
	handle := api.ftsCreate()
	if handle == nil {
		return nil
	}
	return &FTS{handle: handle, owned: true}
}

func (f *FTS) Destroy() {
	if f != nil && f.handle != nil && f.owned {
		if api, err := puregoAPI(); err == nil {
			api.ftsDestroy(f.handle)
		}
		f.handle = nil
	}
}

func (f *FTS) SetQueryString(query string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	if f == nil || f.handle == nil {
		return invalidArgumentError("FTS payload is nil")
	}
	return toError(api.ftsSetQueryString(f.handle, query))
}

func (f *FTS) GetQueryString() string {
	api, err := puregoAPI()
	if err != nil || f == nil || f.handle == nil {
		return ""
	}
	return api.ftsGetQueryString(f.handle)
}

func (f *FTS) SetMatchString(match string) error {
	api, err := puregoAPI()
	if err != nil {
		return err
	}
	if f == nil || f.handle == nil {
		return invalidArgumentError("FTS payload is nil")
	}
	return toError(api.ftsSetMatchString(f.handle, match))
}

func (f *FTS) GetMatchString() string {
	api, err := puregoAPI()
	if err != nil || f == nil || f.handle == nil {
		return ""
	}
	return api.ftsGetMatchString(f.handle)
}

type zvecString struct {
	data     unsafe.Pointer
	length   uintptr
	capacity uintptr
}

type zvecStringArray struct {
	strings unsafe.Pointer
	count   uintptr
}

func optionalCString(value string) ([]byte, unsafe.Pointer) {
	if value == "" {
		return nil, nil
	}
	buf := nullTerminatedBytes(value)
	return buf, unsafe.Pointer(&buf[0])
}

func readZvecStringArray(ptr unsafe.Pointer) []string {
	array := (*zvecStringArray)(ptr)
	if array == nil || array.strings == nil || array.count == 0 {
		return nil
	}
	out := make([]string, int(array.count))
	elemSize := unsafe.Sizeof(zvecString{})
	for i := uintptr(0); i < array.count; i++ {
		item := (*zvecString)(unsafe.Pointer(uintptr(array.strings) + i*elemSize))
		if item.data != nil && item.length > 0 {
			out[i] = string(unsafe.Slice((*byte)(item.data), int(item.length)))
		}
	}
	return out
}

func nullTerminatedBytes(value string) []byte {
	buf := make([]byte, len(value)+1)
	copy(buf, value)
	return buf
}

func cStringArray(values []string) ([]unsafe.Pointer, [][]byte) {
	ptrs := make([]unsafe.Pointer, len(values))
	keep := make([][]byte, len(values))
	for i, value := range values {
		keep[i] = nullTerminatedBytes(value)
		ptrs[i] = unsafe.Pointer(&keep[i][0])
	}
	return ptrs, keep
}
