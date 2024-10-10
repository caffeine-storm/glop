package render

import (
	"runtime"
)

type RenderQueue struct {
	render_funcs chan func()
	purge        chan bool
}

func (q *RenderQueue) loop() {
	runtime.LockOSThread()
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

func MakeQueue() RenderQueue {
	result := RenderQueue{
		render_funcs: make(chan func(), 1000),
		purge:        make(chan bool),
	}

	go result.loop()

	return result
}

// TODO(tmckee): inject a GL dependency to given func for testability and to
// keep arbitrary code from calling GL off of the render thread.
func (q *RenderQueue) Queue(f func()) {
	q.render_funcs <- f
}

// Waits until all render thread functions have been run
func (q *RenderQueue) Purge() {
	q.purge <- true
	<- q.purge
}
