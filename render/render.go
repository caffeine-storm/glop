package render

import (
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/runningwild/glop/render/tls"
)

func MustBeOnRenderThread() {
	if !tls.IsSentinelSet() {
		panic(fmt.Errorf("not on render thread"))
	}
}

// TODO(tmckee): clean: is there a better name for this? RenderContext?
type RenderQueueState interface {
	Shaders() *ShaderBank
}

type RenderJob func(RenderQueueState)

func (j *RenderJob) GetSourceAttribution() string {
	pc := uintptr(reflect.ValueOf(*j).UnsafePointer())
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		panic("couldn't runtime.FuncForPC T_T")
	}
	filename, lineno := fn.FileLine(pc)

	return fmt.Sprintf("%s: %d", filename, lineno)
}

// Accepts jobs for running on a dedicated, internal thread of control. Helpful
// for ensuring certain preconditions needed for calling into OpenGL.
type RenderQueueInterface interface {
	// Attaches a callback to this render queue. It will get called with errors
	// that reach the top of the stack for the inner goroutine.
	AddErrorCallback(func(RenderQueueInterface, error))

	// Eventually runs the given closure on a thread dedicated to OpenGL
	// operations. Jobs are run sequentially in the order queued. Each job is
	// passed a reference to data that must only be used on this
	// RenderQueueInterface's render thread. Callers may assume that the
	// RenderQueueState instance passed to each RenderJob is the same object
	// per-queue.
	Queue(f RenderJob)

	// Blocks until all Queue'd jobs have completed. Note that, if other
	// goroutines are queueing jobs, this will block waiting for them as well!
	Purge()

	// StartProcessing() needs to be called exactly once per queue in order to
	// start running jobs. Queue()ing is allowed before processing has started.
	// Purge()ing is allowed before processing has started with the caveat that
	// even an empty queue will block Purge()ers from continuing until
	// StartProcessing() _is_ called.
	StartProcessing()

	// For debugability, polls the queue's current Purging/NotPurging status.
	IsPurging() bool
}

type renderQueueState struct {
	shaders *ShaderBank
}

var _ RenderQueueState = (*renderQueueState)(nil)

func (state *renderQueueState) Shaders() *ShaderBank {
	return state.shaders
}

type jobWithTiming struct {
	Job      RenderJob
	QueuedAt time.Time
}

type renderQueue struct {
	queueState     *renderQueueState
	workQueue      chan *jobWithTiming
	purge          chan chan bool
	isRunning      atomic.Bool
	isPurging      atomic.Bool
	listener       *JobTimingListener
	errorCallbacks struct {
		fns []func(RenderQueueInterface, error)
		mut sync.Mutex
	}
}

func runAndNotify(request *jobWithTiming, queueState RenderQueueState, listener *JobTimingListener) time.Duration {
	before := time.Now()
	request.Job(queueState)
	after := time.Now()
	delta := after.Sub(before)

	info := &JobTimingInfo{
		RunTime:   delta,
		QueueTime: before.Sub(request.QueuedAt),
	}
	totalTime := info.RunTime + info.QueueTime
	if listener != nil && totalTime >= listener.Threshold {
		listener.OnNotify(info, request.Job.GetSourceAttribution())
	}

	return delta
}

func (q *renderQueue) onError(e error) {
	q.errorCallbacks.mut.Lock()
	defer q.errorCallbacks.mut.Unlock()
	for _, cb := range q.errorCallbacks.fns {
		cb(q, e)
	}
}

func (q *renderQueue) loopWithErrorTracking() {
	for {
		func() {
			defer func() {
				if e := recover(); e != nil {
					if ee, ok := e.(error); ok {
						q.onError(ee)
					} else {
						q.onError(fmt.Errorf("non-error error: %v", e))
					}
				}
			}()

			q.loop()
		}()
	}
}

func (q *renderQueue) loop() {
	for {
		select {
		case job := <-q.workQueue:
			runAndNotify(job, q.queueState, q.listener)
		case ack := <-q.purge:
			func() {
				q.isPurging.Store(true)
				defer q.isPurging.Store(false)

				defer func() {
					if e := recover(); e != nil {
						// put the 'ack' channel back into the sequence-of-purge-requests
						// so that we can continue draining the queue once loop() gets
						// called again. We can't just acknowledge the purge request yet
						// because there could be more work in the workQueue.
						q.purge <- ack

						// Re-raise the error so that RenderQueueInterface error reporting
						// can happen.
						panic(e)
					}
				}()

				for {
					select {
					case job := <-q.workQueue:
						runAndNotify(job, q.queueState, q.listener)
					default:
						// We've just exhausted the workQueue; we can break out of this
						// inner func().
						return
					}
				}
			}()
			ack <- true
		}
	}
}

func MakeQueue(initialization RenderJob) RenderQueueInterface {
	return MakeQueueWithTiming(initialization, nil)
}

func MakeQueueWithTiming(initialization RenderJob, listener *JobTimingListener) RenderQueueInterface {
	result := renderQueue{
		queueState: &renderQueueState{
			shaders: MakeShaderBank(),
		},
		workQueue: make(chan *jobWithTiming, 1000),
		purge:     make(chan chan bool, 16),
		isRunning: atomic.Bool{}, // zero-value is false
		isPurging: atomic.Bool{}, // zero-value is false
		listener:  listener,
	}

	// We're guaranteed that this render job will run first. We can include our
	// own initialization that should happen on the loop's thread.
	result.Queue(func(st RenderQueueState) {
		runtime.LockOSThread()
		tls.SetSentinel()
		initialization(st)
	})
	return &result
}

func (q *renderQueue) AddErrorCallback(fn func(RenderQueueInterface, error)) {
	q.errorCallbacks.mut.Lock()
	defer q.errorCallbacks.mut.Unlock()
	q.errorCallbacks.fns = append(q.errorCallbacks.fns, fn)
}

// TODO(tmckee): inject a GL dependency to given func for testability and to
// keep arbitrary code from calling GL off of the render thread.
func (q *renderQueue) Queue(f RenderJob) {
	q.workQueue <- &jobWithTiming{
		Job:      f,
		QueuedAt: time.Now(),
	}
}

// Waits until all render thread functions have been run
func (q *renderQueue) Purge() {
	if !q.isRunning.Load() {
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
	if !q.isRunning.CompareAndSwap(false, true) {
		panic("must not call 'StartProcessing' twice")
	}
	go q.loopWithErrorTracking()
}

func (q *renderQueue) IsPurging() bool {
	return q.isPurging.Load()
}
