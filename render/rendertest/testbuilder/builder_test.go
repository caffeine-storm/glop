package testbuilder_test

import (
	"log"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
)

func TestWithGl(t *testing.T) {
	testbuilder.New().Run(func() {
		versionString := gl.GetString(gl.VERSION)
		log.Printf("versionString: %q\n", versionString)

		if versionString == "" {
			t.Error("gl.GetString(gl.VERSION) must not return the empty string once OpenGL is initialized")
		}
	})
}
