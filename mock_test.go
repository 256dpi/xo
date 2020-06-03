package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockSink(t *testing.T) {
	Trap(func(mock *Mock) {
		sink := Sink("foo")
		_, _ = sink.Write([]byte("Hello World!"))
		assert.Equal(t, "Hello World!", mock.Sinks["foo"].String())
	})
}
