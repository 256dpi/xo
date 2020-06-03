package xo

import (
	"sort"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

type MemorySpanEvent struct {
	// The event name.
	Name string

	// The event time.
	Time time.Time

	// Attached event attributes.
	Attributes M
}

type MemorySpan struct {
	// The span id.
	ID string

	// The span trace id.
	Trace string

	// The parent span id.
	Parent string

	// The span name.
	Name string

	// The span start and end time.
	Start time.Time
	End   time.Time

	// The span duration.
	Duration time.Duration

	// The span attributes.
	Attributes M

	// Attached span events.
	Events []MemorySpanEvent
}

type MemoryNode struct {
	// The nodes span.
	Span MemorySpan

	// The nodes parent, if any.
	Parent *MemoryNode

	// The node parent, if any.
	Children []*MemoryNode

	// The nodes depth.
	Depth int
}

type MemoryFrame struct {
	Func   string
	Module string
	File   string
	Path   string
	Line   int
}

type MemoryException struct {
	Type   string
	Value  string
	Module string
	Frames []MemoryFrame
}

type MemoryReport struct {
	ID         string
	Level      string
	Time       time.Time
	Context    M
	Tags       SM
	Exceptions []MemoryException
}

func convertReport(event *sentry.Event) MemoryReport {
	// prepare report
	report := MemoryReport{
		ID:    string(event.EventID),
		Level: string(event.Level),
		Time:  event.Timestamp,
	}

	// add context
	if len(event.Contexts) > 0 {
		report.Context = event.Contexts
	}

	// add tags
	if len(event.Tags) > 0 {
		report.Tags = event.Tags
	}

	// add exceptions
	for _, exc := range event.Exception {
		// prepare exception
		exception := MemoryException{
			Type:   exc.Type,
			Value:  exc.Value,
			Module: exc.Module,
		}

		// add frames
		if exc.Stacktrace != nil {
			for _, frame := range exc.Stacktrace.Frames {
				exception.Frames = append(exception.Frames, MemoryFrame{
					Func:   frame.Function,
					Module: frame.Module,
					File:   frame.Filename,
					Path:   frame.AbsPath,
					Line:   frame.Lineno,
				})
			}
		}

		// add exception
		report.Exceptions = append(report.Exceptions, exception)
	}

	return report
}

func buildTraces(list []MemorySpan) []*MemoryNode {
	// prepare nodes
	var roots []*MemoryNode
	nodes := map[string]*MemoryNode{}
	for _, span := range list {
		// create node
		node := &MemoryNode{
			Span: span,
		}

		// add root if no parent
		if span.Parent == "" {
			roots = append(roots, node)
		}

		// add node
		nodes[span.ID] = node
	}

	// link nodes
	for _, node := range nodes {
		if node.Span.Parent != "" {
			parent := nodes[node.Span.Parent]
			if parent != nil {
				node.Parent = parent
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// sort traces
	sortNodes(roots)

	// set depth
	for _, node := range nodes {
		depth := &node.Depth
		for node.Parent != nil {
			node = node.Parent
			*depth++
		}
	}

	return roots
}

func walkTrace(node *MemoryNode, fn func(node *MemoryNode) bool) bool {
	// yield node
	if !fn(node) {
		return false
	}

	// yield children
	for _, child := range node.Children {
		if !walkTrace(child, fn) {
			return false
		}
	}

	return true
}

func sortNodes(nodes []*MemoryNode) {
	// sort children
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Span.Start.Before(nodes[j].Span.Start)
	})

	// sort children
	for _, node := range nodes {
		sortNodes(node.Children)
	}
}

func convertSpan(data *trace.SpanData) MemorySpan {
	// collect events
	var events []MemorySpanEvent
	for _, event := range data.MessageEvents {
		events = append(events, MemorySpanEvent{
			Name:       event.Name,
			Time:       event.Time,
			Attributes: kvToMap(event.Attributes),
		})
	}

	// get parent
	parent := data.ParentSpanID.String()
	if !data.ParentSpanID.IsValid() || data.HasRemoteParent {
		parent = ""
	}

	// add span
	return MemorySpan{
		ID:         data.SpanContext.SpanID.String(),
		Trace:      data.SpanContext.TraceID.String(),
		Parent:     parent,
		Name:       data.Name,
		Start:      data.StartTime,
		End:        data.EndTime,
		Duration:   data.EndTime.Sub(data.StartTime),
		Attributes: kvToMap(data.Attributes),
		Events:     events,
	}
}
