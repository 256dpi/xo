package xo

import (
	"context"
	"fmt"
	"log"
	"time"
)

func Div(ctx context.Context, a, b int) (res int, err error) {
	return res, Run(ctx, func(ctx *Context) error {
		// tag and log
		ctx.Tag("a", a)
		ctx.Log("b: %d", b)

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
	// install
	teardown := Debug(Config{
		TraceResolution:   100 * time.Millisecond,
		NoTraceAttributes: true,
	})
	defer teardown()

	// get context
	ctx := context.Background()

	// track
	ctx1, span1 := Track(ctx, "One")
	time.Sleep(100 * time.Millisecond)
	ctx2, span2 := Track(ctx1, "Two")
	span2.Log("hello world")
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

	// flush
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ===== TRACE =====
	// One         ├──────────────────────────────────────────────────────────────────────────────┤   500ms
	//   Two                       ├──────────────────────────────┤                                   200ms
	//   :log                      •                                                                  100ms
	//     Three                                   ├──────────────┤                                   100ms
	//   Four                                                      ├──────────────────────────────┤   200ms
	//   :error                                                    •                                  300ms
	//     Five                                                                    ├──────────────┤   100ms
}

func ExampleCapture() {
	// install
	teardown := Debug(Config{
		NoEventContext:     true,
		NoEventLineNumbers: true,
	})
	defer teardown()

	// capture error
	Capture(F("some error"))

	// flush
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ===== EVENT =====
	// Level: error
	// Exceptions:
	// - some error (*xo.Err)
	//   > ExampleCapture (github.com/256dpi/xo): github.com/256dpi/xo/examples_test.go
	//   > main (main): _testmain.go
}

func ExampleSink() {
	// intercept
	reset := Intercept()
	defer reset()

	// builtin fmt
	fmt.Println("foo", "bar")
	fmt.Printf("%d %d\n", 7, 42)
	time.Sleep(10 * time.Millisecond)

	// builtin logger
	log.SetFlags(0)
	log.Println("foo", "bar")
	log.Printf("%d %d", 7, 42)
	time.Sleep(10 * time.Millisecond)

	// custom logger
	sink := Sink("FOO")
	logger := log.New(sink, "", 0)
	logger.Println("foo", "bar")
	logger.Printf("%d %d", 7, 42)
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ===== STDOUT =====
	// foo bar
	// 7 42
	// ===== LOG =====
	// foo bar
	// 7 42
	// ===== FOO =====
	// foo bar
	// 7 42
}
