package xo

import (
	"io"
	"time"
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

// Install will install logging, reporting and tracing components. The returned
// function may be called to teardown all installed components.
func Install(config Config) func() {
	// intercept
	var undoIntercept func()
	if !config.NoIntercept {
		undoIntercept = Intercept()
	}

	// create debugger
	debugger := NewDebugger(config)

	// create provider
	resetTracing := SetupTracing(debugger.SpanSyncer())

	// create client
	resetReporting := SetupReporting(debugger.SentryTransport())

	return func() {
		// reset reporting
		resetReporting()

		// set original provider
		resetTracing()

		// reset intercept
		if undoIntercept != nil {
			undoIntercept()
		}
	}
}
