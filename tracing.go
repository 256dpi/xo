package xo

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	exportTrace "go.opentelemetry.io/otel/sdk/export/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

type globalTracer struct {
	tracer trace.Tracer
}

var cachedTracer atomic.Value

func init() {
	cachedTracer.Store(globalTracer{})
}

// GetGlobalTracer will return the global xo tracer. It will cache the tracer
// to increase performance between calls.
func GetGlobalTracer() trace.Tracer {
	// load from cache
	gt := cachedTracer.Load().(globalTracer)

	// store missing tracer
	if gt.tracer == nil {
		gt = globalTracer{tracer: global.Tracer("xo")}
		cachedTracer.Store(gt)
	}

	return gt.tracer
}

// ResetGlobalTracer will reset the cache global tracer.
func ResetGlobalTracer() {
	// reset cache
	cachedTracer.Store(globalTracer{})
}

// HookTracing will hook tracing using the provided span exporter. The returned
// function may be called to revert the previously configured provider.
func HookTracing(exporter exportTrace.SpanExporter) func() {
	// create provider
	provider := sdkTrace.NewTracerProvider(
		sdkTrace.WithSyncer(exporter),
		sdkTrace.WithConfig(sdkTrace.Config{
			DefaultSampler: sdkTrace.AlwaysSample(),
		}),
	)

	// swap provider
	originalProvider := global.TracerProvider()
	global.SetTracerProvider(provider)

	// reset cache
	ResetGlobalTracer()

	return func() {
		// set original provider
		global.SetTracerProvider(originalProvider)

		// reset cache
		ResetGlobalTracer()
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
	ctx, span := GetGlobalTracer().Start(ctx, name)

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
