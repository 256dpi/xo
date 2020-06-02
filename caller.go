package xo

import (
	"runtime"
	"strings"
)

// CallerInfo describes a caller.
type CallerInfo struct {
	Full  string
	Short string
	File  string
	Line  int
	Stack []uintptr
}

// Caller returns information on the current caller.
func Caller(skip ...int) CallerInfo {
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
	name := frame.Func.Name()
	file := frame.File
	line := frame.Line

	// get short name
	short := name
	if idx := strings.LastIndex(short, "/"); idx > 0 {
		short = short[idx+1:]
	}

	return CallerInfo{
		Full:  name,
		Short: short,
		File:  file,
		Line:  line,
		Stack: stack,
	}
}
