package render_test

import (
	"testing"

	"github.com/runningwild/glop/render"
)

var nop = func() {}

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
	})
}
