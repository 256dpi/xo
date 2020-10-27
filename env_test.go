package xo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	err := os.Setenv("FOO", "bar")
	if err != nil {
		panic(err)
	}
}

func TestGet(t *testing.T) {
	assert.Equal(t, "", Get("BAR", ""))
	assert.Equal(t, "bar", Get("BAR", "bar"))

	assert.Equal(t, "bar", Get("FOO", ""))
	assert.Equal(t, "bar", Get("FOO", "baz"))

	assert.Equal(t, "Hello world!", Get("BAR", "@file:file"))
	assert.Panics(t, func() {
		Get("BAR", "@file:_file")
	})

	assert.Equal(t, "Hello world!", Get("BAR", "@config:config"))
	assert.Panics(t, func() {
		Get("BAR", "@config:_config")
	})
}

func TestLoad(t *testing.T) {
	assert.Equal(t, "", Load(Var{Name: "BAR", Main: "", Devel: ""}))
	assert.Equal(t, "main", Load(Var{Name: "BAR", Main: "main", Devel: "devel"}))
	assert.Equal(t, "bar", Load(Var{Name: "FOO", Main: "main", Devel: "devel"}))

	Devel = true

	assert.Equal(t, "", Load(Var{Name: "BAR", Main: "", Devel: ""}))
	assert.Equal(t, "devel", Load(Var{Name: "BAR", Main: "main", Devel: "devel"}))
	assert.Equal(t, "bar", Load(Var{Name: "FOO", Main: "main", Devel: "devel"}))

	assert.Panics(t, func() {
		Load(Var{Name: "BAR", Require: true})
	})

	Devel = false

	assert.Equal(t, "Hello world!", Load(Var{Name: "BAR", Main: "@file:file"}))
	assert.Panics(t, func() {
		Load(Var{Name: "BAR", Main: "@file:_file"})
	})

	assert.Equal(t, "Hello world!", Load(Var{Name: "BAR", Main: "@config:config"}))
	assert.Panics(t, func() {
		Load(Var{Name: "BAR", Main: "@config:_config"})
	})
}
