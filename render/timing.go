package render

import "time"

type JobTimingListener struct {
	OnNotify func()

	// Only jobs that took Threshold will trigger a call to OnNotify.
	Threshold time.Duration
}

func (listener *JobTimingListener) Attach(queue TimedRenderQueueInterface) {
	queue.AttachListener(listener)
}
