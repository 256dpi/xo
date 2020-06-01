package xo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
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
	spans  map[string]MemorySpan
	mutex  sync.Mutex
}

func NewDebugger(config DebuggerConfig) *Debugger {
	// ensure config
	config.Ensure()

	return &Debugger{
		config: config,
		spans:  make(map[string]MemorySpan, 2048),
	}
}

// SpanSyncer will return a span syncer that prints received spans.
func (d *Debugger) SpanSyncer() trace.SpanSyncer {
	return SpanSyncer(func(data *trace.SpanData) {
		// acquire mutex
		d.mutex.Lock()
		defer d.mutex.Unlock()

		// convert span
		span := traceSpanDataToMemorySpan(data)

		// store span if not root
		if span.Parent != "" {
			d.spans[span.ID] = span
			return
		}

		// collect spans
		table := make(map[string]MemorySpan)
		list := make([]MemorySpan, 0, 512)
		for id, s := range d.spans {
			if s.Trace == span.Trace {
				list = append(list, s)
				delete(d.spans, id)
				table[s.ID] = s
			}
		}

		// add root
		list = append(list, span)

		// adjust duration
		for i, span := range list {
			span.Start = span.Start.Round(d.config.TraceResolution)
			span.End = span.End.Round(d.config.TraceResolution)
			span.Duration = span.End.Sub(span.Start)
			list[i] = span
		}

		// build traces
		roots := buildTraces(list)

		// prepare buffer
		var buf bytes.Buffer

		// print header
		_, _ = fmt.Fprintf(&buf, "----- TRACE -----\n")

		// print roots
		for _, root := range roots {
			walkTrace(root, func(node *MemoryNode) bool {
				// prepare prefix
				prefix := strings.Repeat(" ", node.Depth*2)

				// prepare bar
				bar := buildBar(node.Span.Start.Sub(root.Span.Start), node.Span.Duration, root.Span.End.Sub(node.Span.End), 80)

				// print span
				_, _ = fmt.Fprintf(&buf, "%s  %s (%s)\n", bar, prefix+node.Span.Name, node.Span.Duration.String())

				return true
			})
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
			if event.Exception[i].Stacktrace != nil {
				st := event.Exception[i].Stacktrace
				for i, j := 0, len(st.Frames)-1; i < j; i, j = i+1, j-1 {
					st.Frames[i], st.Frames[j] = st.Frames[j], st.Frames[i]
				}
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
			if exc.Stacktrace != nil {
				for _, frame := range exc.Stacktrace.Frames {
					_, _ = fmt.Fprintf(&buf, "  > %s (%s): %s:%d\n", frame.Function, frame.Module, frame.AbsPath, frame.Lineno)
				}
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
