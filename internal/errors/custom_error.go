package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrorType uint

const (
	NoType = ErrorType(iota)
)

type customError struct {
	errorType     ErrorType
	originalError error
	contextInfo   map[string]string
}

func (e customError) Error() string {
	return e.originalError.Error()
}

func (t ErrorType) New(msg string) error {
	return customError{
		errorType:     t,
		originalError: errors.New(msg),
	}
}

func (t ErrorType) Newf(msg string, args ...any) error {
	return customError{
		errorType:     t,
		originalError: fmt.Errorf(msg, args...),
	}
}

func (t ErrorType) Wrap(err error, msg string) error {
	return t.Wrapf(err, msg)
}

func (t ErrorType) Wrapf(err error, msg string, args ...any) error {
	return customError{
		errorType:     t,
		originalError: errors.Wrapf(err, msg, args...),
	}
}

func New(msg string) error {
	return customError{errorType: NoType, originalError: errors.New(msg)}
}

func Newf(msg string, args ...any) error {
	return customError{
		errorType:     NoType,
		originalError: fmt.Errorf(msg, args...),
	}
}

func Cause(err error) error {
	return errors.Cause(err)
}

func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

func Wrapf(err error, msg string, args ...any) error {
	wrappedError := errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(customError); ok {
		return customError{
			errorType:     customErr.errorType,
			originalError: wrappedError,
			contextInfo:   customErr.contextInfo,
		}
	}
	return customError{errorType: NoType, originalError: wrappedError}
}

func AddErrorContext(err error, field, message string) error {
	context := map[string]string{"field": field, "message": message}
	if customErr, ok := err.(customError); ok {
		return customError{
			errorType:     customErr.errorType,
			originalError: customErr.originalError,
			contextInfo:   context,
		}
	}
	return customError{errorType: NoType, originalError: err, contextInfo: context}
}

func ErrorContext(err error) map[string]string {
	if customErr, ok := err.(customError); ok || customErr.contextInfo != nil {
		return customErr.contextInfo
	}
	return nil
}

func Type(err error) ErrorType {
	if customErr, ok := err.(customError); ok {
		return customErr.errorType
	}

	return NoType
}
