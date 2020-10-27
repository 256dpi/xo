package xo

import (
	"io/ioutil"
	"os"
	"strings"
)

// Get will get the specified environment variable and fallback to the specified
// value if it is missing or empty.
func Get(key, fallback string) string {
	// get value
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}

	// eval
	value = eval(value)

	return value
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
}

// Load will return the value of the provided environment variable.
func Load(v Var) string {
	// get variable
	value := os.Getenv(v.Name)
	if value == "" {
		if Devel {
			value = v.Devel
		} else {
			value = v.Main
		}
	}

	// check require
	if value == "" && v.Require {
		Panic(F("missing variable " + v.Name))
	}

	// eval
	value = eval(value)

	return value
}

func eval(value string) string {
	// check for file
	if strings.HasPrefix(value, "@file:") {
		file, err := ioutil.ReadFile(strings.TrimPrefix(value, "@file:"))
		if err != nil {
			Panic(err)
		}
		value = strings.TrimSpace(string(file))
	}

	return value
}
