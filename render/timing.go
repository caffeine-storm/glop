package render

import "time"

// Instances of JobTimingListener can be registered only at Queue construction.
type JobTimingListener struct {
	// NOTE: this notification runs on the render thread that ran the slow job.
	// Care should be taken not to make a bad situation worse!
	//
	// Called after a render job took longer than Threshold. The actual time
	// taken and the job's source attribution is also given.
	OnNotify func(time.Duration, string)

	// Only jobs that took Threshold will trigger a call to OnNotify.
	Threshold time.Duration
}
