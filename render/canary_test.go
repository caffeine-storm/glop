package render_test

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
)

var JobThatCausesAGlErrorFileName = "canary_test.go"
var JobThatCausesAGlErrorLineNumber = computeClosureLineNumber(JobThatCausesAGlError)

func computeClosureLineNumber(fn any) int {
	reflected := reflect.ValueOf(fn)
	up := uintptr(reflected.UnsafePointer())
	funky := runtime.FuncForPC(up)
	_, line := funky.FileLine(up)
	return line
}

func JobThatCausesAGlError(render.RenderQueueState) {
	// do a thing that will cause a GL error
	gl.End()
}

func ContainsExampleError(logContents string) bool {
	if !strings.Contains(logContents, JobThatCausesAGlErrorFileName) {
		return false
	}

	return strings.Contains(logContents, fmt.Sprintf("line=%d", JobThatCausesAGlErrorLineNumber))
}

func TestWeCanCauseAGlError(t *testing.T) {
	testbuilder.Run(func(queue render.RenderQueueInterface) {
		var errval gl.GLenum
		queue.Queue(func(qs render.RenderQueueState) {
			JobThatCausesAGlError(qs)
			errval = gl.GetError()
		})
		queue.Purge()
		if errval == gl.NO_ERROR {
			t.Fatalf("there should have been an error")
		}
	})
}
