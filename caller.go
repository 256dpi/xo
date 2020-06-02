package xo

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
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
	stack := make([]uintptr, 32)
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

// String will format the caller as a string.
func (c Caller) String() string {
	return c.Short
}

// Format will format the caller depending on %s, %v or %+v.
func (c Caller) Format(s fmt.State, verb rune) {
	if verb == 's' {
		_, _ = io.WriteString(s, c.Short)
	} else if verb == 'v' {
		if s.Flag('+') {
			c.Dump(s)
		} else if verb == 's' || verb == 'v' {
			ioWriteString(s, c.Full)
		}
	}
}

// Dump will dump the stack to the provided writer.
func (c Caller) Dump(out io.Writer) {
	// get frames
	frames := runtime.CallersFrames(c.Stack)

	// iterate frames
	var frame runtime.Frame
	more := true
	for more {
		// get next frame
		frame, more = frames.Next()

		// print frame
		ioWriteString(out, frame.Function)
		ioWriteString(out, "\n\t")
		ioWriteString(out, frame.File)
		ioWriteString(out, ":")
		ioWriteString(out, strconv.Itoa(frame.Line))
		if more {
			ioWriteString(out, "\n")
		}
	}
}
