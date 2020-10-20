package xo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTest(t *testing.T) {
	Test(func(tester *Tester) {
		// logging
		sink := Sink("sink")
		_, _ = sink.Write([]byte("sink"))
		_ = sink.Close()
		assert.Equal(t, map[string]*BufferSink{
			"sink": {String: "sink"},
		}, tester.Sinks)

		// tracing
		_, span := Trace(nil, "span")
		span.End()
		assert.Equal(t, []VSpan{
			{Name: "span"},
		}, tester.ReducedSpans(time.Second))

		// reporting
		Capture(F("error"))
		assert.Equal(t, []VReport{
			{
				Level: "error",
				Exceptions: []VException{
					{Type: "*xo.Err", Value: "error"},
				},
			},
		}, tester.ReducedReports(false))
	})
}
