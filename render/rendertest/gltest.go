package rendertest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

type GlTestBuilder struct {
	Dx, Dy uint
}

type queueGlTestBuilder struct {
	ctx *GlTestBuilder
}

func (b *GlTestBuilder) Run(fn func()) {
	dx, dy := b.Dx, b.Dy
	if dx == 0 || dy == 0 {
		// Pick a default of 64x64
		dx = 64
		dy = 64
	}
	WithGlForTest(int(dx), int(dy), func(sys system.System, queue render.RenderQueueInterface) {
		queue.Queue(func(render.RenderQueueState) {
			fn()
		})
		queue.Purge()
	})
}

func (b *GlTestBuilder) WithSize(dx, dy uint) *GlTestBuilder {
	b.Dx = dx
	b.Dy = dy
	return b
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
