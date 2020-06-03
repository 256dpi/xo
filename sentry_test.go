package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var someError = F("some error")

func TestCapture(t *testing.T) {
	Trap(func(mock *Mock) {
		Capture(W(someError))

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Exceptions: []VException{
					{
						Type:  "*xo.Err",
						Value: "some error",
						Frames: []VFrame{
							{
								Func:   "init",
								Module: "github.com/256dpi/xo",
								File:   "sentry_test.go",
								Path:   "github.com/256dpi/xo/sentry_test.go",
							},
						},
					},
					{
						Type:  "*xo.Err",
						Value: "some error",
						Frames: []VFrame{
							{
								Func:   "TestCapture",
								Module: "github.com/256dpi/xo",
								File:   "sentry_test.go",
								Path:   "github.com/256dpi/xo/sentry_test.go",
							},
							{
								Func:   "Trap",
								Module: "github.com/256dpi/xo",
								File:   "mock.go",
								Path:   "github.com/256dpi/xo/mock.go",
							},
							{
								Func:   "TestCapture.func1",
								Module: "github.com/256dpi/xo",
								File:   "sentry_test.go",
								Path:   "github.com/256dpi/xo/sentry_test.go",
							},
						},
					},
				},
			},
		}, mock.ReducedReports())
	})
}

func TestReporter(t *testing.T) {
	Trap(func(mock *Mock) {
		rep := Reporter(SM{
			"foo": "bar",
		})

		rep(someError)

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Tags: SM{
					"foo": "bar",
				},
				Exceptions: []VException{
					{
						Type:  "*xo.Err",
						Value: "some error",
						Frames: []VFrame{
							{
								Func:   "init",
								Module: "github.com/256dpi/xo",
								File:   "sentry_test.go",
								Path:   "github.com/256dpi/xo/sentry_test.go",
							},
						},
					},
				},
			},
		}, mock.ReducedReports())
	})
}
