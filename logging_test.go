package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSink(t *testing.T) {
	lines := captureLines(func() {
		sink := Sink("FOO")
		justPrint(sink, "Hello\nWorld!\n")
	})
	assert.Equal(t, []string{
		"===== FOO =====",
		"Hello",
		"World!",
		"",
	}, lines)
}
