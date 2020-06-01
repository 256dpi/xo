package xo

import (
	"context"
	"math"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/api/global"
	apiTrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

func Trap(fn func(mock *Mock)) {
	// create mock
	mock := &Mock{
		Spans:       make([]MemorySpan, 0, 2048),
		Events:      make([]sentry.Event, 0, 2048),
		CleanEvents: true,
	}

	// create provider
	provider, err := sdkTrace.NewProvider(
		sdkTrace.WithSyncer(mock.SpanSyncer()),
		sdkTrace.WithConfig(sdkTrace.Config{
			DefaultSampler: sdkTrace.AlwaysSample(),
		}),
	)
	if err != nil {
		panic(err)
	}

	// set provider
	global.SetTraceProvider(provider)
	defer global.SetTraceProvider(apiTrace.NoopProvider{})

	// initialize sentry
	err = sentry.Init(sentry.ClientOptions{
		Transport: mock.SentryTransport(),
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			// filter integrations
			var list []sentry.Integration
			for _, integration := range integrations {
				if integration.Name() != "ContextifyFrames" {
					list = append(list, integration)
				}
			}

			return list
		},
	})
	if err != nil {
		panic(err)
	}

	// set noop transport
	defer func() {
		err = sentry.Init(sentry.ClientOptions{
			Transport: SentryTransport(func(*sentry.Event) {}),
		})
		if err != nil {
			panic(err)
		}
	}()

	// yield
	fn(mock)
}

type Mock struct {
	// Whether events should be cleaned.
	//
	// Default: true.
	CleanEvents bool

	// The collected spans.
	Spans []MemorySpan

	// The collected events.
	Events []sentry.Event
}

// ReducedSpans will return a copy of the span list with reduced information.
// This representation can be used in tests for easy direct comparison.
func (m *Mock) ReducedSpans(resolution time.Duration) []MemorySpan {
	// prepare list
	list := make([]MemorySpan, 0, len(m.Spans))

	// cleanup and copy spans
	for _, span := range m.Spans {
		// resize duration
		df := float64(span.Duration) / float64(resolution)
		duration := time.Duration(math.Round(df)) * resolution

		// cleanup span
		span.ID = ""
		span.Trace = ""
		span.Parent = ""
		span.Start = time.Time{}
		span.End = time.Time{}
		span.Duration = duration

		// add span
		list = append(list, span)
	}

	return list
}

// Reset all mock data.
func (m *Mock) Reset() {
	m.Spans = nil
	m.Events = nil
}

// SpanSyncer will return a span syncer that collects spans.
func (m *Mock) SpanSyncer() trace.SpanSyncer {
	return SpanSyncer(func(span *trace.SpanData) {
		m.Spans = append(m.Spans, traceSpanDataToMemorySpan(span))
	})
}

// SentryTransport will return a sentry transport that collects events.
func (m *Mock) SentryTransport() sentry.Transport {
	return SentryTransport(func(event *sentry.Event) {
		// clean if requested
		if m.CleanEvents {
			// unset meta data
			event.Timestamp = time.Time{}
			event.EventID = ""
			event.Platform = ""
			event.ServerName = ""
			event.Sdk = sentry.SdkInfo{}

			// cleanup contexts
			delete(event.Contexts, "device")
			delete(event.Contexts, "os")
			delete(event.Contexts, "runtime")
			if len(event.Contexts) == 0 {
				event.Contexts = nil
			}

			// cleanup extra
			if len(event.Extra) == 0 {
				event.Extra = nil
			}

			// cleanup tags
			if len(event.Tags) == 0 {
				event.Tags = nil
			}

			// rewrite line numbers
			for i := range event.Exception {
				st := event.Exception[i].Stacktrace
				for j := range st.Frames {
					if st.Frames[j].Lineno > 0 {
						st.Frames[j].Lineno = -1
					}
				}
			}
		}

		// add event
		m.Events = append(m.Events, *event)
	})
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
