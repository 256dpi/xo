package xo

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

// Test will temporarily intercept and collect logging, tracing and reporting
// data for testing purposes.
func Test(fn func(tester *Tester)) {
	// create tester
	tester := &Tester{
		Sinks: map[string]*BufferSink{},
	}

	// setup tracing
	teardownTracing := SetupTracing(tester.SpanSyncer())
	defer teardownTracing()

	// setup reporting
	teardownReporting := SetupReporting(tester.SentryTransport())
	defer teardownReporting()

	// swap factory
	originalFactory := SinkFactory
	SinkFactory = tester.SinkFactory()
	defer func() {
		SinkFactory = originalFactory
	}()

	// yield
	fn(tester)
}

// Tester is a virtual logging, tracing and reporting provider for testing
// purposes.
type Tester struct {
	// The collected spans.
	Spans []VSpan

	// The collected reports.
	Reports []VReport

	// The collected sinks.
	Sinks map[string]*BufferSink
}

// ReducedSpans will return a copy of the span list with reduced information.
// This representation can be used in tests for easy direct comparison.
func (t *Tester) ReducedSpans(resolution time.Duration) []VSpan {
	// prepare spans
	spans := make([]VSpan, 0, len(t.Spans))

	// cleanup and copy spans
	for _, span := range t.Spans {
		// recalculate duration
		span.Start = span.Start.Round(resolution)
		span.End = span.End.Round(resolution)
		duration := span.End.Sub(span.Start)
		if resolution == 0 {
			duration = 0
		}

		// cleanup span
		span.ID = ""
		span.Trace = ""
		span.Parent = ""
		span.Start = time.Time{}
		span.End = time.Time{}
		span.Duration = duration

		// copy events
		if len(span.Events) > 0 {
			// prepare events
			events := make([]VEvent, 0, len(span.Events))
			for _, event := range span.Events {
				event.Time = time.Time{}
				events = append(events, event)
			}

			// set event
			span.Events = events
		}

		// add span
		spans = append(spans, span)
	}

	return spans
}

// ReducedReports will return a copy of the report list with reduced information.
// This representation can be used in tests for easy direct comparison.
func (t *Tester) ReducedReports() []VReport {
	// prepare reports
	reports := make([]VReport, 0, len(t.Reports))

	// cleanup and copy reports
	for _, report := range t.Reports {
		// cleanup report
		report.ID = ""
		report.Time = time.Time{}

		// cleanup context
		delete(report.Context, "device")
		delete(report.Context, "os")
		delete(report.Context, "runtime")
		if len(report.Context) == 0 {
			report.Context = nil
		}

		// copy exceptions
		exceptions := make([]VException, 0, len(report.Exceptions))
		for _, exc := range report.Exceptions {
			// copy frames
			frames := make([]VFrame, 0, len(exc.Frames))
			for _, frame := range exc.Frames {
				frame.Line = 0
				frames = append(frames, frame)
			}

			// set frames
			exc.Frames = frames

			// add exception
			exceptions = append(exceptions, exc)
		}

		// set exceptions
		report.Exceptions = exceptions

		// add report
		reports = append(reports, report)
	}

	return reports
}

// Reset collected spans, reports and sinks.
func (t *Tester) Reset() {
	t.Spans = nil
	t.Reports = nil
	t.Sinks = map[string]*BufferSink{}
}

// SpanSyncer will return a span syncer that collects spans.
func (t *Tester) SpanSyncer() trace.SpanSyncer {
	return SpanSyncer(func(span *trace.SpanData) {
		t.Spans = append(t.Spans, ConvertSpan(span))
	})
}

// SentryTransport will return a sentry transport that collects events.
func (t *Tester) SentryTransport() sentry.Transport {
	return SentryTransport(func(event *sentry.Event) {
		t.Reports = append(t.Reports, ConvertReport(event))
	})
}

// SinkFactory will return a sink factory that returns buffer sinks.
func (t *Tester) SinkFactory() func(name string) io.WriteCloser {
	return func(name string) io.WriteCloser {
		// check sinks
		if sink, ok := t.Sinks[name]; ok {
			return sink
		}

		// create buffer
		buf := &BufferSink{
			Buffer: new(bytes.Buffer),
		}

		// store sink
		t.Sinks[name] = buf

		return buf
	}
}

// SpanSyncer is a functional span exporter.
type SpanSyncer func(*trace.SpanData)

// ExportSpan implements the trace.SpanSyncer interface.
func (s SpanSyncer) ExportSpan(_ context.Context, span *trace.SpanData) {
	s(span)
}

// SentryTransport is a functional sentry transport.
type SentryTransport func(*sentry.Event)

// Configure implements the sentry.Transport interface.
func (t SentryTransport) Configure(sentry.ClientOptions) {}

// SendEvent implements the sentry.Transport interface.
func (t SentryTransport) SendEvent(event *sentry.Event) {
	t(event)
}

// Flush implements the sentry.Transport interface.
func (t SentryTransport) Flush(time.Duration) bool {
	return true
}

// BufferSink wraps a bytes buffer.
type BufferSink struct {
	*bytes.Buffer
}

// Close implements the io.Closer interface.
func (s *BufferSink) Close() error {
	// unset
	s.Buffer = nil

	return nil
}
