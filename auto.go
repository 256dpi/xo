package xo

import (
	"io"
	"time"

	"github.com/getsentry/sentry-go"
)

// Config is used to configure xo.
type Config struct {
	// ReportOutput for writing events.
	//
	// Default: Sink("REPORT").
	ReportOutput io.Writer

	// The Sentry DSN.
	SentryDSN string

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

	// prepare finalizers
	var finalizers []func()

	/* Logging */

	// intercept
	reset := Intercept()

	// add reset finalizer
	finalizers = append(finalizers, reset)

	/* Reporting */

	// ensure report output
	if config.ReportOutput == nil {
		config.ReportOutput = Sink("REPORT")
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
			debugger.Report(event)
			return event
		},
	})
	if err != nil {
		Panic(err)
	}

	// add flush finalizer
	finalizers = append(finalizers, func() {
		sentry.Flush(time.Second)
	})

	return func() {
		// recover panics
		Recover(Capture)

		// run finalizers
		for _, fn := range finalizers {
			fn()
		}
	}
}
