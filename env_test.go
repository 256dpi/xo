package xo

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	err := os.Setenv("FOO", "bar")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("NUM", "42")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("FLAG", "true")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("DUR", "5s")
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
	devel := Devel
	t.Cleanup(func() { Devel = devel })

	assert.True(t, Devel)

	Devel = false

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

func TestLoadInt(t *testing.T) {
	devel := Devel
	t.Cleanup(func() { Devel = devel })

	Devel = false

	assert.Equal(t, 0, LoadInt(Var{Name: "BAR"}))
	assert.Equal(t, 7, LoadInt(Var{Name: "BAR", Main: "7", Devel: "8"}))
	assert.Equal(t, 42, LoadInt(Var{Name: "NUM", Main: "7", Devel: "8"}))

	Devel = true

	assert.Equal(t, 8, LoadInt(Var{Name: "BAR", Main: "7", Devel: "8"}))
	assert.Equal(t, 42, LoadInt(Var{Name: "NUM", Main: "7", Devel: "8"}))

	assert.Panics(t, func() {
		LoadInt(Var{Name: "FOO"})
	})
	assert.Panics(t, func() {
		LoadInt(Var{Name: "BAR", Require: true})
	})
}

func TestLoadBool(t *testing.T) {
	devel := Devel
	t.Cleanup(func() { Devel = devel })

	Devel = false

	assert.False(t, LoadBool(Var{Name: "BAR"}))
	assert.True(t, LoadBool(Var{Name: "BAR", Main: "true", Devel: "false"}))
	assert.False(t, LoadBool(Var{Name: "BAR", Main: "false", Devel: "true"}))
	assert.True(t, LoadBool(Var{Name: "FLAG", Main: "false", Devel: "false"}))

	Devel = true

	assert.False(t, LoadBool(Var{Name: "BAR", Main: "true", Devel: "false"}))
	assert.True(t, LoadBool(Var{Name: "BAR", Main: "false", Devel: "true"}))

	assert.Panics(t, func() {
		LoadBool(Var{Name: "FOO"})
	})
	assert.Panics(t, func() {
		LoadBool(Var{Name: "BAR", Require: true})
	})
}

func TestLoadDuration(t *testing.T) {
	devel := Devel
	t.Cleanup(func() { Devel = devel })

	Devel = false

	assert.Equal(t, time.Duration(0), LoadDuration(Var{Name: "BAR"}))
	assert.Equal(t, time.Second, LoadDuration(Var{Name: "BAR", Main: "1s", Devel: "2s"}))
	assert.Equal(t, 5*time.Second, LoadDuration(Var{Name: "DUR", Main: "1s", Devel: "2s"}))

	Devel = true

	assert.Equal(t, 2*time.Second, LoadDuration(Var{Name: "BAR", Main: "1s", Devel: "2s"}))

	assert.Panics(t, func() {
		LoadDuration(Var{Name: "FOO"})
	})
	assert.Panics(t, func() {
		LoadDuration(Var{Name: "BAR", Require: true})
	})
}
