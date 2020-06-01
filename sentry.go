package xo

import (
	"time"

	"github.com/getsentry/sentry-go"
)

// SetupSentry will setup error reporting using sentry. To simplify testing, the
// "ContextifyFrames" integration is removed.
func SetupSentry(dsn string) func() {
	// skip if benchmark
	if isBenchmark() {
		return func() {}
	}

	// initialize sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
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
		sentry.Flush(2 * time.Second)
	}
}

// Capture will capture the error.
func Capture(err error) {
	sentry.CaptureException(err)
}
