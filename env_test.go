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

func TestLoad(t *testing.T) {
	assert.Equal(t, "", Load(Var{Name: "foo", Main: "", Devel: ""}))
	assert.Equal(t, "main", Load(Var{Name: "foo", Main: "main", Devel: "devel"}))
	assert.NoError(t, os.Setenv("foo", "baz"))
	assert.Equal(t, "baz", Load(Var{Name: "foo", Main: "main", Devel: "devel"}))

	Devel = true
	assert.NoError(t, os.Setenv("foo", ""))

	assert.Equal(t, "", Load(Var{Name: "foo", Main: "", Devel: ""}))
	assert.Equal(t, "devel", Load(Var{Name: "foo", Main: "main", Devel: "devel"}))
	assert.NoError(t, os.Setenv("foo", "baz"))
	assert.Equal(t, "baz", Load(Var{Name: "foo", Main: "main", Devel: "devel"}))

	assert.Panics(t, func() {
		Load(Var{Name: "bar", Require: true})
	})

	Devel = false

	assert.NotEmpty(t, Load(Var{Name: "file", Main: "./env.go", File: true}))
	assert.Panics(t, func() {
		assert.NotEmpty(t, Load(Var{Name: "file", Main: "./foo.go", File: true}))
	})
}
