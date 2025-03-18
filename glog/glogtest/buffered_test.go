package glogtest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferedLogger(t *testing.T) {
	t.Run("new buffered loggers are empty", func(t *testing.T) {
		buffered := glogtest.BufferedLogger()
		assert.True(t, buffered.Empty())
	})

	t.Run("logging a message makes the buffer not empty", func(t *testing.T) {
		buffered := glogtest.BufferedLogger()
		buffered.Info("this is a test message")

		assert.False(t, buffered.Empty())
		t.Run("the logged message is in the buffer", func(t *testing.T) {
			assert.True(t, buffered.Contains("test message"))
		})
	})

}
