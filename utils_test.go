package xo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMap(t *testing.T) {
	str := buildMap(nil)
	assert.Equal(t, "", str)

	str = buildMap(M{
		"a": 2,
		"b": "foo",
		"c": M{
			"d": "bar",
		},
	})
	assert.Equal(t, `a:2 b:"foo" c:map[d:bar]`, str)
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
