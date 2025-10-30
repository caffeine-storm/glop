package render

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render/tls"
)

func IsOnRenderThread() bool {
	return tls.IsSentinelSet()
}

func MustBeOnRenderThread() {
	if !IsOnRenderThread() {
		panic(fmt.Errorf("not on render thread but should be"))
	}
}

func MustNotBeOnRenderThread() {
	if IsOnRenderThread() {
		panic(fmt.Errorf("on render thread but shouldn't be"))
	}
}

// TODO(tmckee): clean: is there a better name for this? RenderContext?
type RenderQueueState interface {
	Context() context.Context
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
	// The given callback will be invoked when queued jobs panic over an error
	// value. Note that such panics do _not_ cause the Queue to enter a 'defunct'
	// state; other jobs can still be queued and run.
	AddErrorCallback(func(RenderQueueInterface, error))

	// Eventually runs the given closure on a thread dedicated to OpenGL
	// operations. Jobs are run sequentially in the order queued. Each job is
	// passed a reference to data that must only be used on this
	// RenderQueueInterface's render thread. Callers may assume that the
	// RenderQueueState instance passed to each RenderJob is the same object
	// per-queue. Caveat: if the queue is in a 'defunct' state, calls to Queue()
	// will succeed but the jobs may not run.
	Queue(f RenderJob)

	// Blocks until all Queue'd jobs have completed. Note that, if other
	// goroutines are queueing jobs, this will block waiting for them as well!
	// If the queue enters a 'defunct' state (by calling StopProcessing or if the
	// underlying goroutine exits), current and subsequent calls to Purge() will
	// panic with a 'render.QueueShutdownError'.
	Purge()

	// StartProcessing() needs to be called exactly once per queue in order to
	// start running jobs. Queue()ing is allowed before processing has started.
	// Purge()ing is allowed before processing has started with the caveat that
	// even an empty queue will block Purge()ers from continuing until
	// StartProcessing() _is_ called.
	StartProcessing()

	// StopProcessing() can be called to stop processing jobs on the render
	// queue. It can be called before or after StartProcessing(). It will not
	// interrupt a running job but will pre-empt any other scheduled work. Any
	// subsequent or currently blocked calls to Purge() will panic with a
	// 'render.QueueShutdownError'. Calling StopProcessing() on a defunct queue
	// is a no-op.
	StopProcessing()

	// Returns true iff the queue is in a 'defunct' state.
	IsDefunct() bool

	// For debugability, polls the queue's current Purging/NotPurging status.
	IsPurging() bool
}

type RenderQueueWithLoggerInterface interface {
	RenderQueueInterface
	SetLogger(glog.Logger)
}

type renderQueueState struct {
	ctx     context.Context
	shaders *ShaderBank
}

var _ RenderQueueState = (*renderQueueState)(nil)

func (state *renderQueueState) Context() context.Context {
	return state.ctx
}

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
	isDefunct      atomic.Bool
	listener       *JobTimingListener
	errorCallbacks struct {
		fns []func(RenderQueueInterface, error)
		mut sync.Mutex
	}
	logger glog.Logger
}

