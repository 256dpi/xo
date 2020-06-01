package xo

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/api/global"
	apiTrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export/trace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// TODO: Add logging.
// TODO: Add metrics.

type DebuggerConfig struct {
	TraceOutput     io.Writer
	TraceResolution time.Duration
	EventOutput     io.Writer
}

func (c *DebuggerConfig) Ensure() {
	// set default trace output
	if c.TraceOutput == nil {
		c.TraceOutput = os.Stdout
	}

	// set default trace resolution
	if c.TraceResolution == 0 {
		c.TraceResolution = time.Nanosecond
	}

	// set default event output
	if c.EventOutput == nil {
		c.EventOutput = os.Stdout
	}
}

type Debugger struct {
	config DebuggerConfig
	spans  map[apiTrace.SpanID]*trace.SpanData
	mutex  sync.Mutex
}

func NewDebugger(config DebuggerConfig) *Debugger {
	// ensure config
	config.Ensure()

	return &Debugger{
		config: config,
		spans:  make(map[apiTrace.SpanID]*trace.SpanData, 2048),
	}
}

// SpanSyncer will return a span syncer that prints received spans.
func (d *Debugger) SpanSyncer() trace.SpanSyncer {
	return SpanSyncer(func(span *trace.SpanData) {
		// acquire mutex
		d.mutex.Lock()
		defer d.mutex.Unlock()

		// store trace if not root
		if span.ParentSpanID.IsValid() && !span.HasRemoteParent {
			d.spans[span.SpanContext.SpanID] = span
			return
		}

		// collect spans
		table := make(map[apiTrace.SpanID]*trace.SpanData)
		list := make([]*trace.SpanData, 0, 512)
		for id, s := range d.spans {
			if s.SpanContext.TraceID == span.SpanContext.TraceID {
				list = append(list, s)
				delete(d.spans, id)
				table[s.SpanContext.SpanID] = s
			}
		}

		// add root
		list = append(list, span)

		// TODO: Transform list into a tree and print.

		// sort list
		sort.Slice(list, func(i, j int) bool {
			return list[i].StartTime.Before(list[j].StartTime)
		})

		// prepare buffer
		var buf bytes.Buffer

		// print header
		_, _ = fmt.Fprintf(&buf, "----- TRACE -----\n")

		// print spans
		for _, span := range list {
			// compute duration
			df := float64(span.EndTime.Sub(span.StartTime)) / float64(d.config.TraceResolution)
			duration := time.Duration(math.Round(df)) * d.config.TraceResolution

			// print span
			_, _ = fmt.Fprintf(&buf, "%s: %s\n", span.Name, duration.String())
		}

		// write trace
		_, _ = buf.WriteTo(d.config.TraceOutput)
	})
}

// SentryTransport will return a sentry transport that print received events.
func (d *Debugger) SentryTransport() sentry.Transport {
	return SentryTransport(func(event *sentry.Event) {
		// reverse stack traces
		for i := range event.Exception {
			st := event.Exception[i].Stacktrace
			for i, j := 0, len(st.Frames)-1; i < j; i, j = i+1, j-1 {
				st.Frames[i], st.Frames[j] = st.Frames[j], st.Frames[i]
			}
		}

		// prepare buffer
		var buf bytes.Buffer

		// print header
		_, _ = fmt.Fprintf(&buf, "----- EVENT -----\n")

		// print info
		_, _ = fmt.Fprintf(&buf, "Level: %s\n", event.Level)

		// print context
		_, _ = fmt.Fprintf(&buf, "Context:\n")
		iterateMap(event.Contexts, func(key string, value interface{}) {
			_, _ = fmt.Fprintf(&buf, "- %s: %v\n", key, mustEncode(value))
		})

		// print exceptions
		_, _ = fmt.Fprintf(&buf, "Exceptions:\n")
		for _, exc := range event.Exception {
			_, _ = fmt.Fprintf(&buf, "- %s (%s)\n", exc.Value, exc.Type)
			for _, frame := range exc.Stacktrace.Frames {
				_, _ = fmt.Fprintf(&buf, "  > %s (%s): %s:%d\n", frame.Function, frame.Module, frame.AbsPath, frame.Lineno)
			}
		}

		// write event
		_, _ = buf.WriteTo(d.config.EventOutput)
	})
}

func SetupDebugger(config DebuggerConfig) {
	// create debugger
	debugger := NewDebugger(config)

	// create provider
	tp, err := sdkTrace.NewProvider(
		sdkTrace.WithSyncer(debugger.SpanSyncer()),
		sdkTrace.WithConfig(sdkTrace.Config{
			DefaultSampler: sdkTrace.AlwaysSample(),
		}),
	)
	if err != nil {
		panic(err)
	}

	// set provider
	global.SetTraceProvider(tp)

	// initialize sentry
	err = sentry.Init(sentry.ClientOptions{
		Transport: debugger.SentryTransport(),
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
}

func TeardownDebugger() {
	// set noop provider
	global.SetTraceProvider(apiTrace.NoopProvider{})

	// set noop transport
	defer func() {
		err := sentry.Init(sentry.ClientOptions{
			Transport: SentryTransport(func(*sentry.Event) {}),
		})
		if err != nil {
			panic(err)
		}
	}()
}
