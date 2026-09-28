//go:build cgo && !purego

package zvec

/*
#include "zvec/c_api.h"
#include <stdlib.h>

// zvec_go_status_t bundles an operation's error code with the thread-local
// last-error message fetched in the SAME native call as the operation (see
// the zvec_go_*_ex wrappers in doc.go/collection.go). A goroutine may
// migrate OS threads between two separate cgo calls, and zvec keeps the
// last error in thread-local storage — so reading the message in a later
// cgo call can observe another call's (or another thread's) message under
// concurrent failure load. Returning the pair by value from the wrapper
// pins the read to the thread that saw the failure and costs the happy
// path nothing.
typedef struct zvec_go_status {
  zvec_error_code_t code;
  char *err_msg; // malloc'd copy when non-NULL; free with zvec_free (any thread)
} zvec_go_status_t;
*/
import "C"
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

// ErrClosed is returned when an operation is attempted on a Collection whose
// Close or Destroy has already completed (or is completing concurrently).
var ErrClosed = &Error{Code: FailedPrecondition, Message: "collection is closed"}

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

// lockErrorThread keeps a native call and its subsequent thread-local error
// lookup on the same OS thread. It remains for call sites not yet migrated
// to the zvec_go_*_ex wrappers, which fold the error lookup into the same
// native call and need no thread pinning.
func lockErrorThread() func() {
	runtime.LockOSThread()
	return runtime.UnlockOSThread
}

func invalidArgumentError(message string) error {
	return &Error{Code: InvalidArgument, Message: message}
}

// statusError converts a zvec_go_status_t (error code + last-error message
// captured in the same cgo transition by the zvec_go_*_ex wrappers) into a
// Go error and frees the message buffer. Freeing a malloc'd copy is
// thread-agnostic, so the freeing transition does not need to run on the
// thread that saw the failure. A nil message falls back to a generic text,
// mirroring toError.
func statusError(status C.zvec_go_status_t) error {
	if status.code == C.ZVEC_OK {
		return nil
	}

	defer func() {
		if status.err_msg != nil {
			C.zvec_free(unsafe.Pointer(status.err_msg))
		}
	}()

	message := "unknown error"
	if status.err_msg != nil {
		message = C.GoString(status.err_msg)
	}

	return &Error{
		Code:    ErrorCode(status.code),
		Message: message,
	}
}
