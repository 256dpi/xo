package xo

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

// Config is used to configure xo.
type Config struct {
	// Whether to omit interception.
	//
	// Default: false.
	NoIntercept bool

	// The output for traces.
	//
	// Default: Sink("TRACE").
	TraceOutput io.Writer

	// The trace resolution.
	//
	// Default: 1ns.
	TraceResolution time.Duration

	// The trace width.
	//
	// Default: 80.
	TraceWidth int

	// Whether to omit trace attributes.
	NoTraceAttributes bool

	// The output for events.
	//
	// Default: Sink("EVENT").
	EventOutput io.Writer

	// Whether to omit event context data.
	NoEventContext bool

	// Whether to omit file paths from event stack traces.
	NoEventPaths bool

	// Whether to omit line numbers from event stack traces.
	NoEventLineNumbers bool
}

// Ensure will ensure defaults.
func (c *Config) Ensure() {
	// set default trace output
	if c.TraceOutput == nil {
		c.TraceOutput = Sink("TRACE")
	}

	// set default trace resolution
	if c.TraceResolution == 0 {
		c.TraceResolution = time.Nanosecond
	}

	// set default trace width
	if c.TraceWidth == 0 {
		c.TraceWidth = 80
	}

	// set default event output
	if c.EventOutput == nil {
		c.EventOutput = Sink("EVENT")
	}
}

// Debug will install logging, reporting and tracing components for debugging
// purposes. The returned function may be called to teardown all installed
// components.
func Debug(config Config) func() {
	// intercept
	var undoIntercept func()
	if !config.NoIntercept {
		undoIntercept = Intercept()
	}

	// create debugger
	debugger := NewDebugger(config)

	// setup tracing
	revertTracing := SetupTracing(debugger.SpanSyncer())

	// setup reporting
	revertReporting := SetupReporting(debugger.SentryTransport())

	return func() {
		// revert reporting
		revertReporting()

		// revert tracing
		revertTracing()

		// reset intercept
		if undoIntercept != nil {
			undoIntercept()
		}
	}
}

// Debugger is a virtual logging, tracing and reporting provider for debugging
// purposes.
type Debugger struct {
	config Config
	mutex  sync.Mutex
}

// NewDebugger will create and return a new debugger.
func NewDebugger(config Config) *Debugger {
	// ensure config
	config.Ensure()

	return &Debugger{
		config: config,
	}
}

// SpanSyncer will return a span syncer that prints received spans.
func (d *Debugger) SpanSyncer() trace.SpanSyncer {
	// prepare spans
	spans := map[string]VSpan{}

	return SpanSyncer(func(data *trace.SpanData) {
		// acquire mutex
		d.mutex.Lock()
		defer d.mutex.Unlock()

		// convert span
		span := ConvertSpan(data)

		// store span if not root
		if span.Parent != "" {
			spans[span.ID] = span
			return
		}

		// collect spans
		table := make(map[string]VSpan)
		list := make([]VSpan, 0, 512)
		for id, s := range spans {
			if s.Trace == span.Trace {
				list = append(list, s)
				delete(spans, id)
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
		roots := BuildTraces(list)

		// calculate longest tag
		var longest int
		for _, root := range roots {
			WalkTrace(root, func(node *VNode) bool {
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
			WalkTrace(root, func(node *VNode) bool {
				// prepare name
				name := strings.Repeat(" ", node.Depth*2) + node.Span.Name

				// prepare bar
				bar := buildBar(node.Span.Start.Sub(root.Span.Start), node.Span.Duration, root.Span.End.Sub(node.Span.End), d.config.TraceWidth)

				// rescale duration
				duration := rescale(node.Span.Duration, 3)

				// prepare attributes
				var attributes string
				if !d.config.NoTraceAttributes {
					attributes = buildMeta(node.Span.Attributes)
				}

				// build span
				str := strings.TrimRightFunc(fmt.Sprintf(format, name, bar, duration.String(), attributes), unicode.IsSpace)

				// print span
				check(buf.WriteString(str))
				check(buf.WriteRune('\n'))

				// print events
				for _, event := range node.Span.Events {
					// prepare name
					prefix := strings.Repeat(" ", node.Depth*2)
					name := fmt.Sprintf("%s:%s", prefix, event.Name)

					// prepare dot
					dot := buildDot(event.Time.Sub(root.Span.Start), root.Span.End.Sub(event.Time), 80)

					// rescale timing
					timing := rescale(event.Time.Sub(root.Span.Start), 3)

					// prepare attributes
					var attributes string
					if !d.config.NoTraceAttributes {
						attributes = buildMeta(event.Attributes)
					}

					// build span
					str := strings.TrimRightFunc(fmt.Sprintf(format, name, dot, timing.String(), attributes), unicode.IsSpace)

					// print span
					check(buf.WriteString(str))
					check(buf.WriteRune('\n'))
				}

				return true
			})
		}

		// write trace
		_, err := buf.WriteTo(d.config.TraceOutput)
		if err != nil {
			raise(err)
		}
	})
}

// SentryTransport will return a sentry transport that print received events.
func (d *Debugger) SentryTransport() sentry.Transport {
	return SentryTransport(func(event *sentry.Event) {
		// convert report
		report := ConvertReport(event)

		// reverse stack traces
		for _, exc := range report.Exceptions {
			for i, j := 0, len(exc.Frames)-1; i < j; i, j = i+1, j-1 {
				exc.Frames[i], exc.Frames[j] = exc.Frames[j], exc.Frames[i]
			}
		}

		// prepare buffer
		var buf bytes.Buffer

		// print info
		check(fmt.Fprintf(&buf, "Level: %s\n", report.Level))

		// print context
		if !d.config.NoEventContext {
			check(fmt.Fprintf(&buf, "Context:\n"))
			iterateMap(report.Context, func(key string, value interface{}) {
				check(fmt.Fprintf(&buf, "- %s: %v\n", key, convertValue(value)))
			})
		}

		// print exceptions
		check(fmt.Fprintf(&buf, "Exceptions:\n"))
		for _, exc := range report.Exceptions {
			// print error
			check(fmt.Fprintf(&buf, "- %s (%s)\n", exc.Value, exc.Type))

			// print frames
			for _, frame := range exc.Frames {
				// check path
				if d.config.NoEventPaths {
					check(fmt.Fprintf(&buf, "  > %s (%s)\n", frame.Func, frame.Module))
					continue
				}

				// prepare line
				var line = ""
				if !d.config.NoEventLineNumbers {
					line = ":" + strconv.Itoa(frame.Line)
				}

				// print frame
				check(fmt.Fprintf(&buf, "  > %s (%s): %s%s\n", frame.Func, frame.Module, frame.Path, line))
			}
		}

		// write event
		_, err := buf.WriteTo(d.config.EventOutput)
		if err != nil {
			raise(err)
		}
	})
}
