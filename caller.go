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

// Format will format the caller.
//
//  %s   short name
//  %q   quoted short name
//  %v   full name
//  %+v  stack trace
//
func (c Caller) Format(s fmt.State, verb rune) {
	if verb == 's' {
		_, _ = io.WriteString(s, c.Short)
	} else if verb == 'q' {
		_, _ = fmt.Fprintf(s, "%q", c.Short)
	} else if verb == 'v' {
		if s.Flag('+') {
			c.Print(s)
		} else if verb == 's' || verb == 'v' {
			justPrint(s, c.Full)
		}
	}
}

// Print will print the stack to the provided writer.
func (c Caller) Print(out io.Writer) {
	// get frames
	frames := runtime.CallersFrames(c.Stack)

	// iterate frames
	var frame runtime.Frame
	more := true
	for more {
		// get next frame
		frame, more = frames.Next()

		// print frame
		justPrint(out, frame.Function)
		justPrint(out, "\n\t")
		justPrint(out, frame.File)
		justPrint(out, ":")
		justPrint(out, strconv.Itoa(frame.Line))
		if more {
			justPrint(out, "\n")
		}
	}
}
