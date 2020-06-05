package xo

import (
	"errors"
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
	assert.Equal(t, "foo", res.Error())
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
	assert.Equal(t, "PANIC: foo", res.Error())
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
