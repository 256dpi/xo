package xo

import (
	"io"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/api/global"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config is used to configure xo.
type Config struct {
	// Whether to omit interception.
	//
	// Default: false.
	NoIntercept bool

	// The output for traces.
	//
	// Default: Sink("TRACE").
	TraceOutput io.Writer

	// The trace resolution.
	//
	// Default: 1ns.
	TraceResolution time.Duration

	// The trace width.
	//
	// Default: 80.
	TraceWidth int

	// Whether to omit trace attributes.
	NoTraceAttributes bool

	// The output for events.
	//
	// Default: Sink("EVENT").
	EventOutput io.Writer

	// Whether to omit event context data.
	NoEventContext bool

	// Whether to omit file paths from event stack traces.
	NoEventPaths bool

	// Whether to omit line numbers from event stack traces.
	NoEventLineNumbers bool
}

// Ensure will ensure defaults.
func (c *Config) Ensure() {
	// set default trace output
	if c.TraceOutput == nil {
		c.TraceOutput = Sink("TRACE")
	}

	// set default trace resolution
	if c.TraceResolution == 0 {
		c.TraceResolution = time.Nanosecond
	}

	// set default trace width
	if c.TraceWidth == 0 {
		c.TraceWidth = 80
	}

	// set default event output
	if c.EventOutput == nil {
		c.EventOutput = Sink("EVENT")
	}
}

// Install will install logging, reporting and tracing components.
func Install(config Config) func() {
	// intercept
	var interceptReset func()
	if !config.NoIntercept {
		interceptReset = Intercept()
	}

	// create debugger
	debugger := NewDebugger(config)

	// create provider
	provider, err := sdkTrace.NewProvider(
		sdkTrace.WithSyncer(debugger.SpanSyncer()),
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

	// create client
	client, err := sentry.NewClient(sentry.ClientOptions{
		Transport:    debugger.SentryTransport(),
		Integrations: FilterIntegrations("ContextifyFrames"),
	})
	if err != nil {
		panic(err)
	}

	// swap client
	hub := sentry.CurrentHub()
	originalClient := hub.Client()
	hub.BindClient(client)

	return func() {
		// set original client
		hub.BindClient(originalClient)

		// set original provider
		global.SetTraceProvider(originalProvider)

		// reset intercept
		if interceptReset != nil {
			interceptReset()
		}
	}
}
