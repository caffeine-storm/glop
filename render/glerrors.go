package render

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/go-gl-legacy/glu"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gloptest"
)

func logErrorsWithAttribution(logger glog.Logger, file string, line int) {
	for {
		glErr := gl.GetError()
		if glErr == gl.NO_ERROR {
			return
		}

		glErrHex := fmt.Sprintf("0x%04x", glErr)
		glErrMsg, err := glu.ErrorString(glErr)
		if err != nil {
			// Report the make-an-error string error!
			glErrMsg = fmt.Sprintf("couldn't glu.ErrorString(%d): %s", glErr, err.Error())
		}
		logger.Warn("GlError", "file", file, "line", line, "code", glErrHex, "msg", glErrMsg)
	}
}

func LogAndClearGlErrors(logger glog.Logger) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic(fmt.Errorf("couldn't call runtime.Caller(1)"))
	}

	parts := strings.SplitAfter(file, "glop")
	file = parts[len(parts)-1][1:]

	logErrorsWithAttribution(logger, file, line)
}

func LogAndClearGlErrorsWithAttribution(logger glog.Logger, fn any) {
	file, line := gloptest.FileLineForClosure(fn)
	logErrorsWithAttribution(logger, file, line)
}
