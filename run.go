package xo

import (
	"context"
)

// Context is used by Run to track execution.
type Context struct {
	context.Context

	// The caller.
	Caller CallerInfo

	// The span created to track execution.
	Span *Span
}

// Log will log the provided key and value.
func (c *Context) Log(key string, value interface{}) {
	c.Span.Log(key, value)
}

// Tag will tag the provided key and value.
func (c *Context) Tag(key string, value interface{}) {
	c.Span.Tag(key, value)
}

// Run will run the provided function and automatically handle tracking, error
// handling and panic recovering.
func Run(ctx context.Context, fn func(ctx *Context) error) (err error) {
	// ensure context
	if ctx == nil {
		ctx = context.Background()
	}

	// get caller
	caller := Caller(1)

	// track
	ctx, span := Track(ctx, caller.Short)
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
		err = WF(err, xtc.Caller.Short)

		// record error
		span.Record(err)

		// TODO: Add tags and logs to error?
	}()

	// yield
	err = fn(xtc)

	return
}
