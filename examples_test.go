package xo

import (
	"context"
	"fmt"
	"time"
)

func Div(ctx context.Context, a, b int) (res int, err error) {
	return res, Run(ctx, func(ctx *Context) error {
		// tag and log
		ctx.Tag("a", a)
		ctx.Log("b", b)

		// check negative
		if b < 0 {
			return F("negative division")
		}

		// compute
		res = a / b

		return nil
	})
}

func ExampleRun() {
	// get context
	ctx := context.Background()

	// divide positive
	res, err := Div(ctx, 10, 2)
	fmt.Printf("div: %d, %v\n", res, err)

	// divide negative
	res, err = Div(ctx, 10, -2)
	fmt.Printf("div: %d, %v\n", res, err)

	// divide zero
	res, err = Div(ctx, 10, 0)
	fmt.Printf("div: %d, %v\n", res, err)

	// Output:
	// div: 5, <nil>
	// div: 0, xo.Div: negative division
	// div: 0, xo.Div: PANIC: runtime error: integer divide by zero
}

func ExampleTrack() {
	// enable debugger
	SetupDebugger(DebuggerConfig{
		TraceResolution: 100 * time.Millisecond,
	})

	// ensure teardown
	defer TeardownDebugger()

	// get context
	ctx := context.Background()

	// track
	ctx1, span1 := Track(ctx, "One")
	time.Sleep(100 * time.Millisecond)
	ctx2, span2 := Track(ctx1, "Two")
	span2.Log("hello", "world")
	time.Sleep(100 * time.Millisecond)
	_, span3 := Track(ctx2, "Three")
	span3.Tag("foo", "bar")
	time.Sleep(100 * time.Millisecond)
	span3.End()
	span2.End()
	ctx4, span4 := Track(ctx1, "Four")
	span4.Record(F("fatal"))
	time.Sleep(100 * time.Millisecond)
	_, span5 := Track(ctx4, "Five")
	span5.Tag("baz", 42)
	time.Sleep(100 * time.Millisecond)
	span5.End()
	span4.End()
	span1.End()

	// Output:
	// ===== TRACE =====
	// One         ├──────────────────────────────────────────────────────────────────────────────┤   500ms
	//   Two                       ├──────────────────────────────┤                                   200ms
	//   :log                      •                                                                  100ms   hello:"world"
	//     Three                                   ├──────────────┤                                   100ms   foo:"bar"
	//   Four                                                      ├──────────────────────────────┤   200ms
	//   :error                                                    •                                  300ms   error.message:"fatal" error.type:"*xo.Err"
	//     Five                                                                    ├──────────────┤   100ms   baz:42
}

func ExampleCapture() {
	// enable debugger
	SetupDebugger(DebuggerConfig{})

	// ensure teardown
	defer TeardownDebugger()

	// capture error
	Capture(F("some error"))

	// Output:
	// ===== EVENT =====
	// Level: error
	// Context:
	// - device: {"arch":"amd64","num_cpu":8}
	// - os: {"name":"darwin"}
	// - runtime: {"go_maxprocs":8,"go_numcgocalls":1,"go_numroutines":2,"name":"go","version":"go1.14.1"}
	// Exceptions:
	// - some error (*xo.Err)
	//   > ExampleCapture (github.com/256dpi/xo): /Users/256dpi/Development/GitHub/256dpi/xo/examples_test.go:101
	//   > main (main): _testmain.go:90
}
