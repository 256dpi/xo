package xo

import (
	"net/http"
	"testing"
	"time"

	"github.com/256dpi/serve"
	"github.com/stretchr/testify/assert"
)

func TestRootHandler(t *testing.T) {
	Trap(func(mock *Mock) {
		handler := serve.Compose(
			RootHandler(NumberCleaner),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, span := Track(r.Context(), "foo")
				defer span.End()

				time.Sleep(time.Millisecond)
				w.WriteHeader(http.StatusOK)
			}),
		)

		res := serve.Record(handler, "GET", "/foo/123/bar", nil, "")
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "", res.Body.String())

		assert.Equal(t, []MemorySpan{
			{
				Name:     "foo",
				Duration: time.Millisecond,
			},
			{
				Name:     "GET /foo/#/bar",
				Duration: time.Millisecond,
				Attributes: map[string]interface{}{
					"peer.address": "192.0.2.1:1234",
					"http.proto":   "HTTP/1.1",
					"http.method":  "GET",
					"http.host":    "example.com",
					"http.path":    "/foo/#/bar",
					"http.url":     "/foo/123/bar",
					"http.length":  int64(0),
					"http.header":  "map[]",
				},
			},
		}, mock.ReducedSpans(time.Millisecond))
	})
}
