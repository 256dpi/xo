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

func ExampleTrace() {
	// install
	teardown := Debug(DebugConfig{
		TraceResolution: 100 * time.Millisecond,
	})
	defer teardown()

	// get context
	ctx := context.Background()

	// trace one
	func() {
		ctx, span := Trace(ctx, "One")
		defer span.End()

		time.Sleep(100 * time.Millisecond)

		// trace two
		func() {
			ctx, span := Trace(ctx, "Two")
			span.Log("hello world")
			defer span.End()

			time.Sleep(100 * time.Millisecond)

			// trace three
			func() {
				_, span := Trace(ctx, "Three")
				span.Tag("foo", "bar")
				defer span.End()

				time.Sleep(100 * time.Millisecond)
			}()
		}()

		// trace four
		func() {
			ctx, span := Trace(ctx, "Four")
			span.Record(F("fatal"))
			defer span.End()

			// trace five
			func() {
				_, span := Trace(ctx, "Five")
				span.Tag("baz", 42)
				defer span.End()

				time.Sleep(100 * time.Millisecond)
			}()

			time.Sleep(100 * time.Millisecond)
		}()
	}()

	// flush
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ===== TRACE =====
	// > One         ├──────────────────────────────────────────────────────────────────────────────┤   500ms
	// |   Two                       ├──────────────────────────────┤                                   200ms
	// |   :log                      •                                                                  100ms
	// |     Three                                   ├──────────────┤                                   100ms
	// |   Four                                                      ├──────────────────────────────┤   200ms
	// |   :error                                                    •                                  300ms
	// |     Five                                                    ├──────────────┤                   100ms
}

func ExampleCapture() {
	// install
	teardown := Debug(DebugConfig{
		NoReportContext:     true,
		NoReportLineNumbers: true,
	})
	defer teardown()

	// capture error
	Capture(F("some error"))

	// report error
	Reporter(SM{"foo": "bar"})(F("another error"))

	// flush
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ===== REPORT =====
	// ERROR
	// > some error (*xo.Err)
	// |   ExampleCapture (github.com/256dpi/xo): github.com/256dpi/xo/examples_test.go
	// |   main (main): _testmain.go
	// ERROR
	// • foo: bar
	// > another error (*xo.Err)
	// |   ExampleCapture (github.com/256dpi/xo): github.com/256dpi/xo/examples_test.go
	// |   main (main): _testmain.go
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
