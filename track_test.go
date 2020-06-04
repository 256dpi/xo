package xo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoTrack(t *testing.T) {
	Test(func(tester *Tester) {
		ctx, span := AutoTrack(nil)
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		assert.Equal(t, []VSpan{
			{Name: "xo.TestAutoTrack.func1"},
		}, tester.ReducedSpans(0))
	})
}

func TestTrack(t *testing.T) {
	Test(func(tester *Tester) {
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

		assert.Equal(t, []VSpan{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "track"},
			{Name: "root"},
		}, tester.ReducedSpans(0))
	})
}

func TestNewSpan(t *testing.T) {
	ctx, span := StartSpan(nil, "foo")
	newSpan := NewSpan(ctx, span)
	assert.Equal(t, ctx, newSpan.Context())
	assert.Equal(t, span, newSpan.Native())
}

func TestTrackMeta(t *testing.T) {
	Test(func(tester *Tester) {
		_, span := Track(nil, "foo")
		span.Tag("foo", "bar")
		span.Tag("rich", M{"foo": "bar"})
		span.Attach("foo", M{"bar": "baz"})
		span.Record(F("some error"))
		span.Log("some message: %d", 42)
		span.End()

		assert.Equal(t, []VSpan{
			{
				Name: "foo",
				Attributes: M{
					"foo":  "bar",
					"rich": `{"foo":"bar"}`,
				},
				Events: []VEvent{
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
		}, tester.ReducedSpans(0))
	})
}
