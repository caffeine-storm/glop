package glog_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gloptest"
	"github.com/stretchr/testify/assert"
)

func TestGlogRelevel(t *testing.T) {
	ioBuffer := &bytes.Buffer{}

	oldLevel := slog.LevelInfo
	moreVerbose := oldLevel - 4

	log := glog.New(&glog.Opts{
		Output: ioBuffer,
		Level:  oldLevel,
	})
	releveled := glog.Relevel(log, moreVerbose)

	releveled.Log(context.Background(), moreVerbose, "test-msg")

	result := ioBuffer.Bytes()
	if len(result) == 0 {
		t.Fatalf("there should have been log output")
	}
}

func TestVoidLogger(t *testing.T) {
	assert := assert.New(t)

	logger := glog.VoidLogger()

	outputLines := gloptest.CollectOutput(func() {
		// Log at Error+42 to make sure we're not just under-leveled.
		logger.Log(context.Background(), slog.LevelError+42, "some stuff", "and", "things")
	})

	assert.Empty(outputLines)
}
