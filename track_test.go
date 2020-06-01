package xo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

		assert.Equal(t, []MockSpan{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "track"},
			{Name: "root"},
		}, mock.Spans)
	})
}
