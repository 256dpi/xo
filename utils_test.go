package xo

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var arch = runtime.GOARCH

func TestBuildMeta(t *testing.T) {
	str := buildMeta(nil)
	assert.Equal(t, "", str)

	str = buildMeta(M{
		"a": 2,
		"b": "foo",
		"c": true,
		"d": M{
			"foo": "bar",
		},
	})
	assert.Equal(t, `a:2 b:foo c:true d:{"foo":"bar"}`, str)
}

func TestBuildBar(t *testing.T) {
	str := buildBar(0, 0, 0, 0)
	assert.Equal(t, "", str)

	str = buildBar(0, 0, 0, 1)
	assert.Equal(t, "│", str)

	str = buildBar(1, 0, 0, 1)
	assert.Equal(t, "│", str)

	str = buildBar(0, 1, 0, 1)
	assert.Equal(t, "│", str)

	str = buildBar(0, 0, 1, 1)
	assert.Equal(t, "│", str)

	str = buildBar(0, 2, 0, 2)
	assert.Equal(t, "├┤", str)

	str = buildBar(5, 5, 5, 15)
	assert.Equal(t, "     ├───┤     ", str)
}

func TestBuildDot(t *testing.T) {
	str := buildDot(0, 0, 0)
	assert.Equal(t, "", str)

	str = buildDot(0, 0, 1)
	assert.Equal(t, "•", str)

	str = buildDot(1, 0, 1)
	assert.Equal(t, "•", str)

	str = buildDot(0, 1, 1)
	assert.Equal(t, "•", str)

	str = buildDot(5, 5, 10)
	assert.Equal(t, "     •    ", str)
}
