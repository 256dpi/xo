package xo

import (
	"errors"
	"fmt"
	"io"
)

// Err is the error returned by F(), W(), WS() and WF().
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
		Caller: GetCaller(1, 0),
	}
}

// W will wrap an error. The error is not wrapped if the parent error already
// captured the caller.
func W(err error) error {
	return WS(err, 1)
}

// WS will wrap an error. The error is not wrapped if the parent error already
// captured the caller. If the error is wrapped the specified amount of frames
// is skipped.
func WS(err error, skip int) error {
	// check nil
	if err == nil {
		return nil
	}

	// get caller
	caller := GetCaller(1+skip, 0)

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
		Caller: GetCaller(1, 0),
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
//	%s   message
//	%q   "message"
//	%v   caller: message
//	%+v  err
//	     message
//	     caller
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

// Drop will drop the specified amount of frames from the error if possible.
func Drop(err error, n int) error {
	// check error
	if anErr, ok := err.(*Err); ok {
		anErr.Caller.Drop(n)
	}

	return err
}

// SafeErr wraps Err to indicate presentation safety.
type SafeErr struct {
	Err
}

// SF is a short-hand function to format a safe error.
func SF(format string, args ...interface{}) error {
	return &SafeErr{
		Err: Err{
			Msg:    fmt.Sprintf(format, args...),
			Caller: GetCaller(1, 0),
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
			Caller: GetCaller(1, 0),
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

// BaseErr represents the base of an error chain. It cannot directly be used as
// an error value. Instead, the caller uses Wrap() to get a new wrapped error or
// Self() to get the identity error.
type BaseErr struct {
	err error
}

// BF formats and returns a new base error.
func BF(format string, args ...interface{}) BaseErr {
	return BaseErr{
		err: &Err{
			Msg:    fmt.Sprintf(format, args...),
			Caller: GetCaller(1, 1),
		},
	}
}

// BW wraps and returns a new base error.
func BW(err error) BaseErr {
	return BaseErr{
		err: &Err{
			Err:    err,
			Caller: GetCaller(1, 1),
		},
	}
}

// Self will return the identity error.
func (b *BaseErr) Self() error {
	return b.err
}

// Wrap wraps and returns an error value.
func (b *BaseErr) Wrap() error {
	return &Err{
		Err:    b.err,
		Caller: GetCaller(1, 0),
	}
}

// WrapF wraps, formats and returns an error value.
func (b *BaseErr) WrapF(format string, args ...interface{}) error {
	return &Err{
		Err:    b.err,
		Msg:    fmt.Sprintf(format, args...),
		Caller: GetCaller(1, 0),
	}
}

// Is returns whether the provided error is a descendant.
func (b *BaseErr) Is(err error) bool {
	return errors.Is(err, b.err)
}
