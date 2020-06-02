package xo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	err := Run(nil, func(ctx *Context) error {
		ctx.Tag("tag", 42)
		ctx.Log("log", "bar")
		return nil
	})
	assert.NoError(t, err)

	err = Run(nil, func(ctx *Context) error {
		ctx.Tag("tag", 42)
		ctx.Log("log", "bar")
		return F("error")
	})
	assert.Error(t, err)
	assert.Equal(t, "error", err.Error())

	err = Run(nil, func(ctx *Context) error {
		ctx.Tag("tag", 42)
		ctx.Log("log", "bar")
		panic("error")
	})
	assert.Error(t, err)
	assert.Equal(t, "PANIC: error", err.Error())
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
