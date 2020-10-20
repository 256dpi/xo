package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var someError = F("some error")

func TestCapture(t *testing.T) {
	Test(func(tester *Tester) {
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
								File:   "reporting_test.go",
								Path:   "github.com/256dpi/xo/reporting_test.go",
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
								File:   "reporting_test.go",
								Path:   "github.com/256dpi/xo/reporting_test.go",
							},
							{
								Func:   "Test",
								Module: "github.com/256dpi/xo",
								File:   "tester.go",
								Path:   "github.com/256dpi/xo/tester.go",
							},
							{
								Func:   "TestCapture.func1",
								Module: "github.com/256dpi/xo",
								File:   "reporting_test.go",
								Path:   "github.com/256dpi/xo/reporting_test.go",
							},
						},
					},
				},
			},
		}, tester.ReducedReports(true))
	})
}

func TestReporter(t *testing.T) {
	Test(func(tester *Tester) {
		rep := Reporter(SM{
			"foo": "bar",
		})

		rep(someError)

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Tags: M{
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
								File:   "reporting_test.go",
								Path:   "github.com/256dpi/xo/reporting_test.go",
							},
						},
					},
				},
			},
		}, tester.ReducedReports(true))
	})
}
