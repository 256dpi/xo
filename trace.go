package xo

import (
	"context"
)

type traceContextKey struct{}

var traceKey = traceContextKey{}

type traceContext struct {
	context.Context
	trace *Trace
}

func (c *traceContext) Value(key interface{}) interface{} {
	// check key
	if key == traceKey {
		return c.trace
	}

	// get value
	val := c.Context.Value(key)

	// return tail if root is returned
	if val == c.trace.root.span {
		return c.trace.Tail().span
	}

	return val
}

// Trace implements a span stack that can be used with fat contexts. Rather than
// branching of the context for every function call, a span is pushed onto the
// stack to track execution.
//
// Code that uses Track will automatically discover the stack and branch of its
// tail if no previous branch has been detected. Other opentracing compatible
// code would not be able to find the stack and would use the root span instead.
// This can be prevented by ensuring a branch using the Branch function.
type Trace struct {
	root  Span
	stack []Span
}

// CreateTrace returns a new trace that will use the span found in the provided
// context as its root or start a new one. The returned context is the provided
// context wrapped with the new span and trace.
func CreateTrace(ctx context.Context, name string) (*Trace, context.Context) {
	// check context
	if ctx == nil {
		ctx = context.Background()
	}

	// get span
	ctx, span := Track(ctx, name)

	// create trace
	trace := &Trace{
		root:  span,
		stack: make([]Span, 0, 32),
	}

	// add trace
	ctx = &traceContext{
		Context: ctx,
		trace:   trace,
	}

	return trace, ctx
}

// GetTrace will return the trace from the context or nil if absent.
func GetTrace(ctx context.Context) *Trace {
	// check context
	if ctx == nil {
		return nil
	}

	// get trace
	trace, _ := ctx.Value(traceKey).(*Trace)

	return trace
}

// Push will add a new span to the trace.
func (t *Trace) Push(name string) {
	// get parent
	parent := t.Tail()

	// create child
	_, child := Track(parent.ctx, name)

	// push child
	t.stack = append(t.stack, child)
}

// Tag will add the provided attribute to the last pushed span.
func (t *Trace) Tag(key string, value interface{}) {
	t.Tail().Tag(key, value)
}

// Attach will add the provided event to the last pushed span.
func (t *Trace) Attach(event string, attributes M) {
	t.Tail().Attach(event, attributes)
}

// Log will attach a log event to the last pushed span.
func (t *Trace) Log(format string, args ...interface{}) {
	t.Tail().Log(format, args...)
}

// Record will attach an error event to the last pushed span.
func (t *Trace) Record(err error) {
	t.Tail().Record(err)
}

// Pop ends and removes the last pushed span. This call is usually deferred
// right after a push.
func (t *Trace) Pop() {
	// check list
	if len(t.stack) == 0 {
		return
	}

	// finish last span
	t.stack[len(t.stack)-1].End()

	// resize stack
	t.stack = t.stack[:len(t.stack)-1]
}

// End will end all stacked spans and the root span.
func (t *Trace) End() {
	// end stacked spans
	for _, span := range t.stack {
		span.End()
	}

	// end root span
	t.root.End()
}

// Tail returns the tail or root of the span stack.
func (t *Trace) Tail() Span {
	// return last span if available
	if len(t.stack) > 0 {
		return t.stack[len(t.stack)-1]
	}

	return t.root
}

// Root will return the root of the span stack.
func (t *Trace) Root() Span {
	return t.root
}
