package zvec

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      ErrorCode
		expected int
	}{
		{"OK", OK, 0},
		{"NotFound", NotFound, 1},
		{"AlreadyExists", AlreadyExists, 2},
		{"InvalidArgument", InvalidArgument, 3},
		{"PermissionDenied", PermissionDenied, 4},
		{"FailedPrecondition", FailedPrecondition, 5},
		{"ResourceExhausted", ResourceExhausted, 6},
		{"Unavailable", Unavailable, 7},
		{"InternalError", InternalError, 8},
		{"NotSupported", NotSupported, 9},
		{"Unknown", Unknown, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.got) != tt.expected {
				t.Errorf("ErrorCode %s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestErrorCodeString(t *testing.T) {
	tests := []struct {
		code ErrorCode
	}{
		{OK},
		{NotFound},
		{AlreadyExists},
		{InvalidArgument},
		{InternalError},
		{Unknown},
	}
	for _, tt := range tests {
		s := tt.code.String()
		if s == "" {
			t.Errorf("ErrorCode(%d).String() returned empty string", tt.code)
		}
	}
}

func TestErrorMessage(t *testing.T) {
	err := &Error{Code: NotFound, Message: "item not found"}
	got := err.Error()
	if got == "" {
		t.Error("Error.Error() should not be empty")
	}
	expected := fmt.Sprintf("zvec error [%s]: item not found", NotFound)
	if got != expected {
		t.Errorf("Error.Error() = %q, want %q", got, expected)
	}
}

func TestErrorMessageFormatting(t *testing.T) {
	tests := []struct {
		code    ErrorCode
		message string
	}{
		{OK, "success"},
		{InternalError, "something broke"},
		{Unknown, ""},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			err := &Error{Code: tt.code, Message: tt.message}
			want := fmt.Sprintf("zvec error [%s]: %s", tt.code, tt.message)
			if err.Error() != want {
				t.Errorf("Error.Error() = %q, want %q", err.Error(), want)
			}
		})
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	err := &Error{Code: NotFound, Message: "test"}

	var asError error = err
	if asError.Error() == "" {
		t.Error("Error.Error() should not be empty")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"not found error", &Error{Code: NotFound, Message: "not found"}, true},
		{"other zvec error", &Error{Code: InternalError, Message: "internal"}, false},
		{"standard error", errors.New("some error"), false},
		{"wrapped not found", fmt.Errorf("wrapped: %w", &Error{Code: NotFound, Message: "not found"}), true},
		{"wrapped other", fmt.Errorf("wrapped: %w", &Error{Code: AlreadyExists, Message: "exists"}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"already exists error", &Error{Code: AlreadyExists, Message: "exists"}, true},
		{"not found error", &Error{Code: NotFound, Message: "not found"}, false},
		{"standard error", errors.New("some error"), false},
		{"wrapped already exists", fmt.Errorf("wrapped: %w", &Error{Code: AlreadyExists, Message: "exists"}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAlreadyExists(tt.err); got != tt.want {
				t.Errorf("IsAlreadyExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInvalidArgument(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"invalid argument error", &Error{Code: InvalidArgument, Message: "bad arg"}, true},
		{"not found error", &Error{Code: NotFound, Message: "not found"}, false},
		{"standard error", errors.New("some error"), false},
		{"wrapped invalid argument", fmt.Errorf("wrapped: %w", &Error{Code: InvalidArgument, Message: "bad"}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidArgument(tt.err); got != tt.want {
				t.Errorf("IsInvalidArgument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorsAs(t *testing.T) {
	originalErr := &Error{Code: NotFound, Message: "resource not found"}
	wrappedErr := fmt.Errorf("operation failed: %w", originalErr)

	var zvecErr *Error
	if !errors.As(wrappedErr, &zvecErr) {
		t.Fatal("errors.As should find *Error in wrapped error")
	}
	if zvecErr.Code != NotFound {
		t.Errorf("extracted error code = %d, want %d", zvecErr.Code, NotFound)
	}
	if zvecErr.Message != "resource not found" {
		t.Errorf("extracted error message = %q, want %q", zvecErr.Message, "resource not found")
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []*Error{
		ErrNotFound,
		ErrAlreadyExists,
		ErrInvalidArgument,
		ErrPermissionDenied,
		ErrFailedPrecondition,
		ErrResourceExhausted,
		ErrUnavailable,
		ErrInternalError,
		ErrNotSupported,
		ErrUnknown,
	}
	expectedCodes := []ErrorCode{
		NotFound, AlreadyExists, InvalidArgument, PermissionDenied,
		FailedPrecondition, ResourceExhausted, Unavailable,
		InternalError, NotSupported, Unknown,
	}
	for i, sentinel := range sentinels {
		if sentinel.Code != expectedCodes[i] {
			t.Errorf("sentinel[%d].Code = %d, want %d", i, sentinel.Code, expectedCodes[i])
		}
		if sentinel.Message == "" {
			t.Errorf("sentinel[%d].Message should not be empty", i)
		}
	}
}

func BenchmarkErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &Error{Code: NotFound, Message: "not found"}
	}
}

func BenchmarkIsNotFound(b *testing.B) {
	err := &Error{Code: NotFound, Message: "not found"}
	for i := 0; i < b.N; i++ {
		IsNotFound(err)
	}
}

func BenchmarkErrorMessage(b *testing.B) {
	err := &Error{Code: NotFound, Message: "not found"}
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
