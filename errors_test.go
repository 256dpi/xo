package xo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestF(t *testing.T) {
	err := F("foo %d", 42)
	assert.Error(t, err)

	str := err.Error()
	assert.Equal(t, "foo 42", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "foo 42", str)

	str = fmt.Sprintf("%q", err)
	assert.Equal(t, `"foo 42"`, str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.TestF: foo 42", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo 42",
		"github.com/256dpi/xo.TestF",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:11",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
	}, splitTrace(str))
}

func TestW(t *testing.T) {
	err := F("foo")
	err = W(err)
	assert.Error(t, err)

	str := err.Error()
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%q", err)
	assert.Equal(t, `"foo"`, str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.TestW: foo", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"github.com/256dpi/xo.TestW",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:39",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
	}, splitTrace(str))
}

func TestSuperfluousWrap(t *testing.T) {
	err := W(func() error {
		return F("foo")
	}())
	assert.Error(t, err)

	str := fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"github.com/256dpi/xo.TestSuperfluousWrap.func1",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:69",
		"github.com/256dpi/xo.TestSuperfluousWrap",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:70",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
	}, splitTrace(str))
}

func TestWF(t *testing.T) {
	err := F("foo")
	err = WF(err, "bar %d", 42)
	assert.Error(t, err)

	str := err.Error()
	assert.Equal(t, "bar 42: foo", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "bar 42: foo", str)

	str = fmt.Sprintf("%q", err)
	assert.Equal(t, `"bar 42: foo"`, str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.TestWF: bar 42: foo", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"github.com/256dpi/xo.TestWF",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:88",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
		"bar 42",
		"github.com/256dpi/xo.TestWF",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:89",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
	}, splitTrace(str))
}

func TestSF(t *testing.T) {
	err := SF("foo")
	assert.True(t, IsSafe(err))

	str := err.Error()
	assert.Equal(t, `foo`, str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, `foo`, str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, `xo.TestSF: foo`, str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"github.com/256dpi/xo.TestSF",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:124",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
	}, splitTrace(str))

	/* wrapped */

	err = WF(err, "bar")
	assert.True(t, IsSafe(err))

	str = err.Error()
	assert.Equal(t, "bar: foo", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "bar: foo", str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.TestSF: bar: foo", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"github.com/256dpi/xo.TestSF",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:124",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
		"bar",
		"github.com/256dpi/xo.TestSF",
		"  /Users/256dpi/Development/GitHub/256dpi/xo/errors_test.go:149",
		"testing.tRunner",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992",
		"runtime.goexit",
		"  /usr/local/Cellar/go/1.14.1/libexec/src/runtime/asm_amd64.s:1373",
	}, splitTrace(str))
}

func TestSafeErr(t *testing.T) {
	err1 := F("foo")
	assert.False(t, IsSafe(err1))
	assert.Equal(t, "foo", err1.Error())
	assert.Nil(t, AsSafe(err1))

	err2 := SW(err1)
	assert.True(t, IsSafe(err2))
	assert.Equal(t, "foo", err2.Error())
	assert.Equal(t, err2, AsSafe(err2))

	err3 := WF(err2, "bar")
	assert.True(t, IsSafe(err3))
	assert.Equal(t, "bar: foo", err3.Error())
	assert.Equal(t, err2, AsSafe(err3))
}

func BenchmarkF(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = F("foo: %d", 42)
	}
}

func BenchmarkW(b *testing.B) {
	err := F("foo")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = W(err)
	}
}

func BenchmarkWF(b *testing.B) {
	err := F("foo")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = WF(err, "foo %d", 42)
	}
}
