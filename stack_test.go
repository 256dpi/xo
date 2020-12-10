package xo

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errFoo = errors.New("foo")

func TestAbort(t *testing.T) {
	var res error
	func() {
		defer Resume(func(err error) {
			res = err
		})

		Abort(errFoo)
	}()

	assert.True(t, errors.Is(res, errFoo))

	str := res.Error()
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%s", res)
	assert.Equal(t, "foo", str)

	str = fmt.Sprintf("%q", res)
	assert.Equal(t, `"foo"`, str)

	str = fmt.Sprintf("%v", res)
	assert.Equal(t, "xo.TestAbort.func1: foo", str)

	str = fmt.Sprintf("%+v", res)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestAbort.func1",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestAbort",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))
}

func TestAbortIfNil(t *testing.T) {
	var res bool
	func() {
		defer Resume(func(err error) {
			res = true
		})

		AbortIf(nil)
	}()

	assert.False(t, res)
}

func TestResumePanic(t *testing.T) {
	var res error
	assert.Panics(t, func() {
		defer Resume(func(err error) {
			res = err
		})

		panic(errFoo)
	})

	assert.Nil(t, res)
}

func TestPanic(t *testing.T) {
	var res error
	func() {
		defer Recover(func(err error) {
			res = err
		})

		Panic(errFoo)
	}()

	assert.True(t, errors.Is(res, errFoo))

	str := res.Error()
	assert.Equal(t, "PANIC: foo", str)

	str = fmt.Sprintf("%s", res)
	assert.Equal(t, "PANIC: foo", str)

	str = fmt.Sprintf("%q", res)
	assert.Equal(t, `"PANIC: foo"`, str)

	str = fmt.Sprintf("%v", res)
	assert.Equal(t, "xo.Recover: PANIC: foo", str)

	str = fmt.Sprintf("%+v", res)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestPanic.func1",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestPanic",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
		"PANIC",
		"> github.com/256dpi/xo.Recover",
		">   github.com/256dpi/xo/stack.go:LN",
		"> runtime.gopanic",
		">   runtime/panic.go:LN",
		"> github.com/256dpi/xo.Panic",
		">   github.com/256dpi/xo/stack.go:LN",
		"> github.com/256dpi/xo.TestPanic.func1",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestPanic",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))
}

func TestPanicIf(t *testing.T) {
	assert.NotPanics(t, func() {
		PanicIf(nil)
	})

	assert.Panics(t, func() {
		PanicIf(F("foo"))
	})
}

func TestCatch(t *testing.T) {
	err := Catch(func() error {
		return nil
	})
	assert.NoError(t, err)

	err = Catch(func() error {
		return errFoo
	})
	assert.Error(t, err)

	err = Catch(func() error {
		Panic(errFoo)
		return nil
	})
	assert.Error(t, err)

	assert.True(t, errors.Is(err, errFoo))

	str := err.Error()
	assert.Equal(t, "PANIC: foo", str)

	str = fmt.Sprintf("%s", err)
	assert.Equal(t, "PANIC: foo", str)

	str = fmt.Sprintf("%q", err)
	assert.Equal(t, `"PANIC: foo"`, str)

	str = fmt.Sprintf("%v", err)
	assert.Equal(t, "xo.Recover: PANIC: foo", str)

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestCatch.func3",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.Catch",
		">   github.com/256dpi/xo/stack.go:LN",
		"> github.com/256dpi/xo.TestCatch",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
		"PANIC",
		"> github.com/256dpi/xo.Recover",
		">   github.com/256dpi/xo/stack.go:LN",
		"> runtime.gopanic",
		">   runtime/panic.go:LN",
		"> github.com/256dpi/xo.Panic",
		">   github.com/256dpi/xo/stack.go:LN",
		"> github.com/256dpi/xo.TestCatch.func3",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.Catch",
		">   github.com/256dpi/xo/stack.go:LN",
		"> github.com/256dpi/xo.TestCatch",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))
}

func BenchmarkAbortResume(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		func() {
			defer Resume(func(err error) {
				// do nothing
			})

			Abort(errFoo)
		}()
	}
}
