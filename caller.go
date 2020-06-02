package xo

import (
	"runtime"
	"strings"
)

// Caller describes a caller.
type Caller struct {
	Short string
	Full  string
	File  string
	Line  int
	Stack []uintptr
}

// GetCaller returns information on the current caller.
func GetCaller(skip ...int) Caller {
	// sum skip
	sum := 2
	for _, s := range skip {
		sum += s
	}

	// get callers
	stack := make([]uintptr, 100)
	n := runtime.Callers(sum, stack)
	stack = stack[:n]

	// get first frame
	frame, _ := runtime.CallersFrames(stack).Next()

	// get name, file and line
	name := frame.Function
	file := frame.File
	line := frame.Line

	// get short name
	short := name
	if idx := strings.LastIndex(short, "/"); idx > 0 {
		short = short[idx+1:]
	}

	return Caller{
		Short: short,
		Full:  name,
		File:  file,
		Line:  line,
		Stack: stack,
	}
}
