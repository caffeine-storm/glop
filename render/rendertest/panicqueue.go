package rendertest

import (
	"fmt"

	"github.com/runningwild/glop/render"
)

// An implemenation of render.RenderQueueInterface that panics if anything is
// queued.
type panicQueue struct{}

var _ render.RenderQueueInterface = (*panicQueue)(nil)

func (*panicQueue) Queue(job render.RenderJob) {
	panic(fmt.Errorf("Queue() called on panicQueue"))
}
func (*panicQueue) Purge()           {}
func (*panicQueue) StartProcessing() {}
func (*panicQueue) IsPurging() bool  { return false }

func MakePanicingRenderQueue() render.RenderQueueInterface {
	return &panicQueue{}
}
