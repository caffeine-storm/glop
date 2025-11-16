package glogtest

import (
	"bytes"

	"github.com/caffeine-storm/glop/glog"
)

type BufferedLogger struct {
	glog.Logger
	buffer *bytes.Buffer
}

func NewBufferedLogger() *BufferedLogger {
	buf := &bytes.Buffer{}
	return &BufferedLogger{
		Logger: glog.New(&glog.Opts{
			Output: buf,
		}),
		buffer: buf,
	}
}

func (logger *BufferedLogger) Empty() bool {
	return logger.buffer.String() == ""
}

func (logger *BufferedLogger) Contains(substr string) bool {
	return bytes.Contains(logger.buffer.Bytes(), []byte(substr))
}

func (logger *BufferedLogger) String() string {
	return string(logger.buffer.Bytes())
}
