package xo

import (
	"fmt"

	"github.com/pkg/errors"
)

// TODO: Add own implementations?

type Error struct {
	Err    error
	Msg    string
	Caller Caller
}

// F will format and error similar to errors.New() and fmt.Errorf().
func F(format string, args ...interface{}) error {
	return &Error{
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1),
	}
}

// W will wrap an error.
func W(err error) error {
	return &Error{
		Err:    err,
		Caller: GetCaller(1),
	}
}

// WF will wrap and error with a formatted message.
func WF(err error, format string, args ...interface{}) error {
	return &Error{
		Err:    err,
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1),
	}
}

// Error will return the error as a string.
func (e *Error) Error() string {
	if e.Msg != "" && e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	} else if e.Err != nil {
		return e.Err.Error()
	}

	return e.Msg
}

// Unwrap will return the wrapped error, if any.
func (e *Error) Unwrap() error {
	return e.Err
}

// Format will format the error.
//
//  %s   message
//  %q   quoted message
//  %v   caller + message
//  %+v  parent + message + stack
//
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.Err != nil {
				justFprintf(s, "%+v\n", e.Err)
			}
			if e.Msg != "" {
				justPrint(s, e.Msg)
				justPrint(s, "\n")
			}
			e.Caller.Format(s, verb)
		} else {
			justFprintf(s, "%s: %s", e.Caller.Short, e.Error())
		}
	case 's':
		justPrint(s, e.Error())
	case 'q':
		justFprintf(s, "%q", e.Error())
	}
}

func (e *Error) StackTrace() []uintptr {
	return e.Caller.Stack
}

// Safe error?

// SafeError wraps an error to indicate presentation safety.
type SafeError struct {
	Err error
}

// E is a short-hand function to construct a safe error.
func E(format string, args ...interface{}) error {
	return Safe(F(format, args...))
}

// Safe wraps an error and marks it as safe. Wrapped errors are safe to be
// presented to the client if appropriate.
func Safe(err error) error {
	return &SafeError{
		Err: err,
	}
}

// Error implements the error interface.
func (err *SafeError) Error() string {
	return err.Err.Error()
}

// Unwrap will return the wrapped error.
func (err *SafeError) Unwrap() error {
	return err.Err
}

// Format implements the fmt.Formatter interface.
func (err *SafeError) Format(s fmt.State, verb rune) {
	// check if err implements formatter
	if fErr, ok := err.Err.(fmt.Formatter); ok {
		fErr.Format(s, verb)
		return
	}

	// write string
	justPrint(s, err.Error())
}

// IsSafe can be used to check if an error has been wrapped using Safe. It will
// also detect further wrapped safe errors.
func IsSafe(err error) bool {
	return AsSafe(err) != nil
}

// AsSafe will return the safe error from an error chain.
func AsSafe(err error) *SafeError {
	var safeErr *SafeError
	errors.As(err, &safeErr)
	return safeErr
}