func (q *renderQueue) runAndNotify(request *jobWithTiming, ack chan bool) time.Duration {
	before := time.Now()
	after := time.Time{}
	defer func() {
		if !after.IsZero() {
			// This is the 'happy path'; the job succeeded.
			return
		}

		// Let panics panic but, if e is nil, we know we're running
		// runtime.Goeexit.
		if e := recover(); e != nil {
			panic(e)
		}

		// No matter what we do, the current goroutine is going away. That means no
		// more jobs will run and any GL state will be lost. Set things up so that
		// client code won't hang forever.
		if ack != nil {
			close(ack)
		}

		cancelPurgeRequests(q.purge)

		// Set this flag last so that clients either see 'already defunct' or they
		// try to read/write closed channels.
		q.isDefunct.Store(true)
	}()

	LogAndClearGlErrors(q.logger)
	request.Job(q.queueState)
	LogAndClearGlErrorsWithAttribution(q.logger, request.Job)

	after = time.Now()
	delta := after.Sub(before)

	info := &JobTimingInfo{
		RunTime:   delta,
		QueueTime: before.Sub(request.QueuedAt),
	}
	totalTime := info.RunTime + info.QueueTime
	if q.listener != nil && totalTime >= q.listener.Threshold {
		q.listener.OnNotify(info, request.Job.GetSourceAttribution())
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

func cancelPurgeRequests(reqs chan chan bool) {
	close(reqs)
	for {
		ackChannel, ok := <-reqs
		if !ok {
			return
		}
		close(ackChannel)
	}
}

func (q *renderQueue) loopWithErrorTracking() {
	for {
		if q.isDefunct.Load() {
			// No more work to do but we need to unblock any Purge() callers.
			cancelPurgeRequests(q.purge)
			return
		}
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
			q.runAndNotify(job, nil)
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
						q.runAndNotify(job, ack)
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
	return MakeQueueWithTimingAndLogger(initialization, nil, glog.WarningLogger())
}

func MakeQueueWithLogger(initialization RenderJob, logger glog.Logger) RenderQueueInterface {
	return MakeQueueWithTimingAndLogger(initialization, nil, logger)
}

func MakeQueueWithTiming(initialization RenderJob, listener *JobTimingListener) RenderQueueInterface {
	return MakeQueueWithTimingAndLogger(initialization, listener, glog.WarningLogger())
}

func MakeQueueWithTimingAndLogger(initialization RenderJob, listener *JobTimingListener, logger glog.Logger) RenderQueueInterface {
	if logger == nil {
		logger = glog.VoidLogger()
	}

	result := renderQueue{
		queueState: &renderQueueState{
			ctx:     context.Background(),
			shaders: MakeShaderBank(),
		},
		workQueue: make(chan *jobWithTiming, 1000),
		purge:     make(chan chan bool, 16),
		isRunning: atomic.Bool{}, // zero-value is false
		isPurging: atomic.Bool{}, // zero-value is false
		isDefunct: atomic.Bool{}, // zero-value is false
		listener:  listener,
		logger:    logger,
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

type queueShutdownError struct{}

func (*queueShutdownError) Error() string {
	return "queue has shutdown"
}

var (
	_                  error = (*queueShutdownError)(nil)
	QueueShutdownError       = &queueShutdownError{}
)

func (q *renderQueue) AddErrorCallback(fn func(RenderQueueInterface, error)) {
	q.errorCallbacks.mut.Lock()
	defer q.errorCallbacks.mut.Unlock()
	q.errorCallbacks.fns = append(q.errorCallbacks.fns, fn)
}

// TODO(tmckee): inject a GL dependency to given func for testability and to
// keep arbitrary code from calling GL off of the render thread.
func (q *renderQueue) Queue(f RenderJob) {
	if q.isDefunct.Load() {
		return
	}
	q.workQueue <- &jobWithTiming{
		Job:      f,
		QueuedAt: time.Now(),
	}
}

// Waits until all render thread functions have been run
func (q *renderQueue) Purge() {
	if q.isDefunct.Load() {
		panic(QueueShutdownError)
	}
	if !q.isRunning.Load() {
		slog.Warn("render.RenderQueue.Purge called on non-started queue")
	}
	ack := make(chan bool)
	q.purge <- ack
	_, ok := <-ack
	if !ok {
		panic(QueueShutdownError)
	}
}

func (q *renderQueue) StartProcessing() {
	if !q.isRunning.CompareAndSwap(false, true) {
		panic("must not call 'StartProcessing' twice")
	}
	go func() {
		pprof.Do(context.Background(), pprof.Labels("glop-threadid", "render-thread"), func(ctx context.Context) {
			q.queueState.ctx = ctx
			q.loopWithErrorTracking()
		})
	}()
}

func (q *renderQueue) StopProcessing() {
	q.isDefunct.Store(true)
}

func (q *renderQueue) IsDefunct() bool {
	return q.isDefunct.Load()
}

func (q *renderQueue) IsPurging() bool {
	return q.isPurging.Load()
}

func (q *renderQueue) SetLogger(logger glog.Logger) {
	q.logger = logger
}
