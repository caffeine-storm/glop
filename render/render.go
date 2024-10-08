package render

import (
	"runtime"
)

type RenderQueue struct {
	render_funcs chan func()
	purge        chan bool
}

var (
	defaultQueue RenderQueue
)

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

func (q *RenderQueue) Queue(f func()) {
	q.render_funcs <- f
}

func (q *RenderQueue) Purge() {
	q.purge <- true
	<- q.purge
}

func init() {
	defaultQueue = MakeQueue()
}

// Queues a function to run on the render thread.
// TODO(tmckee): inject a GL dependency to the callback for testability and to
// keep arbitrary code from calling GL off of the render thread.
func Queue(f func()) {
	defaultQueue.Queue(f)
}

// Waits until all render thread functions have been run
func Purge() {
	defaultQueue.Purge()
}
