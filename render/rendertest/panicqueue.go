package rendertest

import (
	"github.com/caffeine-storm/glop/render"
)

// An implemenation of render.RenderQueueInterface that panics if anything is
// queued.
type panicQueue struct{}

var _ render.RenderQueueInterface = (*panicQueue)(nil)

// Adding error callbacks is a no-op; they'd never get called anyways.
func (*panicQueue) AddErrorCallback(func(render.RenderQueueInterface, error)) {}

type PanicQueueShouldNotBeCalledError struct{}

func (*PanicQueueShouldNotBeCalledError) Error() string {
	return "a panic queue must not be Queued()"
}

func (*panicQueue) Queue(job render.RenderJob) {
	panic(&PanicQueueShouldNotBeCalledError{})
}
func (*panicQueue) Purge()           {}
func (*panicQueue) StartProcessing() {}
func (*panicQueue) StopProcessing()  {}
func (*panicQueue) IsDefunct() bool  { return false } // Look like a regular queue even though we'll panic
func (*panicQueue) IsPurging() bool  { return false }

func MakePanicingRenderQueue() render.RenderQueueInterface {
	return &panicQueue{}
}
