package render_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/caffeine-storm/gl"
	"github.com/caffeine-storm/glop/gloptest"
	"github.com/caffeine-storm/glop/render"
	"github.com/caffeine-storm/glop/render/rendertest/testbuilder"
)

var (
	JobThatCausesAGlErrorFileName      = "canary_test.go"
	_, JobThatCausesAGlErrorLineNumber = gloptest.FileLineForClosure(JobThatCausesAGlError)
)

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
