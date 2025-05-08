package rendertest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

type GlTestBuilder struct{}
type queueGlTestBuilder struct{}

func (b *GlTestBuilder) Run(fn func()) {
	WithGl(fn)
}

func (b *queueGlTestBuilder) Run(fn func(render.RenderQueueInterface)) {
	WithGlForTest(64, 64, func(_ system.System, queue render.RenderQueueInterface) {
		fn(queue)
	})
}

func (b *GlTestBuilder) WithQueue() *queueGlTestBuilder {
	return &queueGlTestBuilder{}
}

func GlTest() *GlTestBuilder {
	return &GlTestBuilder{}
}
