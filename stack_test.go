package xo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errAbortTest = errors.New("foo")

func TestAbort(t *testing.T) {
	var res error
	func() {
		defer Resume(func(err error) {
			res = err
		})

		Abort(errAbortTest)
	}()

	assert.True(t, errors.Is(res, errAbortTest))
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

func TestPanic(t *testing.T) {
	var res error
	assert.Panics(t, func() {
		defer Resume(func(err error) {
			res = err
		})

		panic(errAbortTest)
	})

	assert.Nil(t, res)
}

func BenchmarkAbortResume(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		func() {
			defer Resume(func(err error) {
				// do nothing
			})

			Abort(errAbortTest)
		}()
	}
}
