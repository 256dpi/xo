package xo

import (
	"context"
)

type tracerContextKey struct{}

var tracerKey = tracerContextKey{}

type tracerContext struct {
	context.Context
	tracer *Tracer
}

func (c *tracerContext) Value(key interface{}) interface{} {
	// check key
	if key == tracerKey {
		return c.tracer
	}

	// get value
	val := c.Context.Value(key)

	// return tail if root is returned
	if val == c.tracer.root.span {
		return c.tracer.Tail().span
	}

	return val
}

// Tracer manages a span stack that can be used with fat contexts. Rather than
// branching of the context for every function call, a span is pushed onto the
// tracers stack to trace execution.
//
// Code that uses Trace or native opentelemetry APIs will automatically discover
// the stack and branch of its tail if no previous branch has been detected.
type Tracer struct {
	root  Span
	stack []Span
}

// NewTracer returns a new tracer that will use the provided span as its root.
// The returned context is the provided context wrapped with the tracer. The
// provide context should already contain the provided spans native span.
func NewTracer(ctx context.Context, span Span) (*Tracer, context.Context) {
	// check context
	if ctx == nil {
		ctx = context.Background()
	}

	// create tracer
	tracer := &Tracer{
		root:  span,
		stack: make([]Span, 0, 32),
	}

	// add tracer
	ctx = &tracerContext{
		Context: ctx,
		tracer:  tracer,
	}

	return tracer, ctx
}

// CreateTracer returns a new tracer that will use the span found in the provided
// context as its root or start a new one. The returned context is the provided
// context wrapped with the new span and tracer.
func CreateTracer(ctx context.Context, name string) (*Tracer, context.Context) {
	return NewTracer(Trace(ctx, name))
}

// GetTracer will return the tracer from the context or nil if absent.
func GetTracer(ctx context.Context) *Tracer {
	// check context
	if ctx == nil {
		return nil
	}

	// get tracer
	tracer, _ := ctx.Value(tracerKey).(*Tracer)

	return tracer
}

// SmartPush will call Push() with callers short name.
func (t *Tracer) SmartPush() {
	t.Push(GetCaller(1).Short)
}

// Push will add a new span onto the stack.
func (t *Tracer) Push(name string) {
	_, child := Trace(t.Tail().ctx, name)
	t.stack = append(t.stack, child)
}

// Rename will set a new name on the tail span.
func (t *Tracer) Rename(name string) {
	t.Tail().Rename(name)
}

// Tag will add the provided attribute to the tail span.
func (t *Tracer) Tag(key string, value interface{}) {
	t.Tail().Tag(key, value)
}

// Attach will add the provided event to the tail span.
func (t *Tracer) Attach(event string, attributes M) {
	t.Tail().Attach(event, attributes)
}

// Log will attach a log event to the tail span.
func (t *Tracer) Log(format string, args ...interface{}) {
	t.Tail().Log(format, args...)
}

// Record will attach an error event to the tail span.
func (t *Tracer) Record(err error) {
	t.Tail().Record(err)
}

// Pop ends and removes the last pushed span. This call is usually deferred
// right after a push.
func (t *Tracer) Pop() {
	// check list
	if len(t.stack) == 0 {
		return
	}

	// end last span
	t.stack[len(t.stack)-1].End()

	// resize stack
	t.stack = t.stack[:len(t.stack)-1]
}

// End will end all stacked spans and the root span.
func (t *Tracer) End() {
	// end stacked spans
	for _, span := range t.stack {
		span.End()
	}

	// end root span
	t.root.End()
}

// Tail returns the tail or root of the span stack.
func (t *Tracer) Tail() Span {
	// return last span if available
	if len(t.stack) > 0 {
		return t.stack[len(t.stack)-1]
	}

	return t.root
}

// Root will return the root of the span stack.
func (t *Tracer) Root() Span {
	return t.root
}
