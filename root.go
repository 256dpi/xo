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
			segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

			// run cleaners
			for _, cleaner := range cleaners {
				segments = cleaner(segments)
			}

			// construct name
			path := strings.Join(segments, "/")
			name := fmt.Sprintf("%s /%s", r.Method, path)

			// create span from request
			ctx, span := Trace(r.Context(), name)
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
