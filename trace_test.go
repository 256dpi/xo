package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrace(t *testing.T) {
	Trap(func(mock *Mock) {
		trace, ctx := CreateTrace(nil, "trace")
		trace.Log("7")
		assert.NotNil(t, trace)
		assert.Equal(t, trace, GetTrace(ctx))
		assert.Equal(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		trace.Push("foo")
		trace.Tag("key", "7")
		assert.NotEqual(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		trace.Push("bar")
		trace.Record(F("fail"))
		assert.NotEqual(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		trace.Pop()
		assert.NotEqual(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		trace.Pop()
		assert.Equal(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		trace.End()

		assert.Equal(t, []VSpan{
			{
				Name: "bar",
				Events: []VEvent{
					{
						Name: "error",
						Attributes: M{
							"error.message": "fail",
							"error.type":    "*xo.Err",
						},
					},
				},
			},
			{
				Name: "foo",
				Attributes: M{
					"key": "7",
				},
			},
			{
				Name: "trace",
				Events: []VEvent{
					{
						Name: "log",
						Attributes: M{
							"message": "7",
						},
					},
				},
			},
		}, mock.ReducedSpans(0))
	})
}
