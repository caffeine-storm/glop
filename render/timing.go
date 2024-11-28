package render

import "time"

type JobTimingInfo struct {
	// Time elapsed while running the RenderJob func.
	RunTime time.Duration
	// Time elapsed between when a RenderJob was Queue'd on a
	// RenderQueueInterface and when the RenderJob func started.
	QueueTime time.Duration
}

// Instances of JobTimingListener can be registered only at Queue construction.
type JobTimingListener struct {
	// NOTE: this notification runs on the render thread that ran the slow job.
	// Care should be taken not to make a bad situation worse!
	//
	// Called after a render job took longer than Threshold. The actual time
	// taken and the job's source attribution is also given.
	OnNotify func(*JobTimingInfo, string)

	// Only jobs that took Threshold will trigger a call to OnNotify.
	Threshold time.Duration
}
