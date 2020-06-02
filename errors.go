package xo

import (
	"fmt"

	"github.com/pkg/errors"
)

// Err is the error returned by F(), W() and WF().
type Err struct {
	Err    error
	Msg    string
	Caller Caller
}

// F will format and error similar to errors.New() and fmt.Errorf().
func F(format string, args ...interface{}) error {
	return &Err{
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1),
	}
}

// W will wrap an error.
func W(err error) error {
	// get caller
	caller := GetCaller(1)

	// check if wrapping is superfluous
	if anErr, ok := err.(*Err); ok {
		if anErr.Caller.Includes(caller, 1) {
			return anErr
		}
	}

	return &Err{
		Err:    err,
		Caller: caller,
	}
}

// WF will wrap and error with a formatted message.
func WF(err error, format string, args ...interface{}) error {
	return &Err{
		Err:    err,
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1),
	}
}

// Error will return the error as a string.
func (e *Err) Error() string {
	if e.Msg != "" && e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	} else if e.Err != nil {
		return e.Err.Error()
	}

	return e.Msg
}

// Unwrap will return the wrapped error, if any.
func (e *Err) Unwrap() error {
	return e.Err
}

// Format will format the error.
//
//  %s   message
//  %q   quoted message
//  %v   caller + message
//  %+v  parent + message + stack
//
func (e *Err) Format(s fmt.State, verb rune) {
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

// StackTrace will return the stack trace (for sentry).
func (e *Err) StackTrace() []uintptr {
	return e.Caller.Stack
}

// SafeErr wraps an Err to indicate presentation safety.
type SafeErr struct {
	Err
}

// SF is a short-hand function to construct a safe error.
func SF(format string, args ...interface{}) error {
	return &SafeErr{
		Err: Err{
			Msg:    fmt.Sprintf(format, args...),
			Caller: GetCaller(1),
		},
	}
}

// SW wraps an error and marks it as safe. Wrapped errors are safe to be
// presented to the client if appropriate.
func SW(err error) error {
	return &SafeErr{
		Err: Err{
			Err:    err,
			Caller: GetCaller(1),
		},
	}
}

// IsSafe can be used to check if an error has been wrapped using SW. It will
// also detect further wrapped safe errors.
func IsSafe(err error) bool {
	return AsSafe(err) != nil
}

// AsSafe will return the safe error from an error chain.
func AsSafe(err error) *SafeErr {
	var safeErr *SafeErr
	errors.As(err, &safeErr)
	return safeErr
}
