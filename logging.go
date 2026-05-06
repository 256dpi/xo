package xo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Stdout is the original stdout.
var Stdout io.Writer = os.Stdout

// SinkFactory is the factory used by Sink() to create sinks.
var SinkFactory = func(name string) io.WriteCloser {
	// create pipe
	reader, writer := io.Pipe()

	// drain
	go Drain(name, reader)

	return writer
}

// Intercept will replace os.Stdout with a logging sink named "STDOUT". It will
// also redirect the output of the log package to a logging sink named "LOG".
// The returned function can be called to restore the original state.
func Intercept() func() {
	// capture stdout
	stdout := os.Stdout
	Stdout = os.Stdout

	// capture output
	output := log.Writer()

	// create pipe
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// replace stdout
	os.Stdout = writer

	// replace logging output
	log.SetOutput(Sink("LOG"))

	// drain
	go Drain("STDOUT", reader)

	return func() {
		// reset stdout
		os.Stdout = stdout

		// reset logging output
		log.SetOutput(output)

		// close writer
		_ = writer.Close()
	}
}

// Sink will return a new named logging sink.
func Sink(name string) io.WriteCloser {
	return SinkFactory(name)
}

// Drain will read log lines from the reader and write them to Stdout.
func Drain(name string, reader io.Reader) {
	// prepare queue
	queue := make(chan string, 32)

	// prepare output
	var output bytes.Buffer

	// run writer
	go func() {
		// reusable timer for the flush window (avoids per-iteration allocs)
		timer := time.NewTimer(time.Millisecond)
		if !timer.Stop() {
			<-timer.C
		}

		for {
			// await first line
			line, ok := <-queue
			if !ok {
				return
			}

			// reset buffer
			output.Reset()

			// write header
			check(output.WriteString(fmt.Sprintf("===== %s =====\n", name)))

			// write first line
			check(output.WriteString(line))

			// add lines until idle for one flush window
			timer.Reset(time.Millisecond)
		batch:
			for {
				select {
				case line, ok = <-queue:
					if !ok {
						if !timer.Stop() {
							<-timer.C
						}
						break batch
					}
					check(output.WriteString(line))
				case <-timer.C:
					break batch
				}
			}

			// write out
			check(Stdout.Write(output.Bytes()))

			// exit if queue closed
			if !ok {
				return
			}
		}
	}()

	// prepare input buffer
	input := bufio.NewReader(reader)

	// read lines
	for {
		// get next line
		line, err := input.ReadString('\n')
		if err == io.EOF {
			close(queue)
			return
		} else if err != nil {
			panic(err)
		}

		// queue line
		queue <- line
	}
}
