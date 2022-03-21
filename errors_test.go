package xo

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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
		"> github.com/256dpi/xo.TestF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
}

func TestW(t *testing.T) {
	err := W(nil)
	assert.NoError(t, err)

	err = W(errors.New("foo"))
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
		"> github.com/256dpi/xo.TestW",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))

	err = W(func() error {
		return W(func() error {
			return W(errors.New("foo"))
		}())
	}())
	assert.Error(t, err)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestW.func1.1",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> github.com/256dpi/xo.TestW.func1",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> github.com/256dpi/xo.TestW",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
}

func TestWS(t *testing.T) {
	func() {
		err := WS(nil, 1)
		assert.NoError(t, err)

		err = WS(errors.New("foo"), 1)
		assert.Error(t, err)

		str := err.Error()
		assert.Equal(t, "foo", str)

		str = fmt.Sprintf("%s", err)
		assert.Equal(t, "foo", str)

		str = fmt.Sprintf("%q", err)
		assert.Equal(t, `"foo"`, str)

		str = fmt.Sprintf("%v", err)
		assert.Equal(t, "xo.TestWS: foo", str)

		str = fmt.Sprintf("%+v", err)
		assert.Equal(t, []string{
			"foo",
			"> github.com/256dpi/xo.TestWS",
			">   github.com/256dpi/xo/errors_test.go:LN",
			"> testing.tRunner",
			">   testing/testing.go:LN",
			"> runtime.goexit",
			">   runtime/asm_" + arch + ".s:LN",
		}, splitStackTrace(str))

		err = W(func() error {
			return W(func() error {
				return WS(errors.New("foo"), 1)
			}())
		}())
		assert.Error(t, err)

		str = fmt.Sprintf("%+v", err)
		assert.Equal(t, []string{
			"foo",
			"> github.com/256dpi/xo.TestWS.func1.1",
			">   github.com/256dpi/xo/errors_test.go:LN",
			"> github.com/256dpi/xo.TestWS.func1",
			">   github.com/256dpi/xo/errors_test.go:LN",
			"> github.com/256dpi/xo.TestWS",
			">   github.com/256dpi/xo/errors_test.go:LN",
			"> testing.tRunner",
			">   testing/testing.go:LN",
			"> runtime.goexit",
			">   runtime/asm_" + arch + ".s:LN",
		}, splitStackTrace(str))
	}()
}

func TestWF(t *testing.T) {
	err := WF(nil, "foo")
	assert.NoError(t, err)

	err = F("foo")
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
		"> github.com/256dpi/xo.TestWF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
		"bar 42",
		"> github.com/256dpi/xo.TestWF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
}

func TestDrop(t *testing.T) {
	err := Drop(W(nil), 1)
	assert.NoError(t, err)

	err = Drop(W(errors.New("foo")), 1)
	assert.Error(t, err)

	str := err.Error()
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%q", err)
	assert.Equal(t, `"foo"`, str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "testing.tRunner: foo", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))

	err = W(func() error {
		return W(func() error {
			return Drop(W(errors.New("foo")), 1)
		}())
	}())
	assert.Error(t, err)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestDrop.func1",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> github.com/256dpi/xo.TestDrop",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
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
		"> github.com/256dpi/xo.TestSF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))

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
		"> github.com/256dpi/xo.TestSF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
		"bar",
		"> github.com/256dpi/xo.TestSF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
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

var baseFoo = BF("foo")

func TestBFWrap(t *testing.T) {
	err := baseFoo.Wrap()
	assert.Error(t, err)
	assert.True(t, baseFoo.Is(err))
	assert.True(t, baseFoo.Is(baseFoo.Self()))
	assert.NotEqual(t, err, baseFoo.Self())

	str := err.Error()
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.TestBFWrap: foo", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.init",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> github.com/256dpi/xo.TestBFWrap",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
}

var baseBar = BW(errors.New("bar"))

func TestBWWrapF(t *testing.T) {
	err := baseBar.WrapF("baz")
	assert.Error(t, err)
	assert.True(t, baseBar.Is(err))
	assert.True(t, baseBar.Is(baseBar.Self()))
	assert.NotEqual(t, err, baseBar.Self())

	str := err.Error()
	assert.Equal(t, "baz: bar", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "baz: bar", str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.TestBWWrapF: baz: bar", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"bar",
		"> github.com/256dpi/xo.init",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"baz",
		"> github.com/256dpi/xo.TestBWWrapF",
		">   github.com/256dpi/xo/errors_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_" + arch + ".s:LN",
	}, splitStackTrace(str))
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

func splitStackTrace(str string) []string {
	str = strings.ReplaceAll(str, "\t", "  ")
	str = regexp.MustCompile(":\\d+").ReplaceAllString(str, ":LN")
	return strings.Split(str, "\n")
}
