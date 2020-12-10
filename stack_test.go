package xo

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errFoo = errors.New("foo")

func TestAbortResume(t *testing.T) {
	var res error
	func() {
		defer Resume(func(err error) {
			res = err
		})

		Abort(errFoo)
	}()
	assert.Error(t, res)
	assert.True(t, errors.Is(res, errFoo))

	str := fmt.Sprintf("%+v", res)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestAbortResume.func1",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestAbortResume",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))

	func() {
		defer Resume(func(err error) {
			res = err
		})

		Abort(WF(errFoo, "bar"))
	}()
	assert.Error(t, res)
	assert.True(t, errors.Is(res, errFoo))

	str = fmt.Sprintf("%+v", res)
	assert.Equal(t, []string{
		"foo",
		"bar",
		"> github.com/256dpi/xo.TestAbortResume.func2",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestAbortResume",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))
}

func TestAbortIf(t *testing.T) {
	assert.NotPanics(t, func() {
		AbortIf(nil)
	})

	assert.Panics(t, func() {
		AbortIf(errFoo)
	})
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

func TestPanicRecover(t *testing.T) {
	var res error
	func() {
		defer Recover(func(err error) {
			res = err
		})

		Panic(errFoo)
	}()

	assert.Error(t, res)
	assert.True(t, errors.Is(res, errFoo))

	str := fmt.Sprintf("%+v", res)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestPanicRecover.func1",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestPanicRecover",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
		"PANIC",
		"> github.com/256dpi/xo.TestPanicRecover.func1",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestPanicRecover",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))

	func() {
		defer Recover(func(err error) {
			res = err
		})

		Panic(WF(errFoo, "bar"))
	}()

	assert.Error(t, res)
	assert.True(t, errors.Is(res, errFoo))

	str = fmt.Sprintf("%+v", res)
	assert.Equal(t, []string{
		"foo",
		"bar",
		"> github.com/256dpi/xo.TestPanicRecover.func2",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestPanicRecover",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
		"PANIC",
		"> github.com/256dpi/xo.TestPanicRecover.func2",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> github.com/256dpi/xo.TestPanicRecover",
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
	assert.True(t, errors.Is(err, errFoo))

	str := fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestCatch",
		">   github.com/256dpi/xo/stack_test.go:LN",
		"> testing.tRunner",
		">   testing/testing.go:LN",
		"> runtime.goexit",
		">   runtime/asm_amd64.s:LN",
	}, splitStackTrace(str))

	err = Catch(func() error {
		return WF(errFoo, "bar")
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errFoo))

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"bar",
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

	err = Catch(func() error {
		Panic(errFoo)
		return nil
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errFoo))

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"> github.com/256dpi/xo.TestCatch.func4",
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
		"> github.com/256dpi/xo.TestCatch.func4",
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

	err = Catch(func() error {
		Panic(WF(errFoo, "bar"))
		return nil
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errFoo))

	str = fmt.Sprintf("%+v", err)
	assert.Equal(t, []string{
		"foo",
		"bar",
		"> github.com/256dpi/xo.TestCatch.func5",
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
		"> github.com/256dpi/xo.TestCatch.func5",
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
