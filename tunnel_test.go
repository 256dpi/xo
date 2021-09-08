package xo

import (
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/256dpi/serve"
	"github.com/stretchr/testify/assert"
)

var exampleEnvelope = `{ "event_id": "123", "dsn": "http://key@0.0.0.0:1337/42" }
{ "type": "event", "length": 41, "content_type": "application/json", "filename": "application.log" }
{ "message": "hello world", "level": "error" }`

func TestTunnel(t *testing.T) {
	listener, err := net.Listen("tcp", "0.0.0.0:1337")
	assert.NoError(t, err)

	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/42/envelope/", r.URL.String())
		assert.Equal(t, http.Header{
			"Accept-Encoding": {"gzip"},
			"Content-Length":  {"206"},
			"Content-Type":    {"application/x-sentry-envelope"},
			"User-Agent":      {"Go-http-client/1.1"},
		}, r.Header)
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, exampleEnvelope, string(body))
		_, _ = w.Write([]byte("OK"))
	})}
	go srv.Serve(listener)
	defer srv.Close()

	handler := Tunnel(0, VerifyDSN("http://key@0.0.0.0:1337/42"), Panic)
	res := serve.Record(handler, "POST", "/", nil, exampleEnvelope)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, "OK", res.Body.String())

	handler = Tunnel(1, VerifyDSN("http://key@0.0.0.0:1337/42"), Panic)
	res = serve.Record(handler, "POST", "/", nil, exampleEnvelope)
	assert.Equal(t, http.StatusRequestEntityTooLarge, res.Result().StatusCode)

	handler = Tunnel(0, VerifyDSN(), Panic)
	res = serve.Record(handler, "POST", "/", nil, exampleEnvelope)
	assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode)
}
