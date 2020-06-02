package xo

import (
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

func splitTrace(str string) []string {
	str = strings.ReplaceAll(str, "\t", "  ")
	str = regexp.MustCompile(":\\d+").ReplaceAllString(str, ":LN")
	return strings.Split(str, "\n")
}

func captureLines(fn func()) []string {
	// capture stdout
	stdout := Stdout

	// creat pipe
	reader, writer := io.Pipe()

	// set writer
	Stdout = writer

	// prepare result
	res := make(chan []byte)

	// read all
	go func() {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}

		res <- data
	}()

	// yield reader
	fn()

	// wait a bit
	time.Sleep(10 * time.Millisecond)

	// switchback
	Stdout = stdout

	// close writer
	_ = writer.Close()

	// get result
	str := string(<-res)

	return strings.Split(str, "\n")
}
