package render_test

import (
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
)

var JobThatCausesAGlErrorFileName = "canary_test.go"

func JobThatCausesAGlError(render.RenderQueueState) {
	// do a thing that will cause a GL error
	out := [1]int32{}
	notAValidInput := gl.GLenum(0)
	gl.GetIntegerv(notAValidInput, out[:])
}

func ContainsExampleError(logContents string) bool {
	return strings.Contains(logContents, JobThatCausesAGlErrorFileName)
}
