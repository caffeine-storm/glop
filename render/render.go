package render

import (
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
)

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
	queue_state *renderQueueState
	work_queue  chan *jobWithTiming
	purge       chan chan bool
	is_running  bool
	is_purging  atomic.Bool
	listener    *JobTimingListener
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

func (q *renderQueue) loop() {
	defer close(q.purge)
	for {
		select {
		case job := <-q.work_queue:
			runAndNotify(job, q.queue_state, q.listener)
		case ack := <-q.purge:
			defer close(ack)
			q.is_purging.Store(true)
			for {
				select {
				case job := <-q.work_queue:
					runAndNotify(job, q.queue_state, q.listener)
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
	return MakeQueueWithTiming(initialization, nil)
}

func MakeQueueWithTiming(initialization RenderJob, listener *JobTimingListener) RenderQueueInterface {
	result := renderQueue{
		queue_state: &renderQueueState{
			shaders: MakeShaderBank(),
		},
		work_queue: make(chan *jobWithTiming, 1000),
		purge:      make(chan chan bool),
		is_running: false,
		is_purging: atomic.Bool{}, // zero-value is false
		listener:   listener,
	}

	// We're guaranteed that this render job will run first. We can include our
	// own initialization that should happen on the loop's thread.
	result.Queue(func(st RenderQueueState) {
		runtime.LockOSThread()
		initialization(st)
	})
	return &result
}

// TODO(tmckee): inject a GL dependency to given func for testability and to
// keep arbitrary code from calling GL off of the render thread.
func (q *renderQueue) Queue(f RenderJob) {
	q.work_queue <- &jobWithTiming{
		Job:      f,
		QueuedAt: time.Now(),
	}
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
