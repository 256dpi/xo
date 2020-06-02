package xo

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"go.opentelemetry.io/otel/api/kv"
)

func justPrint(out io.Writer, str string) {
	_, _ = io.WriteString(out, str)
}

func justFprintf(out io.Writer, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(out, format, args...)
}

func isBenchmark() bool {
	// check bench flag
	for _, arg := range os.Args {
		if strings.Contains(arg, "test.bench") {
			return true
		}
	}

	return false
}

func isNumber(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	return len(s) > 0
}

func numDigits(i int64) int {
	var count int
	for i != 0 {
		i /= 10
		count++
	}

	return count
}

func autoTruncate(d time.Duration, precision int) time.Duration {
	return d.Truncate(time.Duration(math.Pow10(numDigits(int64(d)) - precision)))
}

func kvToMap(list []kv.KeyValue) map[string]interface{} {
	// convert list to map
	var dict map[string]interface{}
	if len(list) > 0 {
		dict = map[string]interface{}{}
		for _, item := range list {
			dict[string(item.Key)] = item.Value.AsInterface()
		}
	}

	return dict
}

func mapToKV(dict map[string]interface{}) []kv.KeyValue {
	// collect kv
	var list []kv.KeyValue
	for key, value := range dict {
		list = append(list, kv.Infer(key, value))
	}

	return list
}

func iterateMap(dict map[string]interface{}, fn func(key string, value interface{})) {
	// collect keys
	var keys []string
	for key := range dict {
		keys = append(keys, key)
	}

	// sort keys
	sort.Strings(keys)

	// iterate
	for _, key := range keys {
		fn(key, dict[key])
	}
}

func mustEncode(value interface{}) string {
	// encode value
	buf, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

func buildMap(dict map[string]interface{}) string {
	// prepare builder
	var builder strings.Builder

	// add all key values
	iterateMap(dict, func(key string, value interface{}) {
		builder.WriteString(key)
		builder.WriteRune(':')
		switch value.(type) {
		case string:
			builder.WriteString(fmt.Sprintf("%q", value))
		default:
			builder.WriteString(fmt.Sprintf("%v", value))
		}
		builder.WriteRune(' ')
	})

	return builder.String()
}

func buildBar(beforeLength, spanLength, afterLength time.Duration, width int) string {
	// calculate total and step
	total := beforeLength + spanLength + afterLength
	step := total / time.Duration(width)
	if step == 0 {
		step = 1
	}

	// calculate points
	start := int(beforeLength / step)
	end := int((beforeLength + spanLength) / step)

	// handle zero
	if end-start == 0 {
		if end < width {
			end++
		} else {
			start--
		}
	}

	// prepare before part
	beforePart := strings.Repeat(" ", start)
	afterPart := strings.Repeat(" ", width-end)

	// prepare span part
	var spanPart string
	switch end - start {
	case 1:
		spanPart = "│"
	case 2:
		spanPart = "├┤"
	default:
		spanPart = "├" + strings.Repeat("─", (end-start)-2) + "┤"
	}

	return beforePart + spanPart + afterPart
}

func buildDot(beforeLength, afterLength time.Duration, width int) string {
	// calculate total and step
	total := beforeLength + afterLength
	step := total / time.Duration(width)
	if step == 0 {
		step = 1
	}

	// calculate position and rest
	position := int(beforeLength / step)
	rest := width - position - 1

	// handle negative
	if rest < 0 {
		position--
		rest++
	}

	// prepare before part
	beforePart := strings.Repeat(" ", position)
	afterPart := strings.Repeat(" ", rest)

	return beforePart + "•" + afterPart
}
