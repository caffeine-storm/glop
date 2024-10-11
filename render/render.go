package render

import (
	"log"
	"runtime"
)

type RenderQueueInterface interface {
	Queue(f func())
	Purge()
	StartProcessing(f func())
}

type renderQueue struct {
	render_funcs chan func()
	purge        chan bool
	is_running   bool
}

func (q *renderQueue) loop(fn func()) {
	runtime.LockOSThread()
	for {
		select {
		case f := <-q.render_funcs:
			fn() // XXX: T_T: BAD: chicken and egg problem w.r.t. initializing opengl _on_ the render thread...
			f()
		case <-q.purge:
			for {
				select {
				case f := <-q.render_funcs:
					fn() // XXX: T_T: BAD: chicken and egg problem w.r.t. initializing opengl _on_ the render thread...
					f()
				default:
					goto purged
				}
			}
		purged:
			q.purge <- true
		}
	}
}

func MakeQueue() RenderQueueInterface {
	result := renderQueue{
		render_funcs: make(chan func(), 1000),
		purge:        make(chan bool),
		is_running:   false,
	}

	return &result
}

// TODO(tmckee): inject a GL dependency to given func for testability and to
// keep arbitrary code from calling GL off of the render thread.
func (q *renderQueue) Queue(f func()) {
	q.render_funcs <- f
}

// Waits until all render thread functions have been run
func (q *renderQueue) Purge() {
	if !q.is_running {
		log.Printf("WARNING: render.RenderQueue.Purge called on non-started queue")
	}
	q.purge <- true
	<-q.purge
}

func (q *renderQueue) StartProcessing(fn func()) {
	if q.is_running {
		panic("must not call 'StartProcessing' twice")
	}
	q.is_running = true
	go q.loop(fn)
}
