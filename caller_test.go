package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaller(t *testing.T) {
	func() {
		caller := Caller(0)
		assert.Len(t, caller.Stack, 4)
		assert.Equal(t, CallerInfo{
			Full:  "github.com/256dpi/xo.TestCaller.func1",
			Short: "xo.TestCaller.func1",
			File:  "/Users/256dpi/Development/GitHub/256dpi/xo/caller_test.go",
			Line:  11,
			Stack: caller.Stack,
		}, caller)

		caller = Caller(1)
		assert.Len(t, caller.Stack, 3)
		assert.Equal(t, CallerInfo{
			Full:  "github.com/256dpi/xo.TestCaller",
			Short: "xo.TestCaller",
			File:  "/Users/256dpi/Development/GitHub/256dpi/xo/caller_test.go",
			Line:  30,
			Stack: caller.Stack,
		}, caller)
	}()
}

func BenchmarkCaller(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Caller(0)
	}
}
