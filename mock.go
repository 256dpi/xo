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
		Spans:           make([]MockSpan, 0, 2048),
		Events:          make([]sentry.Event, 0, 2048),
		TraceResolution: 10 * time.Millisecond,
		CleanEvents:     true,
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

type MockSpanEvent struct {
	// The event name.
	Name string

	// Attached event attributes.
	Attributes map[string]interface{}
}

type MockSpan struct {
	// The span name.
	Name string

	// The span duration.
	Duration time.Duration

	// The span attributes.
	Attributes map[string]interface{}

	// Attached span events.
	Events []MockSpanEvent
}

type Mock struct {
	// The collected spans.
	Spans []MockSpan

	// The collected events.
	Events []sentry.Event

	// The used trace resolution.
	//
	// Default: 10ms.
	TraceResolution time.Duration

	// Whether events should be cleaned.
	//
	// Default: true.
	CleanEvents bool
}

// SpanSyncer will return a span syncer that stores received spans.
func (m *Mock) SpanSyncer() trace.SpanSyncer {
	return SpanSyncer(func(span *trace.SpanData) {
		// collect events
		var events []MockSpanEvent
		for _, event := range span.MessageEvents {
			events = append(events, MockSpanEvent{
				Name:       event.Name,
				Attributes: otelKVToMap(event.Attributes),
			})
		}

		// compute duration
		df := float64(span.EndTime.Sub(span.StartTime)) / float64(m.TraceResolution)
		duration := time.Duration(math.Round(df)) * m.TraceResolution

		// add span
		m.Spans = append(m.Spans, MockSpan{
			Name:       span.Name,
			Duration:   duration,
			Attributes: otelKVToMap(span.Attributes),
			Events:     events,
		})
	})
}

// SentryTransport will return a sentry transport that stores received events.
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
