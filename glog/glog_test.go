package glog_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/runningwild/glop/glog"
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
