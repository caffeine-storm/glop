package rendertest

import "github.com/runningwild/glop/render"

// An implemenation of render.RenderQueueInterface that discards all jobs.
type discardQueue struct{}

var _ render.RenderQueueInterface = (*discardQueue)(nil)

func (*discardQueue) Queue(job render.RenderJob) {}
func (*discardQueue) Purge()                     {}
func (*discardQueue) StartProcessing()           {}
func (*discardQueue) IsPurging() bool            { return false }

func MakeDiscardingRenderQueue() render.RenderQueueInterface {
	return &discardQueue{}
}
