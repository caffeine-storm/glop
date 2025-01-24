package debug

import (
	"runtime"
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
)

func LogAndClearGlErrors(logger glog.Logger) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("couldn't call runtime.Caller(1)")
	}

	parts := strings.SplitAfter(file, "glop")
	file = parts[len(parts)-1][1:]

	for {
		glErr := gl.GetError()
		if glErr == gl.NO_ERROR {
			return
		}

		logger.Warn("GlError", "file", file, "line", line, "code", glErr)
	}
}
