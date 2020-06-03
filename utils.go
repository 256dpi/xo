package xo

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"go.opentelemetry.io/otel/api/kv"
)

// M is a short-hand for a generic map.
type M = map[string]interface{}

// SM is a short-hand for a string map.
type SM = map[string]string

func raise(err error) {
	log.Printf("%v", W(err))
}

func check(_ int, err error) {
	if err != nil {
		log.Printf("%v", W(err))
	}
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

func rescale(d time.Duration, precision int) time.Duration {
	return d.Truncate(time.Duration(math.Pow10(numDigits(int64(d)) - precision)))
}

func kvToMap(list []kv.KeyValue) M {
	// convert list to map
	var dict M
	if len(list) > 0 {
		dict = M{}
		for _, item := range list {
			dict[string(item.Key)] = item.Value.AsInterface()
		}
	}

	return dict
}

func mapToKV(dict M) []kv.KeyValue {
	// collect kv
	var list []kv.KeyValue
	for key, value := range dict {
		list = append(list, kv.Infer(key, convertValue(value)))
	}

	return list
}

func convertValue(value interface{}) interface{} {
	// check primitive
	switch value.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32,
		uint64, float32, float64, string, fmt.Stringer:
		return value
	}

	// encode value
	buf, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

func iterateMap(dict M, fn func(key string, value interface{})) {
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

func buildMap(dict M) string {
	// check length
	if len(dict) == 0 {
		return ""
	}

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

	// get string
	str := builder.String()

	return str[:len(str)-1]
}

func buildBar(beforeLength, spanLength, afterLength time.Duration, width int) string {
	// check width
	if width == 0 {
		return ""
	}

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
	// check width
	if width == 0 {
		return ""
	}

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
