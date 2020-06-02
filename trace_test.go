package xo

import (
	"testing"
	"time"

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

		time.Sleep(10 * time.Millisecond)

		trace.Pop()
		assert.NotEqual(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		time.Sleep(10 * time.Millisecond)

		trace.Pop()
		assert.Equal(t, trace.Root().Native(), GetSpan(ctx))
		assert.Equal(t, trace.Tail().Native(), GetSpan(ctx))

		time.Sleep(10 * time.Millisecond)

		trace.End()

		assert.Equal(t, []MemorySpan{
			{
				Name:     "bar",
				Duration: 10 * time.Millisecond,
				Events: []MemorySpanEvent{
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
				Name:     "foo",
				Duration: 20 * time.Millisecond,
				Attributes: M{
					"key": "7",
				},
			},
			{
				Name:     "trace",
				Duration: 30 * time.Millisecond,
				Events: []MemorySpanEvent{
					{
						Name: "log",
						Attributes: M{
							"message": "7",
						},
					},
				},
			},
		}, mock.ReducedSpans(10*time.Millisecond))
	})
}
