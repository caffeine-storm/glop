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
	"github.com/runningwild/glop/render/rendertest"
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

func GivenATimedQueue(l *render.JobTimingListener) render.RenderQueueInterface {
	return render.MakeQueueWithTiming(nop, l)
}

func GivenARunningTimedQueue(l *render.JobTimingListener) render.RenderQueueInterface {
	ret := GivenATimedQueue(l)
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

func TestRenderQueueStartProcessing(t *testing.T) {
	t.Run("StartProcessing must be called no more than once", func(t *testing.T) {
		q := render.MakeQueue(nop)
		q.StartProcessing()

		assert.Panics(t, func() {
			q.StartProcessing()
		})
	})
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
	t.Run("calling runtime.Goexit on render thread only stops the render thread", func(t *testing.T) {
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

	t.Run("support error reporting from render thread", func(t *testing.T) {
		queue := render.MakeQueue(nop)
		errorSeen := false
		thisIsFine := &everythingIsFine{}
		queue.AddErrorCallback(func(q render.RenderQueueInterface, e error) {
			if !errors.Is(e, thisIsFine) {
				t.Fatalf("got an unexpected error type: %T", e)
			}
			errorSeen = true
		})
		queue.StartProcessing()

		// Make sure to synchronize with the render thread.
		queue.Queue(nop)
		queue.Purge()

		require.False(t, errorSeen, "we shouldn't have seen the error yet")

		queue.Queue(func(render.RenderQueueState) {
			panic(thisIsFine)
		})
		queue.Purge()

		require.True(t, errorSeen, "panicing on the render thread must trigger the OnError behaviour")
	})
}

func pollingDrain(ch chan bool) int {
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			return count
		}
	}
}

func TestJobTiming(t *testing.T) {
	t.Run("Can listen for jobs", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		onNotifyEvents := make(chan bool, 64)
		listenForAllJobs := &render.JobTimingListener{
			OnNotify: func(*render.JobTimingInfo, string) {
				onNotifyEvents <- true
			},
			Threshold: 0, // get notified for ALL jobs
		}
		queue := GivenATimedQueue(listenForAllJobs)
		jobsSoFar := pollingDrain(onNotifyEvents)
		require.Equal(0, jobsSoFar, "no jobs should have run before StartProcessing()")

		queue.StartProcessing()

		jobsSoFar += pollingDrain(onNotifyEvents)
		require.LessOrEqual(jobsSoFar, 1, "only an initialization job could have run; there shouldn't be any 'user jobs' yet")

		userJobDidRun := false
		queue.Queue(func(render.RenderQueueState) {
			// It doesn't matter how long this takes; the listener has a threshold of
			// 0 so should still get notified about this running.
			userJobDidRun = true
		})
		queue.Purge()

		require.True(userJobDidRun, "we purged the queue, but the job didn't run!")

		jobsSoFar += pollingDrain(onNotifyEvents)
		assert.Equal(2, jobsSoFar, "the listener should have been notified about an initialization job and one user job")
	})

	t.Run("Source attribution is reported", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		didRun := false
		var job render.RenderJob = func(render.RenderQueueState) {
			didRun = true
		}

		attribution := ""
		listener := &render.JobTimingListener{
			OnNotify: func(info *render.JobTimingInfo, attrib string) {
				attribution = attrib
			},
			Threshold: 0,
		}
		queue := GivenARunningTimedQueue(listener)
		queue.Queue(job)
		queue.Purge()

		require.True(didRun, "queued a job and purged the queue; that job should have run")

		assert.Contains(attribution, "render/render_test.go")
	})
}

func TestRenderJob(t *testing.T) {
	t.Run("Can get source attribution", func(t *testing.T) {
		some_closure := func(render.RenderQueueState) {}
		someJob := render.RenderJob(some_closure)

		attribution := someJob.GetSourceAttribution()

		assert.Contains(t, attribution, "render/render_test.go")
	})
}

func TestAssertingOnRenderThread(t *testing.T) {
	t.Run("If not on render thread, panic", func(t *testing.T) {
		assert.Panics(t, render.MustBeOnRenderThread)
	})
	t.Run("If on a render thread, relax", func(t *testing.T) {
		rendertest.DeprecatedWithGl(func() {
			render.MustBeOnRenderThread()
		})
	})
}
