package xo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoTrack(t *testing.T) {
	Trap(func(mock *Mock) {
		ctx, span := AutoTrack(nil)
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		assert.Equal(t, []MemorySpan{
			{Name: "xo.TestAutoTrack.func1"},
		}, mock.ReducedSpans(0))
	})
}

func TestTrack(t *testing.T) {
	Trap(func(mock *Mock) {
		ctx, span := Track(nil, "foo")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		ctx, span = Track(context.Background(), "bar")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		ctx, root := StartSpan(context.Background(), "root")

		ctx, span = Track(ctx, "track")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		root.End()

		assert.Equal(t, []MemorySpan{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "track"},
			{Name: "root"},
		}, mock.ReducedSpans(0))
	})
}

func TestTrackMeta(t *testing.T) {
	Trap(func(mock *Mock) {
		_, span := Track(nil, "foo")
		span.Tag("foo", "bar")
		span.Tag("rich", M{"foo": "bar"})
		span.Attach("foo", M{"bar": "baz"})
		span.Record(F("some error"))
		span.Log("some message: %d", 42)
		span.End()

		assert.Equal(t, []MemorySpan{
			{
				Name: "foo",
				Attributes: M{
					"foo":  "bar",
					"rich": `{"foo":"bar"}`,
				},
				Events: []MemorySpanEvent{
					{
						Name: "foo",
						Attributes: M{
							"bar": "baz",
						},
					},
					{
						Name: "error",
						Attributes: M{
							"error.message": "some error",
							"error.type":    "*xo.Err",
						},
					},
					{
						Name: "log",
						Attributes: M{
							"message": "some message: 42",
						},
					},
				},
			},
		}, mock.ReducedSpans(0))
	})
}
