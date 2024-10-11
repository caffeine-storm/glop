package rendertest

import "github.com/runningwild/glop/render"

// An implemenation of render.RenderQueueInterface that discards all jobs.
type discardQueue struct{}

var _ render.RenderQueueInterface = (*discardQueue)(nil)

func (*discardQueue) Queue(f func())   {}
func (*discardQueue) Purge()           {}
func (*discardQueue) StartProcessing() {}

func MakeDiscardingRenderQueue() render.RenderQueueInterface {
	return &discardQueue{}
}
