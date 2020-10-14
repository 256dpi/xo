package xo

import "os"

// Get will get the specified environment variable and fallback to the specified
// value if it is missing or empty.
func Get(key, fallback string) string {
	// get value
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	return fallback
}
