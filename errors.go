package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// ErrorCode represents a zvec error code.
type ErrorCode int

const (
	OK                ErrorCode = 0
	NotFound          ErrorCode = 1
	AlreadyExists     ErrorCode = 2
	InvalidArgument   ErrorCode = 3
	PermissionDenied  ErrorCode = 4
	FailedPrecondition ErrorCode = 5
	ResourceExhausted ErrorCode = 6
	Unavailable       ErrorCode = 7
	InternalError     ErrorCode = 8
	NotSupported      ErrorCode = 9
	Unknown           ErrorCode = 10
)

// String returns the string representation of the error code.
func (c ErrorCode) String() string {
	cStr := C.zvec_error_code_to_string(C.zvec_error_code_t(c))
	return C.GoString(cStr)
}

// Error represents a zvec error with code and message.
type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("zvec error [%s]: %s", e.Code, e.Message)
}

// Sentinel errors for common error codes.
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

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	var zvecErr *Error
	return errors.As(err, &zvecErr) && zvecErr.Code == NotFound
}

// IsAlreadyExists checks if the error is an already exists error.
func IsAlreadyExists(err error) bool {
	var zvecErr *Error
	return errors.As(err, &zvecErr) && zvecErr.Code == AlreadyExists
}

// IsInvalidArgument checks if the error is an invalid argument error.
func IsInvalidArgument(err error) bool {
	var zvecErr *Error
	return errors.As(err, &zvecErr) && zvecErr.Code == InvalidArgument
}

// toError converts a C error code to a Go error.
// Returns nil if the error code is ZVEC_OK.
func toError(code C.zvec_error_code_t) error {
	if code == C.ZVEC_OK {
		return nil
	}

	var cMsg *C.char
	C.zvec_get_last_error(&cMsg)
	defer func() {
		if cMsg != nil {
			C.zvec_free(unsafe.Pointer(cMsg))
		}
	}()

	message := "unknown error"
	if cMsg != nil {
		message = C.GoString(cMsg)
	}

	return &Error{
		Code:    ErrorCode(code),
		Message: message,
	}
}
