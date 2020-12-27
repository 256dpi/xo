package xo

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

// VEvent is a virtual span event.
type VEvent struct {
	Name       string
	Time       time.Time
	Attributes M
}

// VSpan is a virtual span.
type VSpan struct {
	ID         string
	Trace      string
	Parent     string
	Name       string
	Start      time.Time
	End        time.Time
	Duration   time.Duration
	Attributes M
	Events     []VEvent
}

// VNode is a virtual trace node.
type VNode struct {
	Span     VSpan
	Parent   *VNode
	Children []*VNode
	Depth    int
}

// VFrame is a virtual exception frame.
type VFrame struct {
	Func   string
	Module string
	File   string
	Path   string
	Line   int
}

// VException is a virtual report exception.
type VException struct {
	Type   string
	Value  string
	Module string
	Frames []VFrame
}

// VReport is a virtual report.
type VReport struct {
	ID         string
	Level      string
	Time       time.Time
	Context    M
	Tags       M
	Exceptions []VException
}

// ConvertSpan will convert a raw span to a virtual span.
func ConvertSpan(data *trace.SpanData) VSpan {
	// collect events
	var events []VEvent
	for _, event := range data.MessageEvents {
		events = append(events, VEvent{
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
	return VSpan{
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

// ConvertReport will convert a raw event to virtual report.
func ConvertReport(event *sentry.Event) VReport {
	// prepare report
	report := VReport{
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
		report.Tags = map[string]interface{}{}
		for key, value := range event.Tags {
			report.Tags[key] = value
		}
	}

	// add exceptions
	for _, exc := range event.Exception {
		// prepare exception
		exception := VException{
			Type:   exc.Type,
			Value:  exc.Value,
			Module: exc.Module,
		}

		// add frames
		if exc.Stacktrace != nil {
			for _, frame := range exc.Stacktrace.Frames {
				// get file and path
				file := frame.Filename
				path := frame.AbsPath
				if file == "" && path != "" {
					file = filepath.Base(path)
				} else if file != "" && path == "" {
					path = file
					file = filepath.Base(file)
				} else {
					file = filepath.Base(file)
				}

				// add frame
				exception.Frames = append(exception.Frames, VFrame{
					Func:   frame.Function,
					Module: frame.Module,
					File:   file,
					Path:   path,
					Line:   frame.Lineno,
				})
			}
		}

		// add exception
		report.Exceptions = append(report.Exceptions, exception)
	}

	return report
}

// BuildTraces will assemble traces from a list of spans.
func BuildTraces(list []VSpan) []*VNode {
	// prepare nodes
	var roots []*VNode
	nodes := map[string]*VNode{}
	for _, span := range list {
		// create node
		node := &VNode{
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
	SortNodes(roots)

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

// WalkTrace will walk the specified trace.
func WalkTrace(node *VNode, fn func(node *VNode) bool) bool {
	// yield node
	if !fn(node) {
		return false
	}

	// yield children
	for _, child := range node.Children {
		if !WalkTrace(child, fn) {
			return false
		}
	}

	return true
}

// SortNodes will sort the specified nodes.
func SortNodes(nodes []*VNode) {
	// sort children
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Span.Start.Before(nodes[j].Span.Start)
	})

	// sort children
	for _, node := range nodes {
		SortNodes(node.Children)
	}
}

// VSink provides a virtual string buffer.
type VSink struct {
	String string
}

// Write implements the io.Writer interface.
func (s *VSink) Write(p []byte) (n int, err error) {
	s.String += string(p)
	return len(p), nil
}

// Close implements the io.Closer interface.
func (s *VSink) Close() error {
	return nil
}
