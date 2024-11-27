package render

import "time"

// Instances of JobTimingListener can be registered only at Queue construction.
type JobTimingListener struct {
	OnNotify func()

	// Only jobs that took Threshold will trigger a call to OnNotify.
	Threshold time.Duration
}
