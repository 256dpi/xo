package xo

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// Config is used to configure xo.
type Config struct {
	// The Sentry DSN.
	SentryDSN string

	// The OTLP endpoint URL.
	OTLPEndpointURL string

	// The trace service name.
	TraceServiceName string

	// ReportOutput for writing errors.
	//
	// Default: os.Stderr.
	ReportOutput io.Writer

	// The debug config.
	DebugConfig DebugConfig
}

// Auto will automatically install logging, reporting and tracing components.
// The returned function should be deferred to catch panics and ensure flushing.
func Auto(config Config) func() {
	// check if development
	if Devel {
		Debug(config.DebugConfig)
		return func() {}
	}

	// ensure report output
	if config.ReportOutput == nil {
		config.ReportOutput = os.Stderr
	}

	// check sentry dsn
	if config.SentryDSN == "" {
		panic("missing required sentry dsn")
	}

	// prepare debugger
	debugger := NewDebugger(DebugConfig{
		ReportOutput: config.ReportOutput,
	})

	// init sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:          config.SentryDSN,
		Integrations: FilterSentryIntegrations("ContextifyFrames"),
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// check event silent tag
			if event.Tags["xo:silent"] == "true" {
				delete(event.Tags, "silent")
			} else {
				debugger.Report(event)
			}

			return event
		},
	})
	if err != nil {
		Panic(err)
	}

	// install OTLP if provided
	if config.OTLPEndpointURL != "" {
		exporter, err := otlp.New(context.Background(), otlp.WithEndpointURL(config.OTLPEndpointURL))
		if err != nil {
			Capture(err)
		} else {
			HookTracing(exporter, config.TraceServiceName, true)
		}
	}

	return func() {
		// recover panics
		Recover(Capture)

		// flush
		sentry.Flush(time.Second)
	}
}
