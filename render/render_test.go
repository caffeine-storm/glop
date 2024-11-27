package render_test

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nop = func(render.RenderQueueState) {}

func GivenAQueue() render.RenderQueueInterface {
	return render.MakeQueue(nop)
}

func GivenARunningQueue() render.RenderQueueInterface {
	ret := GivenAQueue()
	ret.StartProcessing()
	return ret
}

func GivenATimedQueue() render.TimedRenderQueueInterface {
	ret := GivenAQueue()
	return ret.(render.TimedRenderQueueInterface)
}

func GivenARunningTimedQueue() render.TimedRenderQueueInterface {
	ret := GivenATimedQueue()
	ret.StartProcessing()
	return ret
}

func requeueUntilPurging(q render.RenderQueueInterface, success chan bool) {
	if q.IsPurging() {
		success <- true
		return
	}

	q.Queue(func(render.RenderQueueState) {
		requeueUntilPurging(q, success)
	})
}

func runWithDeadline(deadline time.Duration, op func()) error {
	completed := make(chan bool)
	errchan := make(chan error)
	go func() {
		defer func() {
			// If 'op' panics, return the error value it paniced on.
			if err := recover(); err != nil {
				errchan <- err.(error)
			}
		}()
		op()
		completed <- true
	}()

	select {
	case <-completed:
		return nil
	case err := <-errchan:
		return err
	case <-time.After(deadline):
		return fmt.Errorf("deadline (%s) exceeded", deadline)
	}
}

func TestRenderQueueIsPurging(t *testing.T) {
	t.Run("a new queue is not purging", func(t *testing.T) {
		q := render.MakeQueue(nop)
		if q.IsPurging() {
			t.Fatalf("a new queue should not be purging")
		}

		q.StartProcessing()

		sync := make(chan bool)
		q.Queue(func(render.RenderQueueState) {
			sync <- true
		})
		<-sync

		if q.IsPurging() {
			t.Fatalf("a running queue shouldn't be purging before any purge requests")
		}

		q.Queue(func(render.RenderQueueState) {
			if q.IsPurging() {
				t.Fatalf("a running queue shouldn't be purging before any purge requests even from within a running job")
			}
			sync <- true
		})
		<-sync
	})

	t.Run("a queue is no longer purging after the Purge() call returns", func(t *testing.T) {
		q := render.MakeQueue(nop)
		if q.IsPurging() {
			t.Fatalf("a new queue should not be purging")
		}

		sync := make(chan bool)

		// Before the queue is running, requests for Purge should block but not
		// change IsPurging.
		go func() {
			q.Purge()
			sync <- true
		}()

		if q.IsPurging() {
			t.Fatalf("_requests_ to purge should not change IsPurging; the queue needs to enter that state internally")
		}

		q.StartProcessing()
		<-sync

		if q.IsPurging() {
			t.Fatalf("we've synchronized to 'after' the q.Purge(); it shouldn't be purging anymore")
		}
	})

	t.Run("a queue must return true from IsPurging if it is purging", func(t *testing.T) {
		success := make(chan bool, 1)

		q := render.MakeQueue(nop)
		q.Queue(func(render.RenderQueueState) {
			// Note that, by requeueing from a render job, we guarantee that the
			// channel buffering render jobs always has at least one job.
			requeueUntilPurging(q, success)
		})

		poll := func(c chan bool) bool {
			select {
			case <-c:
				return true
			default:
				return false
			}
		}

		// 'success' must not have been written yet.
		if poll(success) {
			t.Fatalf("the queue must not be purging before it has started")
		}
		q.StartProcessing()

		err := runWithDeadline(5*time.Millisecond, func() {
			q.Purge()
		})
		if err != nil {
			t.Fatalf("deadline exceeded!")
		}

		<-success
	})
}

type everythingIsFine struct{}

func (e *everythingIsFine) Error() string {
	return "everything is fine"
}

func TestExitOnRenderQueue(t *testing.T) {
	t.Run("runtime.GoexitOnRenderQueueIsDetectable", func(t *testing.T) {
		output := gloptest.CollectOutput(func() {
			queue := render.MakeQueue(nop)
			queue.Queue(func(render.RenderQueueState) {
				fmt.Printf("we expect to see this string in the logs\n")
				runtime.Goexit()
			})
			queue.StartProcessing()

			shouldTimeout := runWithDeadline(5*time.Millisecond, func() {
				defer func() {
					if err := recover(); err != nil {
						// re-panic with a specific error for detection
						panic(&everythingIsFine{})
					}
				}()
				queue.Purge()
				t.Fatalf("queue.Purge() should not have returned; panic is okay")
			})

			if errors.Is(shouldTimeout, &everythingIsFine{}) {
				fmt.Printf("Everything is fine!\n")
			} else {
				fmt.Printf("timeout presumably! %v\n", shouldTimeout)
				assert.NotNil(t, shouldTimeout)
			}
		})

		allOutput := strings.Join(output, "\n")
		assert.Contains(t, allOutput, "we expect to see this string in the logs")
	})
}

func TestJobTiming(t *testing.T) {
	t.Run("Can listen for jobs", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		queue := GivenATimedQueue()

		jobsSeen := 0
		allJobs := &render.JobTimingListener{
			OnNotify: func() {
				jobsSeen++
			},
			Threshold: 0, // get notified for ALL jobs
		}
		allJobs.Attach(queue)

		queue.StartProcessing()

		require.Equal(0, jobsSeen, "no job notifications should have been sent before any jobs were queued")

		jobDidRun := false
		queue.Queue(func(render.RenderQueueState) {
			// It doesn't matter what we do here; the listener has a threshold of 0
			// so should still get notified about this running.
			jobDidRun = true
		})
		queue.Purge()

		require.True(jobDidRun, "we purged the queue, but the job didn't run!")

		assert.Less(0, jobsSeen, "the listener should have been notified")
	})

	t.Run("Registering a listener must happen before StartProcessing", func(t *testing.T) {
		queue := GivenARunningTimedQueue()
		someListener := &render.JobTimingListener{}

		assert.Panics(t, func() {
			queue.SetListener(someListener)
		}, "setting a listener should not succeed if the queue is already running")
	})
}
