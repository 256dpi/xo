package xo

import (
	"net/http"
	"testing"

	"github.com/256dpi/serve"
	"github.com/stretchr/testify/assert"
)

func TestRootHandler(t *testing.T) {
	Test(func(tester *Tester) {
		handler := serve.Compose(
			RootHandler(NumberCleaner(false)),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, span := Trace(r.Context(), "foo")
				defer span.End()

				w.WriteHeader(http.StatusOK)
			}),
		)

		res := serve.Record(handler, "GET", "/foo/123/bar/1bc5", nil, "")
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "", res.Body.String())

		assert.Equal(t, []VSpan{
			{
				Name: "foo",
			},
			{
				Name: "GET /foo/#/bar/#",
				Attributes: M{
					"http.proto": "HTTP/1.1",
					"http.host":  "example.com",
					"http.url":   "/foo/123/bar/1bc5",
				},
			},
		}, tester.ReducedSpans(0))
	})
}
