package render

import (
	"runtime"
	"sync"
)

var (
	render_funcs chan func()
	purge        chan bool
	init_once    sync.Once
)

func init() {
	render_funcs = make(chan func(), 1000)
	purge = make(chan bool)
}

// Queues a function to run on the render thread.
// TODO(tmckee): inject a GL dependency to the callback for testability and to
// keep arbitrary code from calling GL off of the render thread.
func Queue(f func()) {
	render_funcs <- f
}

// Waits until all render thread functions have been run
func Purge() {
	purge <- true
	<-purge
}

func Init() {
	// TODO(tmckee): this approach means we can't use multiple GL contexts from
	// one process. Currently, this is only a problem for testing but we can fix
	// it by spawning a render-goroutine per-context.
	init_once.Do(func() {
		go func() {
			runtime.LockOSThread()
			for {
				select {
				case f := <-render_funcs:
					f()
				case <-purge:
					for {
						select {
						case f := <-render_funcs:
							f()
						default:
							goto purged
						}
					}
				purged:
					purge <- true
				}
			}
		}()
	})
}
