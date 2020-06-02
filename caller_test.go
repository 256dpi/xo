package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCaller(t *testing.T) {
	func() {
		caller := GetCaller(0)
		assert.Len(t, caller.Stack, 4)
		assert.Equal(t, Caller{
			Short: "xo.TestGetCaller.func1",
			Full:  "github.com/256dpi/xo.TestGetCaller.func1",
			File:  "/Users/256dpi/Development/GitHub/256dpi/xo/caller_test.go",
			Line:  11,
			Stack: caller.Stack,
		}, caller)

		caller = GetCaller(1)
		assert.Len(t, caller.Stack, 3)
		assert.Equal(t, Caller{
			Short: "xo.TestGetCaller",
			Full:  "github.com/256dpi/xo.TestGetCaller",
			File:  "/Users/256dpi/Development/GitHub/256dpi/xo/caller_test.go",
			Line:  30,
			Stack: caller.Stack,
		}, caller)
	}()
}

func BenchmarkGetCaller(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetCaller(0)
	}
}
