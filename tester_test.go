package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockSink(t *testing.T) {
	Test(func(tester *Tester) {
		sink := Sink("foo")
		_, _ = sink.Write([]byte("Hello World!"))
		assert.Equal(t, "Hello World!", tester.Sinks["foo"].String())
	})
}
