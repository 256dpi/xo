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

// Track is used to mark and annotate a function call. It will automatically
// wrap the context with a child from the span history found in the provided
// context. If no span history was found it will return a noop span.
//
// If the function finds a trace in the context and its root span matches
// the span from the context it will create a child from the traces tail.
// If not it considers the span history to have diverged from the trace.
func Track(ctx context.Context, name string) (context.Context, *Span) {
	// start span
	ctx, span := StartSpan(ctx, name)

	return ctx, &Span{
		ctx:  ctx,
		span: span,
	}
}

// Tag will tag the span.
func (s *Span) Tag(key string, value interface{}) {
	if s != nil && s.span != nil {
		s.span.SetAttribute(key, value)
	}
}

// Attach will add and event to the span.
func (s *Span) Attach(event string, attributes map[string]interface{}) {
	if s != nil && s.span != nil {
		s.span.AddEvent(s.ctx, event, mapToKV(attributes)...)
	}
}

// Log will add a "log" event to the span.
func (s *Span) Log(format string, args ...interface{}) {
	if s != nil && s.span != nil {
		s.span.AddEvent(s.ctx, "log", kv.Infer("message", fmt.Sprintf(format, args...)))
	}
}

// Record will add an "error" event to the span.
func (s *Span) Record(err error) {
	if s != nil && s.span != nil {
		s.span.RecordError(s.ctx, err)
	}
}

// End will end the span.
func (s *Span) End() {
	if s != nil && s.span != nil {
		s.span.End()
	}
}

// Native will return the underlying native span.
func (s *Span) Native() trace.Span {
	if s != nil {
		return s.span
	}

	return nil
}
