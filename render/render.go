package render

import (
	"log/slog"
	"runtime"
	"sync/atomic"
)

type RenderJob func()

type RenderQueueInterface interface {
	Queue(f RenderJob)
	Purge()
	StartProcessing()
	IsPurging() bool
}

type renderQueue struct {
	render_funcs chan RenderJob
	purge        chan chan bool
	is_running   bool
	is_purging   atomic.Bool
}

func (q *renderQueue) loop() {
	defer close(q.purge)
	for {
		select {
		case f := <-q.render_funcs:
			f()
		case ack := <-q.purge:
			defer close(ack)
			q.is_purging.Store(true)
			for {
				select {
				case f := <-q.render_funcs:
					f()
				default:
					goto purged
				}
			}
		purged:
			q.is_purging.Store(false)
			ack <- true
		}
	}
}

func MakeQueue(initialization RenderJob) RenderQueueInterface {
	result := renderQueue{
		render_funcs: make(chan RenderJob, 1000),
		purge:        make(chan chan bool),
		is_running:   false,
		is_purging:   atomic.Bool{}, // zero-value is false
	}

	// We're guaranteed that this render job will run first. We can include our
	// own initialization that should happen on the loop's thread.
	result.Queue(func() {
		runtime.LockOSThread()
		initialization()
	})
	return &result
}

// TODO(tmckee): inject a GL dependency to given func for testability and to
// keep arbitrary code from calling GL off of the render thread.
func (q *renderQueue) Queue(f RenderJob) {
	q.render_funcs <- f
}

// Waits until all render thread functions have been run
func (q *renderQueue) Purge() {
	if !q.is_running {
		slog.Warn("render.RenderQueue.Purge called on non-started queue")
	}
	ack := make(chan bool)
	q.purge <- ack
	_, ok := <-ack
	if !ok {
		panic("ack channel was closed!")
	}
}

func (q *renderQueue) StartProcessing() {
	if q.is_running {
		panic("must not call 'StartProcessing' twice")
	}
	q.is_running = true
	go q.loop()
}

func (q *renderQueue) IsPurging() bool {
	return q.is_purging.Load()
}
