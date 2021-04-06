package xo

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Span is the underlying span used for tracing.
type Span struct {
	ctx  context.Context
	span trace.Span
}

// SmartTrace calls Trace with the callers short name.
func SmartTrace(ctx context.Context) (context.Context, Span) {
	return Trace(ctx, GetCaller(1, 1).Short)
}

// Trace is used to trace a function call. It will start a new span based on the
// history in the provided context. It will return a new span and context that
// contains the created spans native span.
func Trace(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := StartSpan(ctx, name)
	return ctx, NewSpan(ctx, span)
}

// NewSpan will create and return a new span from the provided context and
// native span. The context should already carry the native span.
func NewSpan(ctx context.Context, span trace.Span) Span {
	return Span{
		ctx:  ctx,
		span: span,
	}
}

// Rename will set a new name on the span.
func (s Span) Rename(name string) {
	s.span.SetName(name)
}

// Tag will add the provided attribute to the span.
func (s Span) Tag(key string, value interface{}) {
	// get label
	var kv attribute.KeyValue
	switch v := value.(type) {
	case bool:
		kv = attribute.Bool(key, v)
	case int:
		kv = attribute.Int(key, v)
	case int64:
		kv = attribute.Int64(key, v)
	case string:
		kv = attribute.String(key, v)
	default:
		kv = attribute.Any(key, convertValue(value))
	}

	// set attribute
	s.span.SetAttributes(kv)
}

// Attach will add the provided event to the span.
func (s Span) Attach(event string, attributes M) {
	s.span.AddEvent(event, trace.WithAttributes(mapToKV(attributes)...))
}

// Log will attach a log event to the span.
func (s Span) Log(format string, args ...interface{}) {
	s.span.AddEvent("log", trace.WithAttributes(attribute.String("message", fmt.Sprintf(format, args...))))
}

// Record will attach an error event to the span.
func (s Span) Record(err error) {
	s.span.RecordError(err)
}

// End will end the span.
func (s Span) End() {
	s.span.End()
}

// Context will return the context carrying the native span.
func (s Span) Context() context.Context {
	return s.ctx
}

// Native will return the underlying native span.
func (s Span) Native() trace.Span {
	return s.span
}
