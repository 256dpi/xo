package xo

import (
	"fmt"
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

func TestCallerFormat(t *testing.T) {
	caller := GetCaller(0)

	str := caller.String()
	assert.Equal(t, "xo.TestCallerFormat", str)

	str = fmt.Sprintf("%v", caller)
	assert.Equal(t, "github.com/256dpi/xo.TestCallerFormat", str)

	str = fmt.Sprintf("%+v", caller)
	assert.Equal(t, []string{
		"github.com/256dpi/xo.TestCallerFormat",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/caller_test.go:LN",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:LN",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:LN",
	}, splitTrace(str))
}

func BenchmarkGetCaller(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetCaller(0)
	}
}
