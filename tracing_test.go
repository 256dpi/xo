package xo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartSpan(t *testing.T) {
	Test(func(tester *Tester) {
		ctx, span := StartSpan(nil, "foo")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		assert.Equal(t, span, GetSpan(ctx))

		span.End()
		assert.Equal(t, []VSpan{
			{Name: "foo"},
		}, tester.ReducedSpans(0))
	})
}

func TestGetSpan(t *testing.T) {
	Test(func(tester *Tester) {
		span := GetSpan(nil)
		assert.Nil(t, span)

		span = GetSpan(context.Background())
		assert.Nil(t, span)

		ctx, root := StartSpan(nil, "root")
		span = GetSpan(ctx)
		assert.Equal(t, root, span)

		ctx, sub := StartSpan(ctx, "sub")
		span = GetSpan(ctx)
		assert.Equal(t, sub, span)
		assert.Equal(t, root.SpanContext().TraceID, sub.SpanContext().TraceID)

		sub.End()
		root.End()
		assert.Equal(t, []VSpan{
			{Name: "sub"},
			{Name: "root"},
		}, tester.ReducedSpans(0))
	})
}
