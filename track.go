package xo

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
)

// Span is the underlying span used for tracing.
type Span struct {
	ctx  context.Context
	span trace.Span
}

// SmartTrack calls Track with the callers short name.
func SmartTrack(ctx context.Context) (context.Context, Span) {
	return Track(ctx, GetCaller(1).Short)
}

// Track is used to mark and annotate a function call. It will automatically
// wrap the context with a child from the span history found in the provided
// context. If no span history was found it will start a new span.
func Track(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := StartSpan(ctx, name)
	return ctx, NewSpan(ctx, span)
}

// NewSpan will create and return a new span from the provided context and
// native span. The context should already carry the span as if created with
// StartSpan().
func NewSpan(ctx context.Context, span trace.Span) Span {
	return Span{
		ctx:  ctx,
		span: span,
	}
}

// Rename will set a new name on the span.
func (s Span) Rename(name string) {
	if s.span != nil {
		s.span.SetName(name)
	}
}

// Tag will add the provided attribute to the span.
func (s Span) Tag(key string, value interface{}) {
	if s.span != nil {
		s.span.SetAttribute(key, convertValue(value))
	}
}

// Attach will add the provided event to the span.
func (s Span) Attach(event string, attributes M) {
	if s.span != nil {
		s.span.AddEvent(s.ctx, event, mapToKV(attributes)...)
	}
}

// Log will attach a log event to the span.
func (s Span) Log(format string, args ...interface{}) {
	if s.span != nil {
		s.span.AddEvent(s.ctx, "log", kv.Infer("message", fmt.Sprintf(format, args...)))
	}
}

// Record will attach an error event to the span.
func (s Span) Record(err error) {
	if s.span != nil {
		s.span.RecordError(s.ctx, err)
	}
}

// End will end the span.
func (s Span) End() {
	if s.span != nil {
		s.span.End()
	}
}

// Context will return the context carrying the native span.
func (s Span) Context() context.Context {
	return s.ctx
}

// Native will return the underlying native span.
func (s Span) Native() trace.Span {
	return s.span
}
