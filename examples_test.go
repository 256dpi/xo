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
		TraceResolution: 10 * time.Millisecond,
	})

	// ensure teardown
	defer TeardownDebugger()

	// get context
	ctx := context.Background()

	// track one
	ctx, span1 := Track(ctx, "One")
	defer span1.End()

	time.Sleep(10 * time.Millisecond)

	// track two
	ctx, span2 := Track(ctx, "Two")
	defer span2.End()

	time.Sleep(10 * time.Millisecond)

	// track three
	_, span3 := Track(ctx, "Three")
	defer span3.End()

	// wait a bit
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ----- TRACE -----
	// One: 30ms
	// Two: 20ms
	// Three: 10ms

	// ----- TRACE -----
	// api.MainHandler (150ms)         |---------------------- 150ms ----------------------|
	//    auth.Handler (45ms)              |----- 45ms ------|
	//       auth.VerifyUser (10ms)            |-------------|
	//    api.Handler (105ms)                                    |--------------------|
	//       api.GetPost (80ms)                                   |---------------|
	//          db.LoadPosts (30ms)                               |------|
	//          db.LoadComments (50ms)                                   |--------|
}

func ExampleCapture() {
	// enable debugger
	SetupDebugger(DebuggerConfig{})

	// ensure teardown
	defer TeardownDebugger()

	// capture error
	Capture(W(F("some error")))

	// Output:
	// ----- EVENT -----
	// Level: error
	// Context:
	// - device: {"arch":"amd64","num_cpu":8}
	// - os: {"name":"darwin"}
	// - runtime: {"go_maxprocs":8,"go_numcgocalls":1,"go_numroutines":2,"name":"go","version":"go1.14.1"}
	// Exceptions:
	// - some error (*errors.fundamental)
	//   > ExampleCapture (github.com/256dpi/xo): /Users/256dpi/Development/GitHub/256dpi/xo/examples_test.go:104
	//   > main (main): _testmain.go:82
	// - some error (*errors.withStack)
	//   > ExampleCapture (github.com/256dpi/xo): /Users/256dpi/Development/GitHub/256dpi/xo/examples_test.go:104
	//   > main (main): _testmain.go:82
}
