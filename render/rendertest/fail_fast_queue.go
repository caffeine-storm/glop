package rendertest

import (
	"fmt"

	"github.com/caffeine-storm/glop/glog"
	"github.com/caffeine-storm/glop/render"
)

// Like a render.renderQueue but, if there were on-render-thread errors,
// subsequent Purge() and Queue() calls will panic.
type failfast struct {
	render.RenderQueueInterface
	Ctx *glContext
}

var _ render.RenderQueueInterface = (*failfast)(nil)

func (ff *failfast) checkErrors() {
	err := ff.Ctx.takeLastError()
	if err != nil {
		panic(fmt.Errorf("failfast queue checkErrors: %w", err))
	}
}

func (ff *failfast) Queue(job render.RenderJob) {
	ff.checkErrors()
	ff.RenderQueueInterface.Queue(job)
}

func (ff *failfast) Purge() {
	ff.RenderQueueInterface.Purge()
	ff.checkErrors()
}

func Failfastqueue(ctx *glContext) render.RenderQueueInterface {
	return &failfast{
		RenderQueueInterface: ctx.render,
		Ctx:                  ctx,
	}
}

func (ff *failfast) SetLogger(logger glog.Logger) {
	ff.RenderQueueInterface.(render.RenderQueueWithLoggerInterface).SetLogger(logger)
}
