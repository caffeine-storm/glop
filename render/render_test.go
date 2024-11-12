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
)

var nop = func(render.RenderQueueState) {}

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
