package xo

import (
	"io/ioutil"
	"os"
)

// Get will get the specified environment variable and fallback to the specified
// value if it is missing or empty.
func Get(key, fallback string) string {
	// get value
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	return fallback
}

// Devel is true when the program runs in development mode.
var Devel = os.Getenv("DEVEL") == "true"

// Var defines an environment variable.
type Var struct {
	// Then name of the variable e.g. "SOME_VAR".
	Name string

	// Whether providing a value is required.
	Require bool

	// The main default value.
	Main string

	// The development default value.
	Devel string

	// Whether the value names a file that should be read.
	File bool
}

// Load will return the value of the provided environment variable.
func Load(v Var) string {
	// get variable
	val := os.Getenv(v.Name)
	if val == "" {
		if Devel {
			val = v.Devel
		} else {
			val = v.Main
		}
	}

	// check require
	if val == "" && v.Require {
		Panic(F("missing variable " + v.Name))
	}

	// load file
	if v.File {
		buf, err := ioutil.ReadFile(val)
		if err != nil {
			Panic(WF(err, "unable to load file"))
		}
		val = string(buf)
	}

	return val
}
