package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/256dpi/serve"

	"github.com/256dpi/xo"
)

func main() {
	// install
	xo.Debug(xo.Config{})

	// run repl
	go repl()

	// prepare mux
	mux := http.NewServeMux()
	mux.HandleFunc("/calc", calculate)

	// prepare handler
	handler := serve.Compose(
		xo.RootHandler(),
		mux,
	)

	// listen and serve
	_ = http.ListenAndServe(":8000", handler)
}

func calculate(w http.ResponseWriter, r *http.Request) {
	// trace
	ctx, span := xo.Trace(r.Context(), "calculate")
	defer span.End()

	// get params
	a := r.URL.Query().Get("a")
	b := r.URL.Query().Get("b")

	// call business logic
	result, err := businessLogic(ctx, a, b)
	if err != nil {
		xo.Capture(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// tag result
	span.Tag("result", result)

	// log result
	log.Printf("result: %s", result)

	// write result
	_, _ = w.Write([]byte(result))
}

func businessLogic(ctx context.Context, a, b string) (res string, err error) {
	return res, xo.Run(ctx, func(ctx *xo.Context) error {
		// tag params
		ctx.Tag("a", a)
		ctx.Tag("b", b)

		// parse param a
		aa, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			return xo.W(err)
		}

		// parse param b
		bb, err := strconv.ParseInt(b, 10, 64)
		if err != nil {
			return xo.W(err)
		}

		// check negative
		if aa < 0 || bb < 0 {
			return xo.F("negative params")
		}

		// compute result
		res = strconv.FormatInt(aa/bb, 10)

		return nil
	})
}

func repl() {
	// prepare buffer
	buf := bufio.NewReader(os.Stdin)

	// read lines
	for {
		// read line
		str, err := buf.ReadString('\n')
		if err != nil {
			panic(err)
		}

		// trim space
		str = strings.TrimSpace(str)

		// prepare params
		var a string
		var b string

		// split string
		params := strings.Split(str, " ")
		if len(params) == 2 {
			a = params[0]
			b = params[1]
		}

		// compute url
		url := fmt.Sprintf("http://0.0.0.0:8000/calc?a=%s&b=%s", a, b)

		// make request
		res, err := http.Post(url, "text/plain", nil)
		if err != nil {
			panic(err)
		}

		// close body
		_ = res.Body.Close()
	}
}
