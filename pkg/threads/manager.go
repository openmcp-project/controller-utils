package threads

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

var sigs chan os.Signal

func init() {
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
}

// WorkFunc is the function that holds the actual workload of a thread.
// The ThreadManager cancels the provided context when being stopped, so the workload should listen to the context's Done channel.
type WorkFunc func(context.Context) error

// OnFinishFunc can be used to react to a thread finishing.
// Note that its context might already be cancelled (if the ThreadManager is being stopped).
type OnFinishFunc func(context.Context, ThreadReturn)

// NewThreadManager creates a new ThreadManager.
// The mgrCtx is used for two purposes:
//  1. If the context is cancelled, the ThreadManager is stopped. Alternatively, its Stop() method can be called.
//  2. If the context contains a logger, it is used for logging.
//
// The threadCtx will be passed to the threads and is cancelled when the ThreadManager is stopped.
// If onFinish is not nil, it will be called whenever a thread finishes. It is called after the thread's own onFinish function, if any.
func NewThreadManager(mgrCtx, threadCtx context.Context, onFinish OnFinishFunc) *ThreadManager {
	threadCtx, stopFunc := context.WithCancel(threadCtx)
	return &ThreadManager{
		threadCtx:   threadCtx,
		returns:     make(chan ThreadReturn, 100),
		onFinish:    onFinish,
		log:         logging.FromContextOrDiscard(mgrCtx),
		runOnStart:  []*Thread{},
		ctxStop:     mgrCtx.Done(),
		stopThreads: stopFunc,
	}
}

type ThreadManager struct {
	lock           sync.Mutex
	threadCtx      context.Context    // context that is passed to threads, is cancelled when the ThreadManager is stopped
	returns        chan ThreadReturn  // channel to receive thread returns
	onFinish       OnFinishFunc       // function to call when a thread finishes
	log            logging.Logger     // logger for the ThreadManager
	runOnStart     []*Thread          // is filled if threads are added before the ThreadManager is started
	ctxStop        <-chan struct{}    // channel to stop the ThreadManager
	stopThreads    context.CancelFunc // function to cancel the context that is passed to threads
	stopSelf       func()             // convenience function to use the internalStop channel
	internalStop   chan struct{}      // if the Stop() method is called, we need to stop the internal main loop by using this channel
	stopped        atomic.Bool        // indicates if the ThreadManager is stopped
	waitForThreads sync.WaitGroup     // used to wait for threads to finish when stopping the ThreadManager
}

// Start starts the ThreadManager.
// This starts a goroutine that listens for thread returns and os signals.
// Calling Start() multiple times is a no-op, unless the ThreadManager has already been stopped, then it panics.
// It is possible to add threads before the ThreadManager is started, but they will only be run after Start() is called.
// Threads added after Start() will be run immediately.
// There are three ways to stop the ThreadManager again:
//  1. Cancel the context passed to the ThreadManager during creation.
//  2. Call the ThreadManager's Stop() method.
//  3. Send a SIGINT or SIGTERM signal to the process.
func (tm *ThreadManager) Start() {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	if tm.stopped.Load() {
		panic("Start called on a stopped ThreadManager")
	}
	if tm.isStarted() {
		tm.log.Debug("Start called, but ThreadManager is already started, nothing to do")
		return
	}
	tm.log.Info("Starting ThreadManager")
	go func() {
		for {
			select {
			case tr, ok := <-tm.returns:
				if !ok {
					// channel has been closed, this means the Stop() method has been called
					return
				}
				if tr.Err != nil {
					tm.log.Error(tr.Err, "Error in thread", "thread", tr.Thread.id)
				}
			case sig := <-sigs:
				tm.log.Info("Received os signal, stopping ThreadManager", "signal", sig)
				tm.stop()
				return
			case <-tm.ctxStop:
				tm.stop()
				return
			}
		}
	}()
	runOnStart := tm.runOnStart
	tm.runOnStart = nil
	if len(runOnStart) > 0 {
		tm.log.Info("Running threads added before ThreadManager was started", "threadCount", len(runOnStart))
		for _, t := range runOnStart {
			tm.run(t)
		}
	}
}

// Stop stops the ThreadManager.
// Calling Stop() multiple times is a no-op.
// It is not possible to start the ThreadManager again after it has been stopped, a new instance must be created.
// Adding threads after the ThreadManager has been stopped is a no-op.
// The ThreadManager is also stopped when the context passed to the ThreadManager during creation is cancelled or when a SIGINT or SIGTERM signal is received.
func (tm *ThreadManager) Stop() {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	if !tm.isStarted() {
		panic("Stop called on a ThreadManager that has not been started yet")
	}
	tm.stop()
}

func (tm *ThreadManager) stop() {
	if tm.stopped.Load() {
		tm.log.Debug("Stop called, but ThreadManager is already stopped, nothing to do")
		return
	}
	tm.log.Info("Stopping ThreadManager, waiting for remaining threads to finish")
	tm.stopped.Store(true)
	tm.stopThreads()

	tm.waitForThreads.Wait()
	close(tm.returns)
	tm.log.Info("ThreadManager stopped")
	return
}

// Run gives a new thread to run to the ThreadManager.
// id is only used for logging and debugging purposes.
// work is the actual workload of the thread.
// onFinish can be used to react to the thread having finished.
// Note that there are some pre-defined functions that can be used as onFinish functions, e.g. the ThreadManager's Restart method.
func (tm *ThreadManager) Run(id string, work func(context.Context) error, onFinish OnFinishFunc) {
	tm.RunThread(NewThread(id, work, onFinish))
}

