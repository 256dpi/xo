package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartSpan(t *testing.T) {
	ctx, span := StartSpan(nil, "foo")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	assert.Equal(t, span, GetSpan(ctx))
}

func TestGetSpan(t *testing.T) {
	span := GetSpan(nil)
	assert.Nil(t, span)

	ctx, root := StartSpan(nil, "root")
	span = GetSpan(ctx)
	assert.Equal(t, root, span)

	ctx, sub := StartSpan(ctx, "sub")
	span = GetSpan(ctx)
	assert.Equal(t, sub, span)
	assert.Equal(t, root.SpanContext().TraceID, sub.SpanContext().TraceID)
}
