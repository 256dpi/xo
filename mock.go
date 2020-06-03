package xo

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

// Trap will temporarily intercept and collect logging, tracing and reporting
// data for testing purposes.
func Trap(fn func(mock *Mock)) {
	// create mock
	mock := &Mock{
		CleanEvents: true,
		Sinks:       map[string]*BufferSink{},
	}

	// setup tracing
	teardownTracing := SetupTracing(mock.SpanSyncer())
	defer teardownTracing()

	// setup reporting
	teardownReporting := SetupReporting(mock.SentryTransport())
	defer teardownReporting()

	// swap factory
	originalFactory := SinkFactory
	SinkFactory = mock.SinkFactory()
	defer func() {
		SinkFactory = originalFactory
	}()

	// yield
	fn(mock)
}

// Mock is a virtual logging, tracing and reporting provider.
type Mock struct {
	// Whether events should be cleaned.
	//
	// Default: true.
	CleanEvents bool

	// The collected spans.
	Spans []VSpan

	// The collected reports.
	Reports []VReport

	// The collected sinks.
	Sinks map[string]*BufferSink
}

// ReducedSpans will return a copy of the span list with reduced information.
// This representation can be used in tests for easy direct comparison.
func (m *Mock) ReducedSpans(resolution time.Duration) []VSpan {
	// prepare spans
	spans := make([]VSpan, 0, len(m.Spans))

	// cleanup and copy spans
	for _, span := range m.Spans {
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
func (m *Mock) ReducedReports() []VReport {
	// prepare reports
	reports := make([]VReport, 0, len(m.Reports))

	// cleanup and copy reports
	for _, report := range m.Reports {
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

// Reset all mock data.
func (m *Mock) Reset() {
	m.Spans = nil
	m.Reports = nil
}

// SpanSyncer will return a span syncer that collects spans.
func (m *Mock) SpanSyncer() trace.SpanSyncer {
	return SpanSyncer(func(span *trace.SpanData) {
		m.Spans = append(m.Spans, convertSpan(span))
	})
}

// SentryTransport will return a sentry transport that collects events.
func (m *Mock) SentryTransport() sentry.Transport {
	return SentryTransport(func(event *sentry.Event) {
		m.Reports = append(m.Reports, convertReport(event))
	})
}

// SinkFactory will return a sink factory that returns buffer sinks.
func (m *Mock) SinkFactory() func(name string) io.WriteCloser {
	return func(name string) io.WriteCloser {
		// check sinks
		if sink, ok := m.Sinks[name]; ok {
			return sink
		}

		// create buffer
		buf := &BufferSink{
			Buffer: new(bytes.Buffer),
		}

		// store sink
		m.Sinks[name] = buf

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
