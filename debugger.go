package xo

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

type Debugger struct {
	config Config
	spans  map[string]MemorySpan
	mutex  sync.Mutex
}

func NewDebugger(config Config) *Debugger {
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
		span := convertSpan(data)

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

		// adjust times and durations
		for i, span := range list {
			span.Start = span.Start.Round(d.config.TraceResolution)
			span.End = span.End.Round(d.config.TraceResolution)
			span.Duration = span.End.Sub(span.Start)
			for j, event := range span.Events {
				event.Time = event.Time.Round(d.config.TraceResolution)
				span.Events[j] = event
			}
			list[i] = span
		}

		// build traces
		roots := buildTraces(list)

		// calculate longest tag
		var longest int
		for _, root := range roots {
			walkTrace(root, func(node *MemoryNode) bool {
				// check span name
				length := node.Depth*2 + len(node.Span.Name)
				if length > longest {
					longest = length
				}

				// check event names
				for _, event := range node.Span.Events {
					length := node.Depth*2 + 1 + len(event.Name)
					if length > longest {
						longest = length
					}
				}

				return true
			})
		}

		// prepare buffer
		var buf bytes.Buffer

		// prepare format
		format := fmt.Sprintf("%%-%ds   %%s   %%-6s  %%s", longest)

		// print roots
		for _, root := range roots {
			walkTrace(root, func(node *MemoryNode) bool {
				// prepare name
				name := strings.Repeat(" ", node.Depth*2) + node.Span.Name

				// prepare bar
				bar := buildBar(node.Span.Start.Sub(root.Span.Start), node.Span.Duration, root.Span.End.Sub(node.Span.End), d.config.TraceWidth)

				// prepare duration
				duration := autoTruncate(node.Span.Duration, 3)

				// prepare attributes
				var attributes string
				if !d.config.NoTraceAttributes {
					attributes = buildMap(node.Span.Attributes)
				}

				// build span
				str := strings.TrimRightFunc(fmt.Sprintf(format, name, bar, duration.String(), attributes), unicode.IsSpace)

				// print span
				_, _ = buf.WriteString(str)
				_, _ = buf.WriteRune('\n')

				// print events
				for _, event := range node.Span.Events {
					// prepare name
					prefix := strings.Repeat(" ", node.Depth*2)
					name := fmt.Sprintf("%s:%s", prefix, event.Name)

					// prepare dot
					dot := buildDot(event.Time.Sub(root.Span.Start), root.Span.End.Sub(event.Time), 80)

					// prepare timing
					timing := autoTruncate(event.Time.Sub(root.Span.Start), 3)

					// prepare attributes
					var attributes string
					if !d.config.NoTraceAttributes {
						attributes = buildMap(event.Attributes)
					}

					// build span
					str := strings.TrimRightFunc(fmt.Sprintf(format, name, dot, timing.String(), attributes), unicode.IsSpace)

					// print span
					_, _ = buf.WriteString(str)
					_, _ = buf.WriteRune('\n')
				}

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

		// print info
		_, _ = fmt.Fprintf(&buf, "Level: %s\n", event.Level)

		// print context
		if !d.config.NoEventContext {
			_, _ = fmt.Fprintf(&buf, "Context:\n")
			iterateMap(event.Contexts, func(key string, value interface{}) {
				_, _ = fmt.Fprintf(&buf, "- %s: %v\n", key, convertValue(value))
			})
		}

		// print exceptions
		_, _ = fmt.Fprintf(&buf, "Exceptions:\n")
		for _, exc := range event.Exception {
			_, _ = fmt.Fprintf(&buf, "- %s (%s)\n", exc.Value, exc.Type)
			if exc.Stacktrace != nil {
				for _, frame := range exc.Stacktrace.Frames {
					var line = ""
					if !d.config.NoEventLineNumbers {
						line = ":" + strconv.Itoa(frame.Lineno)
					}
					_, _ = fmt.Fprintf(&buf, "  > %s (%s): %s%s\n", frame.Function, frame.Module, frame.AbsPath, line)
				}
			}
		}

		// write event
		_, _ = buf.WriteTo(d.config.EventOutput)
	})
}
