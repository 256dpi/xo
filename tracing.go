package xo

import (
	"context"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	exportTrace "go.opentelemetry.io/otel/sdk/export/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupTracing will setup tracing using the provided span exporter. The
// returned function may be called to revert the previously configured provider.
func SetupTracing(exporter exportTrace.SpanExporter) func() {
	// create provider
	provider := sdkTrace.NewTracerProvider(
		sdkTrace.WithSyncer(exporter),
		sdkTrace.WithConfig(sdkTrace.Config{
			DefaultSampler: sdkTrace.AlwaysSample(),
		}),
	)

	// wap provider
	originalProvider := global.TracerProvider()
	global.SetTracerProvider(provider)

	return func() {
		// set original provider
		global.SetTracerProvider(originalProvider)
	}
}

// StartSpan will start a native span using the globally configured tracer. It
// will continue any span found in the provided context or start a new span with
// the specified name if absent.
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	// ensure context
	if ctx == nil {
		ctx = context.Background()
	}

	// start span
	ctx, span := global.Tracer("xo").Start(ctx, name)

	return ctx, span
}

// GetSpan will return the first native span found in the provided context. It
// will return nil if no span has been found.
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
