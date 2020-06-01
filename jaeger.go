package xo

import (
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SetupLocalJaeger will configure the local jaeger instance for tracing.
func SetupLocalJaeger(name string) func() {
	// skip if benchmark
	if isBenchmark() {
		return func() {}
	}

	// create export pipeline
	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: name,
		}),
		jaeger.RegisterAsGlobal(),
		jaeger.WithSDK(&trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		}),
	)
	if err != nil {
		panic(err)
	}

	return func() {
		flush()
	}
}
