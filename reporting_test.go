package xo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapture(t *testing.T) {
	Test(func(tester *Tester) {
		Capture(F("foo"))

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Exceptions: []VException{
					{
						Type:  "*xo.Err",
						Value: "foo",
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

	Test(func(tester *Tester) {
		Capture(errors.New("foo"))

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Exceptions: []VException{
					{
						Type:   "*errors.errorString",
						Value:  "foo",
						Frames: []VFrame{},
					},
					{
						Type:  "*xo.Err",
						Value: "foo",
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
								Func:   "TestCapture.func2",
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

		rep(F("foo"))

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Tags: M{
					"foo": "bar",
				},
				Exceptions: []VException{
					{
						Type:  "*xo.Err",
						Value: "foo",
						Frames: []VFrame{
							{
								Func:   "TestReporter",
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
								Func:   "TestReporter.func1",
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

	Test(func(tester *Tester) {
		rep := Reporter(SM{
			"foo": "bar",
		})

		rep(errors.New("foo"))

		assert.Equal(t, []VReport{
			{
				Level: "error",
				Tags: M{
					"foo": "bar",
				},
				Exceptions: []VException{
					{
						Type:   "*errors.errorString",
						Value:  "foo",
						Frames: []VFrame{},
					},
					{
						Type:  "*xo.Err",
						Value: "foo",
						Frames: []VFrame{
							{
								Func:   "TestReporter",
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
								Func:   "TestReporter.func2",
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
