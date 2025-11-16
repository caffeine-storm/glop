// package rendertest implements testing helpers that are useful for projects
// that are using the 'render' package.
package rendertest

import (
	"testing"

	"github.com/caffeine-storm/glop/render"
	"github.com/stretchr/testify/assert"
)

func AssertOnRenderThread(*testing.T) {
	render.MustBeOnRenderThread()
}

func AssertOffRenderThread(t *testing.T) {
	assert.Panics(t, func() {
		render.MustBeOnRenderThread()
	})
}
