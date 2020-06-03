package xo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	Trap(func(mock *Mock) {
		err := Run(nil, func(ctx *Context) error {
			ctx.Tag("tag", 42)
			ctx.Log("bar")
			return nil
		})
		assert.NoError(t, err)

		err = Run(nil, func(ctx *Context) error {
			ctx.Tag("tag", 42)
			ctx.Log("bar")
			return F("error")
		})
		assert.Error(t, err)
		assert.Equal(t, "error", err.Error())
		assert.Equal(t, "xo.TestRun.func1: error", fmt.Sprintf("%v", err))

		err = Run(nil, func(ctx *Context) error {
			ctx.Tag("tag", 42)
			ctx.Log("bar")
			panic("error")
		})
		assert.Error(t, err)
		assert.Equal(t, "PANIC: error", err.Error())
		assert.Equal(t, "xo.TestRun.func1: PANIC: error", fmt.Sprintf("%v", err))

		assert.Equal(t, []MemorySpan{
			{
				Name: "xo.TestRun.func1",
				Attributes: M{
					"tag": int64(42),
				},
				Events: []MemorySpanEvent{
					{
						Name: "log",
						Attributes: M{
							"message": "bar",
						},
					},
				},
			},
			{
				Name:       "xo.TestRun.func1",
				Attributes: M{"tag": int64(42)},
				Events: []MemorySpanEvent{
					{
						Name: "log",
						Attributes: M{
							"message": "bar",
						},
					},
					{
						Name: "error",
						Attributes: M{
							"error.message": "error",
							"error.type":    "*xo.Err",
						},
					},
				},
			},
			{
				Name: "xo.TestRun.func1",
				Attributes: M{
					"tag": int64(42),
				},
				Events: []MemorySpanEvent{
					{
						Name: "log",
						Attributes: M{
							"message": "bar",
						},
					},
					{
						Name: "error",
						Attributes: M{
							"error.message": "PANIC: error",
							"error.type":    "*xo.Err",
						},
					},
				},
			},
		}, mock.ReducedSpans(10*time.Millisecond))
	})
}

func BenchmarkRun(b *testing.B) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Run(ctx, func(ctx *Context) error {
			return nil
		})
	}
}
