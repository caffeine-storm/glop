package render

import (
	"log"
	"runtime"
)

type RenderQueueInterface interface {
	Queue(f func())
	Purge()
	StartProcessing()
}

type renderQueue struct {
	render_funcs chan func()
	purge        chan bool
	is_running   bool
}

func (q *renderQueue) loop() {
	for {
		select {
		case f := <-q.render_funcs:
			f()
		case <-q.purge:
			for {
				select {
				case f := <-q.render_funcs:
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

func MakeQueue(initialization func ()) RenderQueueInterface {
	result := renderQueue{
		render_funcs: make(chan func(), 1000),
		purge:        make(chan bool),
		is_running:   false,
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

func (q *renderQueue) StartProcessing() {
	if q.is_running {
		panic("must not call 'StartProcessing' twice")
	}
	q.is_running = true
	go q.loop()
}
