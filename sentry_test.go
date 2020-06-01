package xo

import (
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
)

var someError = F("some error")

func TestSentry(t *testing.T) {
	Trap(func(mock *Mock) {
		sentry.CaptureException(W(someError))
		sentry.Flush(time.Second)

		assert.Equal(t, []sentry.Event{
			{
				Level: "error",
				Exception: []sentry.Exception{
					{
						Type:     "*errors.fundamental",
						Value:    "some error",
						Module:   "",
						ThreadID: "",
						Stacktrace: &sentry.Stacktrace{
							Frames: []sentry.Frame{
								{
									Function: "init",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "/Users/256dpi/Development/GitHub/256dpi/xo/sentry_test.go",
									Lineno:   -1,
									InApp:    true,
								},
							},
						},
					},
					{
						Type:  "*errors.withStack",
						Value: "some error",
						Stacktrace: &sentry.Stacktrace{
							Frames: []sentry.Frame{
								{
									Function: "TestSentry",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "/Users/256dpi/Development/GitHub/256dpi/xo/sentry_test.go",
									Lineno:   -1,
									InApp:    true,
								},
								{
									Function: "Trap",
									Module:   "github.com/256dpi/xo",
									Filename: "mock.go",
									AbsPath:  "/Users/256dpi/Development/GitHub/256dpi/xo/mock.go",
									Lineno:   -1,
									InApp:    true,
								},
								{
									Function: "TestSentry.func1",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "/Users/256dpi/Development/GitHub/256dpi/xo/sentry_test.go",
									Lineno:   -1,
									InApp:    true,
								},
							},
						},
					},
				},
			},
		}, mock.Events)
	})
}
