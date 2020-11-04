package xo

import (
	"io"
	"time"

	"github.com/getsentry/sentry-go"
)

// Config is used to configure xo.
type Config struct {
	// The Sentry DSN.
	SentryDSN string

	// Whether to require a Sentry DSN if not development.
	RequireDSN bool

	// The debug config.
	DebugConfig DebugConfig

	// ReportOutput for writing events.
	//
	// Default: Sink("REPORT").
	ReportOutput io.Writer
}

// Auto will automatically install logging, reporting and tracing components.
// The returned function should be deferred to catch panics and ensure flushing.
func Auto(config Config) func() {
	// check sentry dsn
	if !Devel && config.RequireDSN && config.SentryDSN == "" {
		panic("missing required sentry dsn")
	}

	// run debug in development or when sentry DSN is missing
	if Devel || config.SentryDSN == "" {
		Debug(config.DebugConfig)
		return func() {}
	}

	// ensure report output
	if config.ReportOutput == nil {
		config.ReportOutput = Sink("REPORT")
	}

	// intercept
	reset := Intercept()

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

	return func() {
		// reset
		reset()

		// recover panics
		Recover(Capture)

		// await flush
		sentry.Flush(2 * time.Second)
	}
}
