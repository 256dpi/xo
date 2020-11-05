package xo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCaller(t *testing.T) {
	func() {
		caller := GetCaller(0, 0)
		assert.Len(t, caller.Stack, 4)
		assert.Equal(t, Caller{
			Short: "xo.TestGetCaller.func1",
			Full:  "github.com/256dpi/xo.TestGetCaller.func1",
			File:  "github.com/256dpi/xo/caller_test.go",
			Line:  12,
			Stack: caller.Stack,
		}, caller)

		caller = GetCaller(1, 2)
		assert.Len(t, caller.Stack, 2)
		assert.Equal(t, Caller{
			Short: "xo.TestGetCaller",
			Full:  "github.com/256dpi/xo.TestGetCaller",
			File:  "github.com/256dpi/xo/caller_test.go",
			Line:  31,
			Stack: caller.Stack,
		}, caller)
	}()
}

func TestCallerIncludes(t *testing.T) {
	parent := GetCaller(1, 0)
	child := GetCaller(0, 0)

	assert.False(t, parent.Includes(child, 0))
	assert.True(t, parent.Includes(parent, 0))
	assert.True(t, child.Includes(child, 0))
	assert.True(t, child.Includes(parent, 0))

	parent = GetCaller(0, 0)
	child = func() Caller {
		return GetCaller(0, 0)
	}()

	assert.False(t, parent.Includes(child, 0))
	assert.True(t, parent.Includes(parent, 0))
	assert.True(t, child.Includes(child, 0))
	assert.False(t, child.Includes(parent, 0))
	assert.True(t, child.Includes(parent, 1))
}

func TestCallerFormat(t *testing.T) {
	caller := GetCaller(0, 0)

	str := caller.String()
	assert.Equal(t, "xo.TestCallerFormat", str)

	str = fmt.Sprintf("%s", caller)
	assert.Equal(t, "xo.TestCallerFormat", str)

	str = fmt.Sprintf("%q", caller)
	assert.Equal(t, `"xo.TestCallerFormat"`, str)

	str = fmt.Sprintf("%v", caller)
	assert.Equal(t, "github.com/256dpi/xo.TestCallerFormat", str)

	str = fmt.Sprintf("%+v", caller)
	assert.Equal(t, []string{
		"> github.com/256dpi/xo.TestCallerFormat",
		">   github.com/256dpi/xo/caller_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))
}

func BenchmarkGetCaller(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetCaller(0, 0)
	}
}
