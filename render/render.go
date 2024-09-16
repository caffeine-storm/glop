package render

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	render_funcs chan func()
	purge        chan bool
	init_once    sync.Once
)

func init() {
	fmt.Println("render/render.go: init()ing")
	render_funcs = make(chan func(), 1000)
	purge = make(chan bool)
}

// Queues a function to run on the render thread
func Queue(f func()) {
	fmt.Println("render/render.go: Queue()ing")
	render_funcs <- f
}

// Waits until all render thread functions have been run
func Purge() {
	fmt.Println("render/render.go: Purge()ing")
	purge <- true
	<-purge
}

func Init() {
	fmt.Println("render/render.go: Init()ing")
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
