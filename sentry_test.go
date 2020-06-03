package xo

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
)

var someError = F("some error")

func TestCapture(t *testing.T) {
	Trap(func(mock *Mock) {
		Capture(W(someError))

		assert.Equal(t, []sentry.Event{
			{
				Level: "error",
				Exception: []sentry.Exception{
					{
						Type:  "*xo.Err",
						Value: "some error",
						Stacktrace: &sentry.Stacktrace{
							Frames: []sentry.Frame{
								{
									Function: "init",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "github.com/256dpi/xo/sentry_test.go",
									InApp:    true,
								},
							},
						},
					},
					{
						Type:  "*xo.Err",
						Value: "some error",
						Stacktrace: &sentry.Stacktrace{
							Frames: []sentry.Frame{
								{
									Function: "TestCapture",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "github.com/256dpi/xo/sentry_test.go",
									InApp:    true,
								},
								{
									Function: "Trap",
									Module:   "github.com/256dpi/xo",
									Filename: "mock.go",
									AbsPath:  "github.com/256dpi/xo/mock.go",
									InApp:    true,
								},
								{
									Function: "TestCapture.func1",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "github.com/256dpi/xo/sentry_test.go",
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

func TestReporter(t *testing.T) {
	Trap(func(mock *Mock) {
		rep := Reporter(SM{
			"foo": "bar",
		})

		rep(someError)

		assert.Equal(t, []sentry.Event{
			{
				Level: "error",
				Tags: SM{
					"foo": "bar",
				},
				Exception: []sentry.Exception{
					{
						Type:  "*xo.Err",
						Value: "some error",
						Stacktrace: &sentry.Stacktrace{
							Frames: []sentry.Frame{
								{
									Function: "init",
									Module:   "github.com/256dpi/xo",
									Filename: "sentry_test.go",
									AbsPath:  "github.com/256dpi/xo/sentry_test.go",
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
