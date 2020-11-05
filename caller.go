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

// GetCaller returns information on the caller.
func GetCaller(skip, limit int) Caller {
	// ensure limit
	if limit == 0 {
		limit = 32
	}

	// get callers
	stack := make([]uintptr, limit)
	n := runtime.Callers(skip+2, stack)
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

// Includes returns whether the receiver fully includes the provided caller.
// Ignore can be set to ignore n bottom frames. Two adjacent callers will have
// the same stack except for the last frame which represents the call site.
func (c Caller) Includes(cc Caller, ignore int) bool {
	// get lengths
	cl := len(c.Stack)
	ccl := len(cc.Stack)

	// check length
	if cl < ccl {
		return false
	}

	// prepare depth
	depth := ccl - ignore

	// reverse compare stacks
	for i := 0; i < depth; i++ {
		if c.Stack[cl-1-i] != cc.Stack[ccl-1-i] {
			return false
		}
	}

	return true
}

// String will format the caller as a string.
func (c Caller) String() string {
	return c.Short
}

// Format will format the caller.
//
//  %s   short name
//  %q   "short name"
//  %v   full name
//  %+v  stack trace
//
func (c Caller) Format(s fmt.State, verb rune) {
	if verb == 's' {
		check(io.WriteString(s, c.Short))
	} else if verb == 'q' {
		check(fmt.Fprintf(s, "%q", c.Short))
	} else if verb == 'v' {
		if s.Flag('+') {
			c.Print(s)
		} else {
			check(io.WriteString(s, c.Full))
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
		check(io.WriteString(out, "> "))
		check(io.WriteString(out, frame.Function))
		check(io.WriteString(out, "\n> \t"))
		check(io.WriteString(out, frame.File))
		check(io.WriteString(out, ":"))
		check(io.WriteString(out, strconv.Itoa(frame.Line)))
		if more {
			check(io.WriteString(out, "\n"))
		}
	}
}
