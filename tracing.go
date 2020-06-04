package xo

import (
	"context"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	export "go.opentelemetry.io/otel/sdk/export/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupTracing will setup tracing using the provided span syncer. The returned
// function may be called to revert the previously configured provider.
func SetupTracing(syncer export.SpanSyncer) func() {
	// create provider
	provider, err := sdkTrace.NewProvider(
		sdkTrace.WithSyncer(syncer),
		sdkTrace.WithConfig(sdkTrace.Config{
			DefaultSampler: sdkTrace.AlwaysSample(),
		}),
	)
	if err != nil {
		panic(err)
	}

	// wap provider
	originalProvider := global.TraceProvider()
	global.SetTraceProvider(provider)

	return func() {
		// set original provider
		global.SetTraceProvider(originalProvider)
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
	} else if _, ok := span.(trace.NoopSpan); ok {
		return nil
	}

	return span
}
