package errorx

import (
	"errors"
	"fmt"
)

// 参考：https://github.com/go-kratos/kratos/tree/main/errors

const (
	// UnknownCode is unknown code for error info.
	UnknownCode = 500
	// UnknownReason is unknown reason for error info.
	UnknownReason = "服务器繁忙，请稍后再试！"
	// SupportPackageIsVersion1 this constant should not be referenced by any other code.
	SupportPackageIsVersion1 = true
)

type Error struct {
	Code     int
	Message  string
	Reason   string
	Metadata map[string]string
	cause    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v", e.Code, e.Reason, e.Message, e.Metadata, e.cause)
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *Error) Unwrap() error { return e.cause }

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}

	return false
}

// WithCause with the underlying cause of the error.
func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause

	return err
}

// WithMetadata with an MD formed by the mapping of key, value.
func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)
	err.Metadata = md

	return err
}

// New returns an error object for the code, message.
func New(code int, reason, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Reason:  reason,
	}
}

// Newf New(code fmt.Sprintf(format, a...))
func Newf(code int, reason, format string, a ...interface{}) *Error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

// Errorf returns an error object for the code, message and error info.
func Errorf(code int, reason, format string, a ...interface{}) error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

// Code returns the http code for an error.
// It supports wrapped errorx.
func Code(err error) int {
	if err == nil {
		return 200 //nolint:gomnd
	}

	return FromError(err).Code
}

// Message returns the message for an error.
// It supports wrapped errorx.
func Message(err error) string {
	if err == nil {
		return UnknownReason //nolint:gomnd
	}

	return FromError(err).Message
}

// Reason returns the reason for a particular error.
// It supports wrapped errorx.
func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}
	return FromError(err).Reason
}

// Clone deep clone error to a new error.
func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}

	metadata := make(map[string]string, len(err.Metadata))

	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		cause:    err.cause,
		Code:     err.Code,
		Reason:   err.Reason,
		Message:  err.Message,
		Metadata: metadata,
	}
}

// FromError try to convert an error to *Error.
// It supports wrapped errorx.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	if se := new(Error); errors.As(err, &se) {
		return se
	}
	return New(UnknownCode, UnknownReason, err.Error())
}
