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

// Reporter will return a capture function that adds the provided tags.
func Reporter(tags SM) func(error) {
	// prepare scope
	scope := sentry.NewScope()
	scope.SetTags(tags)

	return func(err error) {
		// ensure caller
		if _, ok := err.(*Err); !ok {
			err = W(err)
		}

		// get client
		client := sentry.CurrentHub().Client()

		// forward exception
		client.CaptureException(err, nil, scope)
	}
}

// SetupReporting will setup error reporting using sentry. The returned
// function may be called to teardown the component.
func SetupReporting(transport sentry.Transport) func() {
	// create client
	client, err := sentry.NewClient(sentry.ClientOptions{
		Transport:    transport,
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
		// flush
		sentry.Flush(2 * time.Second)

		// set original client
		hub.BindClient(originalClient)
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
