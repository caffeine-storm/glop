package glog_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/caffeine-storm/glop/glog"
	"github.com/caffeine-storm/glop/gloptest"
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

func TestGlogGetOpts(t *testing.T) {
	t.Run("options passed to constructor should be accessible", func(t *testing.T) {
		defaultOpts := glog.Opts{
			DoNotAddSource: false,
		}
		differentOpts := glog.Opts{
			DoNotAddSource: true,
		}
		defaultLog := glog.New(&defaultOpts)
		differentLog := glog.New(&differentOpts)

		lhs := defaultLog.GetOpts()
		rhs := differentLog.GetOpts()
		if lhs.AddSource == rhs.AddSource {
			t.Errorf("expecting differnt AddSource options")
		}
	})
}

func TestGlogRedirect(t *testing.T) {
	buffer1 := &bytes.Buffer{}
	buffer2 := &bytes.Buffer{}

	logger1 := glog.New(&glog.Opts{
		Output: buffer1,
	})
	logger2 := glog.WithRedirect(logger1, buffer2)

	logger1.Error("log1 message")
	logger2.Error("log2 message")

	log1 := buffer1.String()
	log2 := buffer2.String()

	if strings.Contains(log1, "log2 message") {
		t.Error("log1 should not have messages from logger2")
	}
	if strings.Contains(log2, "log1 message") {
		t.Error("log2 should not have messages from logger1")
	}

	if !strings.Contains(log1, "log1 message") {
		t.Error("log1 should have messages from logger1")
	}
	if !strings.Contains(log2, "log2 message") {
		t.Error("log2 should have messages from logger2")
	}
}

func TestGlogRedirectPreservesLevel(t *testing.T) {
	buffer1 := &bytes.Buffer{}
	buffer2 := &bytes.Buffer{}
	lvl := &slog.LevelVar{}
	lvl.Set(42)
	logger1 := glog.New(&glog.Opts{
		Output: buffer1,
		Level:  lvl,
	})
	logger2 := glog.WithRedirect(logger1, buffer2)

	if logger2.GetOpts().Level == nil {
		t.Fatalf("the wrapped logger had a level; so should the wrapper")
	}

	if logger2.GetOpts().Level.Level() != 42 {
		t.Fatalf("the wrapped logger had a specific level; so should the wrapper")
	}
}

func TestVoidLogger(t *testing.T) {
	assert := assert.New(t)

	outputLines := gloptest.CollectOutput(func() {
		logger := glog.VoidLogger()

		// Log at Error+42 to make sure we're not just under-leveled.
		logger.Log(context.Background(), slog.LevelError+42, "some stuff", "and", "things")
	})

	assert.Empty(outputLines)
}

func TestSourceAttribution(t *testing.T) {
	t.Run("trace messages should contain a source file and line number", func(t *testing.T) {
		assert := assert.New(t)

		outputLines := gloptest.CollectOutput(func() {
			logger := glog.TraceLogger()
			logger = glog.Relevel(logger, glog.LevelTrace)
			logger.Trace("a message for you")
		})

		output := strings.Join(outputLines, "\n")
		assert.Contains(output, "a message for you")
		assert.Contains(output, "glog_test.go")
	})
	t.Run("info messages should contain a source file and line number", func(t *testing.T) {
		assert := assert.New(t)

		outputLines := gloptest.CollectOutput(func() {
			logger := glog.InfoLogger()
			logger.Info("a message for you")
		})

		output := strings.Join(outputLines, "\n")
		assert.Contains(output, "a message for you")
		assert.Contains(output, "glog_test.go")
	})
}

func TestTraceLoggerDisabledByDefault(t *testing.T) {
	outputLines := gloptest.CollectOutput(func() {
		glog.TraceLogger().Trace("should not show up; tracing needs explicit opt-in")
	})

	assert.Empty(t, outputLines, "there should not be any output")
}

func TestGlogWithAttrs(t *testing.T) {
	outputLines := gloptest.CollectOutput(func() {
		rawLog := glog.InfoLogger()
		logWithContext := rawLog.WithAttrs("foo", "bar")

		rawLog.Info("but why?", "because", "nature")
		logWithContext.Info("some stuff", "and", "things")
	})

	catted := strings.Join(outputLines, "\n")
	assert.Contains(t, catted, "foo=bar")
}
