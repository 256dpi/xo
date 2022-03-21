package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracer(t *testing.T) {
	Test(func(tester *Tester) {
		tracer, ctx := CreateTracer(nil, "tracer")
		tracer.Log("7")
		assert.NotNil(t, tracer)
		assert.Equal(t, tracer, GetTracer(ctx))
		assert.Equal(t, tracer.Root().Native(), GetSpan(ctx))
		assert.Equal(t, tracer.Tail().Native(), GetSpan(ctx))

		tracer.Push("foo")
		tracer.Tag("key", "7")
		assert.NotEqual(t, tracer.Root().Native(), GetSpan(ctx))
		assert.Equal(t, tracer.Tail().Native(), GetSpan(ctx))

		tracer.SmartPush()
		tracer.Record(F("fail"))
		assert.NotEqual(t, tracer.Root().Native(), GetSpan(ctx))
		assert.Equal(t, tracer.Tail().Native(), GetSpan(ctx))

		tracer.Pop()
		assert.NotEqual(t, tracer.Root().Native(), GetSpan(ctx))
		assert.Equal(t, tracer.Tail().Native(), GetSpan(ctx))

		tracer.Pop()
		assert.Equal(t, tracer.Root().Native(), GetSpan(ctx))
		assert.Equal(t, tracer.Tail().Native(), GetSpan(ctx))

		tracer.End()
		assert.Equal(t, []VSpan{
			{
				Name: "xo.TestTracer.func1",
				Events: []VEvent{
					{
						Name: "exception",
						Attributes: M{
							"exception.message": "fail",
							"exception.type":    "*xo.Err",
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
				Name: "tracer",
				Events: []VEvent{
					{
						Name: "log",
						Attributes: M{
							"message": "7",
						},
					},
				},
			},
		}, tester.ReducedSpans(0))
	})
}
