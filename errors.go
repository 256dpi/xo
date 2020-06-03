package xo

import (
	"errors"
	"fmt"
	"io"
)

// Err is the error returned by F(), W() and WF().
type Err struct {
	Err    error
	Msg    string
	Caller Caller
}

// F will format an error. This function can be used instead of errors.New() and
// fmt.Errorf().
func F(format string, args ...interface{}) error {
	return &Err{
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1),
	}
}

// W will wrap an error. The error is not wrapped if the parent error already
// captured the caller.
func W(err error) error {
	// check nil
	if err == nil {
		return nil
	}

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

// WF will wrap and error with a formatted message. This function can be used
// instead of wrapping with fmt.Errorf().
func WF(err error, format string, args ...interface{}) error {
	// check nil
	if err == nil {
		return nil
	}

	return &Err{
		Err:    err,
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1),
	}
}

// Error will return the error string.
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
//  %q   "message"
//  %v   caller: message
//  %+v  err
//       message
//       caller
//
func (e *Err) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.Err != nil {
				check(fmt.Fprintf(s, "%+v\n", e.Err))
			}
			if e.Msg != "" {
				check(io.WriteString(s, e.Msg))
				check(io.WriteString(s, "\n"))
			}
			e.Caller.Format(s, verb)
		} else {
			check(fmt.Fprintf(s, "%s: %s", e.Caller.Short, e.Error()))
		}
	case 's':
		check(io.WriteString(s, e.Error()))
	case 'q':
		check(fmt.Fprintf(s, "%q", e.Error()))
	}
}

// StackTrace will return the stack trace (for sentry compatibility).
func (e *Err) StackTrace() []uintptr {
	return e.Caller.Stack
}

// SafeErr wraps an Err to indicate presentation safety.
type SafeErr struct {
	Err
}

// SF is a short-hand function to format a safe error.
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
	// check nil
	if err == nil {
		return nil
	}

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
