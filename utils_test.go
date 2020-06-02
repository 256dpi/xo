package xo

import (
	"regexp"
	"strings"
)

func splitTrace(str string) []string {
	str = strings.ReplaceAll(str, "\t", "  ")
	str = regexp.MustCompile(":\\d+").ReplaceAllString(str, ":LN")
	return strings.Split(str, "\n")
}
