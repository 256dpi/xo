package xo

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"unicode"

	"go.opentelemetry.io/otel/api/kv"
)

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

func otelKVToMap(list []kv.KeyValue) map[string]interface{} {
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
