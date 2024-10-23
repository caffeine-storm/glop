package render_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/runningwild/glop/render"
)

var nop = func() {}

func requeueUntilPurging(q render.RenderQueueInterface, success chan bool) {
	if q.IsPurging() {
		success <- true
		return
	}

	q.Queue(func() {
		requeueUntilPurging(q, success)
	})
}

func runWithDeadline(deadline time.Duration, op func()) error {
	completed := make(chan bool)
	go func() {
		op()
		completed <- true
	}()

	select {
	case <-completed:
		return nil
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
		q.Queue(func() {
			sync <- true
		})
		<-sync

		if q.IsPurging() {
			t.Fatalf("a running queue shouldn't be purging before any purge requests")
		}

		q.Queue(func() {
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
		q.Queue(func() {
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
