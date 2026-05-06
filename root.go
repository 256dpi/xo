package xo

import (
	"fmt"
	"net/http"
	"strings"
)

// RootHandler is the middleware used to create the root trace span for
// incoming HTTP requests. Span attributes follow the OpenTelemetry HTTP
// semantic conventions (server side).
func RootHandler(cleaners ...func([]string) []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// split url
			segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

			// run cleaners
			for _, cleaner := range cleaners {
				segments = cleaner(segments)
			}

			// construct route and span name
			route := "/" + strings.Join(segments, "/")
			name := fmt.Sprintf("%s %s", r.Method, route)

			// split protocol (e.g. "HTTP/1.1" -> "http", "1.1")
			protoName, protoVersion := splitProto(r.Proto)

			// create span from request
			ctx, span := Trace(r.Context(), name)
			span.Tag("http.request.method", r.Method)
			span.Tag("http.route", route)
			span.Tag("url.full", r.URL.String())
			span.Tag("server.address", r.Host)
			span.Tag("network.protocol.name", protoName)
			span.Tag("network.protocol.version", protoVersion)

			// ensure end
			defer span.End()

			// call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func splitProto(proto string) (name, version string) {
	if i := strings.IndexByte(proto, '/'); i > 0 {
		return strings.ToLower(proto[:i]), proto[i+1:]
	}
	return strings.ToLower(proto), ""
}

// NumberCleaner will return a function that replaces number-like URL segments
// with a "#". If fullNumber is true it will only replace if the whole segment is
// a number instead of just the first character.
//
// Note: BSON ObjectIDs start with a number until 2055.
func NumberCleaner(fullNumber bool) func([]string) []string {
	return func(segments []string) []string {
		// replace numbers
		for i, s := range segments {
			// skip empty segments
			if s == "" {
				continue
			}

			// replace if number-like or number
			if (!fullNumber && isNumber(s[0:1])) || isNumber(s) {
				segments[i] = "#"
			}
		}

		return segments
	}
}
