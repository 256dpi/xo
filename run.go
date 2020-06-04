package xo

import (
	"context"
)

// Context is used by Run to track execution.
type Context struct {
	context.Context

	// The caller.
	Caller Caller

	// The span.
	Span Span
}

// Rename will set a new name on the span.
func (c *Context) Rename(name string) {
	c.Span.Rename(name)
}

// Attach will add the provided event to the span.
func (c *Context) Attach(event string, attributes M) {
	c.Span.Attach(event, attributes)
}

// Log will attach a log event to the span.
func (c *Context) Log(format string, args ...interface{}) {
	c.Span.Log(format, args...)
}

// Tag will add the provided attribute to the span.
func (c *Context) Tag(key string, value interface{}) {
	c.Span.Tag(key, value)
}

// Run will run the provided function and automatically handle tracing, error
// handling and panic recovering.
func Run(ctx context.Context, fn func(ctx *Context) error) (err error) {
	// ensure context
	if ctx == nil {
		ctx = context.Background()
	}

	// get caller
	caller := GetCaller(1)

	// trace
	ctx, span := Trace(ctx, caller.Short)
	defer span.End()

	// wrap
	xtc := &Context{
		Caller:  caller,
		Context: ctx,
		Span:    span,
	}

	// defer cleanup
	defer func() {
		// recover panic
		val := recover()
		if val != nil {
			switch val := val.(type) {
			case error:
				err = WF(val, "PANIC")
			case string:
				err = F("PANIC: %s", val)
			default:
				err = F("PANIC: %v", val)
			}
		}

		// check error
		if err == nil {
			return
		}

		// wrap error
		err = &Err{
			Err:    err,
			Caller: xtc.Caller,
		}

		// record error
		span.Record(err)
	}()

	// yield
	err = fn(xtc)

	return
}