// RunThread is the same as Run, but takes a Thread struct instead of the individual parameters.
func (tm *ThreadManager) RunThread(t Thread) {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.run(&t)
}

func (tm *ThreadManager) run(t *Thread) {
	if t == nil {
		tm.log.Error(nil, "run(t *Thread) called with nil Thread, this should never happen")
		return
	}
	if tm.stopped.Load() {
		tm.log.Info("Skipping thread run because ThreadManager is already stopped", "thread", t.id)
		return
	}
	if !tm.isStarted() {
		tm.runOnStart = append(tm.runOnStart, t)
		tm.log.Debug("ThreadManager has not been started yet, enqueuing thread to run on start", "thread", t.ID())
		return
	}
	tm.log.Debug("Running thread", "thread", t.id)
	tm.waitForThreads.Add(1)
	go func() {
		defer tm.waitForThreads.Done()
		var err error
		if t.work != nil {
			err = t.work(tm.threadCtx)
		} else {
			tm.log.Debug("Thread has no work function", "thread", t.id)
		}
		tr := NewThreadReturn(t, err)
		if t.onFinish != nil {
			tm.log.Debug("Calling the thread's onFinish function", "thread", t.id)
			t.onFinish(tm.threadCtx, tr)
		}
		if tm.onFinish != nil {
			tm.log.Debug("Calling the thread manager's onFinish function", "thread", tr.Thread.id)
			tm.onFinish(tm.threadCtx, tr)
		}
		tm.returns <- tr
		tm.log.Debug("Thread finished", "thread", t.id)
	}()
}

func (tm *ThreadManager) isStarted() bool {
	return tm.runOnStart == nil
}

// IsStarted returns true if the ThreadManager has been started.
// Note that this will return true if the ThreadManager has been started at some point, even if it has been stopped by now.
func (tm *ThreadManager) IsStarted() bool {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	return tm.isStarted()
}

// IsStopped returns true if the ThreadManager has been stopped.
// Note that this will return false if the ThreadManager has not been started yet.
func (tm *ThreadManager) IsStopped() bool {
	return tm.stopped.Load()
}

// IsRunning returns true if the ThreadManager is currently running,
// meaning it has been started and not yet been stopped.
// This is a convenience function that is equivalent to calling IsStarted() && !IsStopped().
func (tm *ThreadManager) IsRunning() bool {
	return tm.IsStarted() && !tm.IsStopped()
}

var _ OnFinishFunc = (*ThreadManager)(nil).Restart

// Restart is a pre-defined onFinish function that can be used to restart a thread after it has finished.
// This method is not meant to be called directly, instead pass it to the ThreadManager's Run method as the onFinish parameter:
//
//	tm.Run("myThread", myWorkFunc, tm.Restart)
func (tm *ThreadManager) Restart(_ context.Context, tr ThreadReturn) {
	if tm.stopped.Load() {
		return
	}
	tm.RunThread(*tr.Thread)
}

var _ OnFinishFunc = (*ThreadManager)(nil).RestartOnError

// RestartOnError is a pre-defined onFinish function that can be used to restart a thread after it has finished, if it finished with an error.
// It is the opposite of RestartOnSuccess.
// This method is not meant to be called directly, instead pass it to the ThreadManager's Run method as the onFinish parameter:
//
//	tm.Run("myThread", myWorkFunc, tm.RestartOnError)
func (tm *ThreadManager) RestartOnError(_ context.Context, tr ThreadReturn) {
	if tr.Err != nil {
		tm.Restart(tm.threadCtx, tr)
	}
}

var _ OnFinishFunc = (*ThreadManager)(nil).RestartOnSuccess

// RestartOnSuccess is a pre-defined onFinish function that can be used to restart a thread after it has finished, if it didn't throw an error.
// It is the opposite of RestartOnError.
// This method is not meant to be called directly, instead pass it to the ThreadManager's Run method as the onFinish parameter:
//
//	tm.Run("myThread", myWorkFunc, tm.RestartOnSuccess)
func (tm *ThreadManager) RestartOnSuccess(_ context.Context, tr ThreadReturn) {
	if tr.Err == nil {
		tm.Restart(tm.threadCtx, tr)
	}
}

// NewThread creates a new thread with the given id, work function and onFinish function.
// It is usually not required to call this function directly, instead use the ThreadManager's Run method.
// The Thread's fields are considered immutable after creation.
func NewThread(id string, work WorkFunc, onFinish OnFinishFunc) Thread {
	return Thread{
		id:       id,
		work:     work,
		onFinish: onFinish,
	}
}

// Thread represents a thread that can be run by the ThreadManager.
type Thread struct {
	id       string
	work     WorkFunc
	onFinish OnFinishFunc
}

// ID returns the id of the thread.
func (t *Thread) ID() string {
	return t.id
}

// WorkFunc returns the workload function of the thread.
func (t *Thread) WorkFunc() WorkFunc {
	return t.work
}

// OnFinishFunc returns the onFinish function of the thread.
func (t *Thread) OnFinishFunc() OnFinishFunc {
	return t.onFinish
}

// NewThreadReturn constructs a new ThreadReturn object.
// This is used by the ThreadManager internally and it should rarely be necessary to call this function directly.
func NewThreadReturn(thread *Thread, err error) ThreadReturn {
	return ThreadReturn{
		Err:    err,
		Thread: thread,
	}
}

// ThreadReturn represents the result of a thread's execution.
// It contains a reference to the thread and an error, if any occurred.
type ThreadReturn struct {
	Err    error
	Thread *Thread
}
