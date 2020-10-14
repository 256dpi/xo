package xo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	assert.Equal(t, "", Get("FOO", ""))
	assert.Equal(t, "bar", Get("FOO", "bar"))
	assert.NoError(t, os.Setenv("FOO", "baz"))
	assert.Equal(t, "baz", Get("FOO", ""))
	assert.Equal(t, "baz", Get("FOO", "bar"))
}
