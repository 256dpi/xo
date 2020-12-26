package xo

import (
	"io/ioutil"
	"os"
	"strings"
)

// Devel is true when the program runs in development or testing mode.
var Devel = os.Getenv("DEVEL") == "true" || Testing()

// Testing will return true if the program is likely being tested.
func Testing() bool {
	// detect if an argument has the prefix "-test." or suffix ".test"
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") || strings.HasSuffix(arg, ".test") {
			return true
		}
	}

	return false
}

// Get will get the specified environment variable and fallback to the specified
// value if it is missing or empty. The retrieved value is evaluated using Eval.
func Get(key, fallback string) string {
	// get value
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}

	// eval value
	value = Eval(key, value)

	return value
}

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

// Load will return the value of the provided environment variable. The loaded
// value is evaluated using Eval.
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

	// eval value
	value = Eval(v.Name, value)

	return value
}

// Eval will evaluate a rule found in value. If a value is formatted as
// "@file:foo" the contents of foo is loaded and returned. If a value is
// formatted as "@config:foo" the config foo is loaded and the value at the
// name key returned.
func Eval(key, value string) string {
	// handle file
	if strings.HasPrefix(value, "@file:") {
		// read file
		file, err := ioutil.ReadFile(strings.TrimPrefix(value, "@file:"))
		if err != nil {
			Panic(err)
		}

		return strings.TrimSpace(string(file))
	}

	// handle config
	if strings.HasPrefix(value, "@config:") {
		// read file
		file, err := ioutil.ReadFile(strings.TrimPrefix(value, "@config:"))
		if err != nil {
			Panic(err)
		}

		// parse lines
		for _, line := range strings.Split(string(file), "\n") {
			if strings.HasPrefix(line, key+":") {
				return strings.TrimSpace(strings.TrimPrefix(line, key+":"))
			}
		}
	}

	return value
}
