package xo

import (
	"sort"
	"time"

	"go.opentelemetry.io/otel/sdk/export/trace"
)

type MemorySpanEvent struct {
	// The event name.
	Name string

	// Attached event attributes.
	Attributes map[string]interface{}
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
	Attributes map[string]interface{}

	// Attached span events.
	Events []MemorySpanEvent
}

type MemoryNode struct {
	Span     MemorySpan
	Parent   *MemoryNode
	Children []*MemoryNode
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
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// sort traces
	sortNodes(roots)

	return roots
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

func traceSpanDataToMemorySpan(data *trace.SpanData) MemorySpan {
	// collect events
	var events []MemorySpanEvent
	for _, event := range data.MessageEvents {
		events = append(events, MemorySpanEvent{
			Name:       event.Name,
			Attributes: otelKVToMap(event.Attributes),
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
		Attributes: otelKVToMap(data.Attributes),
		Events:     events,
	}
}
