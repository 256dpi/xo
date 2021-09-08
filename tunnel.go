package xo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/256dpi/serve"
)

const envelopeContentType = "application/x-sentry-envelope"

var newlineBytes = []byte("\n")

// VerifyDSN will require the submitted DNS to match one of the provided DSNs.
func VerifyDSN(list ...string) func(*http.Request, string, *url.URL) bool {
	return func(_ *http.Request, reqDSN string, _ *url.URL) bool {
		for _, dsn := range list {
			if reqDSN == dsn {
				return true
			}
		}
		return false
	}
}

// Tunnel returns a handler that forwards received Sentry envelopes to the
// endpoint specified by the DSN received in the envelopes. The optional verify
// callback may set to verify received DSNs. This handler can be used together
// with the sentry tunnel feature to relay browser errors via a custom endpoint.
func Tunnel(limit int64, verify func(*http.Request, string, *url.URL) bool, reporter func(error)) http.Handler {
	// prepare client
	var client http.Client

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enforce limit
		if limit != 0 {
			serve.LimitBody(w, r, limit)
		}

		// read body
		body, err := io.ReadAll(r.Body)
		if err == serve.ErrBodyLimitExceeded {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		} else if err != nil {
			reporter(W(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// split out header
		parts := bytes.SplitN(body, newlineBytes, 2)

		// parse header data
		var headerData struct {
			DSN string `json:"dsn"`
		}
		err = json.Unmarshal(parts[0], &headerData)
		if err != nil {
			reporter(W(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// parse dsn
		dsn, err := url.Parse(headerData.DSN)
		if err != nil {
			reporter(W(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// verify dsn if available
		if verify != nil && !verify(r, headerData.DSN, dsn) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// prepare target
		target := fmt.Sprintf("%s://%s/api/%s/envelope/", dsn.Scheme, dsn.Host, strings.Trim(dsn.Path, "/"))

		// post envelope
		res, err := client.Post(target, envelopeContentType, bytes.NewReader(body))
		if err != nil {
			reporter(W(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		// check status
		if res.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		// copy result
		_, err = io.Copy(w, res.Body)
		if err != nil {
			reporter(W(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
