package xo

import (
	"io"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/api/global"
	apiTrace "go.opentelemetry.io/otel/api/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// TODO: Add metrics?
// TODO: Add profiling?

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

	// The output for events.
	//
	// Default: Sink("EVENT").
	EventOutput io.Writer

	// Whether to omit event context data.
	NoEventContext bool

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
	tp, err := sdkTrace.NewProvider(
		sdkTrace.WithSyncer(debugger.SpanSyncer()),
		sdkTrace.WithConfig(sdkTrace.Config{
			DefaultSampler: sdkTrace.AlwaysSample(),
		}),
	)
	if err != nil {
		panic(err)
	}

	// set provider
	global.SetTraceProvider(tp)

	// initialize sentry
	err = sentry.Init(sentry.ClientOptions{
		Transport: debugger.SentryTransport(),
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			// filter integrations
			var list []sentry.Integration
			for _, integration := range integrations {
				if integration.Name() != "ContextifyFrames" {
					list = append(list, integration)
				}
			}

			return list
		},
	})
	if err != nil {
		panic(err)
	}

	return func() {
		// set noop transport
		defer func() {
			err := sentry.Init(sentry.ClientOptions{
				Transport: SentryTransport(func(*sentry.Event) {}),
			})
			if err != nil {
				panic(err)
			}
		}()

		// set noop provider
		global.SetTraceProvider(apiTrace.NoopProvider{})

		// reset intercept
		if interceptReset != nil {
			interceptReset()
		}
	}
}
