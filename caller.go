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
}

// UnknownCaller is returned if the caller is unknown.
var UnknownCaller = CallerInfo{
	Full:  "Unknown",
	Short: "Unknown",
	File:  "unknown.go",
	Line:  1,
}

// Caller returns information on the current caller.
func Caller(skip ...int) CallerInfo {
	// sum
	sum := 1
	for _, s := range skip {
		sum += s
	}

	// get caller
	pc, file, line, ok := runtime.Caller(sum)
	if !ok {
		return UnknownCaller
	}

	// get name
	name := runtime.FuncForPC(pc).Name()

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
	}
}
