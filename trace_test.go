package xo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmartTrace(t *testing.T) {
	Test(func(tester *Tester) {
		ctx, span := SmartTrace(nil)
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		assert.Equal(t, []VSpan{
			{Name: "xo.TestSmartTrace.func1"},
		}, tester.ReducedSpans(0))
	})
}

func TestTrace(t *testing.T) {
	Test(func(tester *Tester) {
		ctx, span := Trace(nil, "foo")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		ctx, span = Trace(context.Background(), "bar")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		ctx, root := StartSpan(context.Background(), "root")

		ctx, span = Trace(ctx, "trace")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()

		root.End()

		assert.Equal(t, []VSpan{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "trace"},
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

func TestTraceMeta(t *testing.T) {
	Test(func(tester *Tester) {
		_, span := Trace(nil, "foo")
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
						Name: "exception",
						Attributes: M{
							"exception.message": "some error",
							"exception.type":    "*xo.Err",
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

func BenchmarkTraceRoot(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := Trace(nil, "foo")
		span.Tag("a", 1)
		span.End()
	}
}

func BenchmarkTraceRootParallel(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, span := Trace(nil, "foo")
			span.Tag("a", 1)
			span.End()
		}
	})
}

func BenchmarkTraceChild(b *testing.B) {
	b.ReportAllocs()

	ctx, _ := Trace(nil, "foo")

	for i := 0; i < b.N; i++ {
		_, span := Trace(ctx, "bar")
		span.Tag("a", 1)
		span.End()
	}
}

func BenchmarkTraceChildParallel(b *testing.B) {
	b.ReportAllocs()

	ctx, _ := Trace(nil, "foo")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, span := Trace(ctx, "bar")
			span.Tag("a", 1)
			span.End()
		}
	})
}
