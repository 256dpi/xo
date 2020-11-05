package xo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuto(t *testing.T) {
	devel := Devel
	Devel = false
	defer func() {
		Devel = devel
	}()

	var buf bytes.Buffer

	defer Auto(Config{
		SentryDSN:    "http://token@sentry/1234",
		ReportOutput: &buf,
	})()

	Capture(F("foo"))
	CaptureSilent(F("bar"))

	assert.Contains(t, buf.String(), "foo (*xo.Err)")
	assert.NotContains(t, buf.String(), "bar (*xo.Err)")
}
