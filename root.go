package xo

import (
	"fmt"
	"net/http"
	"strings"
)

// RootHandler is the middleware used to create the root trace span for
// incoming HTTP requests.
func RootHandler(cleaners ...func([]string) []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// split url
			segments := strings.Split(r.URL.Path, "/")

			// run cleaners
			for _, cleaner := range cleaners {
				segments = cleaner(segments)
			}

			// construct name
			path := strings.Join(segments, "/")
			name := fmt.Sprintf("%s %s", r.Method, path)

			// create span from request
			ctx, span := Track(r.Context(), name)
			span.Tag("http.proto", r.Proto)
			span.Tag("http.host", r.Host)
			span.Tag("http.url", r.URL.String())

			// ensure end
			defer span.End()

			// call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// NumberCleaner will replace number URL segments with a "#".
func NumberCleaner(segments []string) []string {
	// replace numbers
	for i, s := range segments {
		if isNumber(s) {
			segments[i] = "#"
		}
	}

	return segments
}
