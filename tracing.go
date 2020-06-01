package xo

import (
	"context"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
)

// StartSpan will start a native span using the configured tracer. It will
// continue any span found in the context or start a new one if absent.
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	// ensure context
	if ctx == nil {
		ctx = context.Background()
	}

	// start span
	ctx, span := global.Tracer("xo").Start(ctx, name)

	return ctx, span
}

// GetSpan will return the first native span from the provided context. It will
// return nil if no span has been found.
func GetSpan(ctx context.Context) trace.Span {
	// check context
	if ctx == nil {
		return nil
	}

	// get span
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	return span
}
