package xo

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	exportTrace "go.opentelemetry.io/otel/sdk/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.5.0"
	"go.opentelemetry.io/otel/trace"
)

// atomic.Value does not work as it requires the same concrete type
var tracerCache sync.Map

var currentSpanKey any

type sniffContext struct {
	context.Context
}

func (s *sniffContext) Value(key any) any {
	currentSpanKey = key
	return nil
}

func init() {
	// set initial tracer
	tracerCache.Store("xo", otel.Tracer("xo"))

	// sniff current span key
	var ctx sniffContext
	trace.SpanFromContext(&ctx)
}

// GetGlobalTracer will return the global xo tracer. It will cache the tracer
// to increase performance between calls.
func GetGlobalTracer() trace.Tracer {
	// load from cache
	tracer, _ := tracerCache.Load("xo")

	return tracer.(trace.Tracer)
}

// ResetGlobalTracer will reset the global tracer cache.
func ResetGlobalTracer() {
	// set new tracer
	tracerCache.Store("xo", otel.Tracer("xo"))
}

// HookTracing will hook tracing using the provided span exporter. The returned
// function may be called to revert the previously configured provider.
func HookTracing(exporter exportTrace.SpanExporter, serviceName string, async bool) func() {
	// prepare span processor
	var spanProcessor sdkTrace.TracerProviderOption
	if async {
		spanProcessor = sdkTrace.WithBatcher(exporter)
	} else {
		spanProcessor = sdkTrace.WithSyncer(exporter)
	}

	// create provider
	provider := sdkTrace.NewTracerProvider(
		spanProcessor,
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// swap provider
	originalProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)

	// reset cache
	ResetGlobalTracer()

	return func() {
		// set original provider
		otel.SetTracerProvider(originalProvider)

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
