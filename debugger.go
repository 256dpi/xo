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

// DebugConfig is used to configure xo for debugging.
type DebugConfig struct {
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

	// Whether to include trace attributes.
	TraceAttributes bool

	// The output for reports.
	//
	// Default: Sink("REPORT").
	ReportOutput io.Writer

	// Whether to omit report context data.
	NoReportContext bool

	// Whether to omit file paths from report stack traces.
	NoReportPaths bool

	// Whether to omit line numbers from report stack traces.
	NoReportLineNumbers bool
}

// Ensure will ensure defaults.
func (c *DebugConfig) Ensure() {
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

	// set default report output
	if c.ReportOutput == nil {
		c.ReportOutput = Sink("REPORT")
	}
}

// Debug will install logging, reporting and tracing components for debugging
// purposes. The returned function may be called to teardown all installed
// components.
func Debug(config DebugConfig) func() {
	// intercept
	var undoIntercept func()
	if !config.NoIntercept {
		undoIntercept = Intercept()
	}

	// create debugger
	debugger := NewDebugger(config)

	// hook tracing
	revertTracing := HookTracing(debugger.SpanExporter())

	// hook reporting
	revertReporting := HookReporting(debugger.SentryTransport())

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
	config DebugConfig
	mutex  sync.Mutex
}

// NewDebugger will create and return a new debugger.
func NewDebugger(config DebugConfig) *Debugger {
	// ensure config
	config.Ensure()

	return &Debugger{
		config: config,
	}
}

// SpanExporter will return a span exporter that prints received spans.
func (d *Debugger) SpanExporter() trace.SpanExporter {
	// prepare spans
	spans := map[string]VSpan{}

	return SpanExporter(func(data *trace.SpanData) error {
		// acquire mutex
		d.mutex.Lock()
		defer d.mutex.Unlock()

		// convert span
		span := ConvertSpan(data)

		// store span if not root
		if span.Parent != "" {
			spans[span.ID] = span
			return nil
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
				length := 2 + node.Depth*2 + len(node.Span.Name)
				if length > longest {
					longest = length
				}

				// check event names
				for _, event := range node.Span.Events {
					length := 2 + node.Depth*2 + 1 + len(event.Name)
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
				// prepare prefix
				prefix := strings.Repeat(" ", node.Depth*2)
				if node.Depth == 0 {
					prefix = "> " + prefix
				} else if node.Depth > 0 {
					prefix = "| " + prefix
				}

				// prepare name
				name := prefix + node.Span.Name

				// prepare bar
				bar := buildBar(node.Span.Start.Sub(root.Span.Start), node.Span.Duration, root.Span.End.Sub(node.Span.End), d.config.TraceWidth)

				// rescale duration
				duration := rescale(node.Span.Duration, 3)

				// prepare attributes
				var attributes string
				if d.config.TraceAttributes {
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
					name := fmt.Sprintf("%s:%s", prefix, event.Name)

					// prepare dot
					dot := buildDot(event.Time.Sub(root.Span.Start), root.Span.End.Sub(event.Time), 80)

					// rescale timing
					timing := rescale(event.Time.Sub(root.Span.Start), 3)

					// prepare attributes
					var attributes string
					if d.config.TraceAttributes {
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

		return nil
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
		check(fmt.Fprintf(&buf, "%s\n", strings.ToUpper(report.Level)))

		// print context
		if !d.config.NoReportContext && len(report.Context) > 0 {
			iterateMap(report.Context, func(key string, value interface{}) {
				check(fmt.Fprintf(&buf, "• %s: %v\n", key, convertValue(value)))
			})
		}

		// print tags
		if len(report.Tags) > 0 {
			iterateMap(report.Tags, func(key string, value interface{}) {
				check(fmt.Fprintf(&buf, "• %s: %v\n", key, convertValue(value)))
			})
		}

		// print exceptions
		for _, exc := range report.Exceptions {
			// print error
			check(fmt.Fprintf(&buf, "> %s (%s)\n", exc.Value, exc.Type))

			// print frames
			for _, frame := range exc.Frames {
				// check path
				if d.config.NoReportPaths {
					check(fmt.Fprintf(&buf, "|   %s (%s)\n", frame.Func, frame.Module))
					continue
				}

				// prepare line
				var line = ""
				if !d.config.NoReportLineNumbers {
					line = ":" + strconv.Itoa(frame.Line)
				}

				// print frame
				check(fmt.Fprintf(&buf, "|   %s (%s): %s%s\n", frame.Func, frame.Module, frame.Path, line))
			}
		}

		// write event
		_, err := buf.WriteTo(d.config.ReportOutput)
		if err != nil {
			raise(err)
		}
	})
}
