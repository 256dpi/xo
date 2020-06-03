package xo

import (
	"time"

	"github.com/getsentry/sentry-go"
)

// Capture will capture the error.
func Capture(err error) {
	// ensure caller
	if _, ok := err.(*Err); !ok {
		err = W(err)
	}

	// forward exception
	sentry.CaptureException(err)
}

// SetupSentry will setup error reporting using sentry. To simplify testing, the
// "ContextifyFrames" integration is removed.
func SetupSentry(dsn string) func() {
	// initialize sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:          dsn,
		Integrations: FilterIntegrations("ContextifyFrames"),
	})
	if err != nil {
		panic(err)
	}

	return func() {
		sentry.Flush(2 * time.Second)
	}
}

// FilterIntegrations will return a sentry integrations filter that will drop
// the named integrations.
func FilterIntegrations(drop ...string) func(i []sentry.Integration) []sentry.Integration {
	return func(integrations []sentry.Integration) []sentry.Integration {
		// filter integrations
		var list []sentry.Integration
		for _, integration := range integrations {
			var found bool
			for _, name := range drop {
				if integration.Name() == name {
					found = true
				}
			}
			if !found {
				list = append(list, integration)
			}
		}

		return list
	}
}
