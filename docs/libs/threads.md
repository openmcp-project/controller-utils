# Thread Management

The `pkg/threads` package provides a simple thread managing library. It can be used to run go routines in a non-blocking manner and provides the possibility to react if the routine has exited.

The most relevant use-case for this library in the context of k8s controllers is to handle dynamic watches on multiple clusters. To start a watch, that cluster's cache's `Start` method has to be used. Because this method is blocking, it has to be executed in a different go routine, and because it can return an error, a simple `go cache.Start(...)` is not enough, because it would hide the error.

### Noteworthy Functions

- `NewThreadManager` creates a new thread manager.
	- The first argument is a `context.Context` used by the manager itself. Cancelling this context will stop the manager, and if the context contains a `logging.Logger`, the manager will use it for logging.
	- The second argument is an optional function that is executed after any go routine executed with this manager has finished. It is also possible to provide such a function for a specific go routine, instead for all of them, see below.
- Use the `Run` method to start a new go routine.
	- Starting a go routine cancels the context of any running go routine with the same id.
	- This method also takes an optional function to be executed after the actual workload is done.
		- A on-finish function specified here is executed before the on-finish function of the manager is executed.
	- Note that go routines will wait for the thread manager to be started, if that has not yet happened. If the manager has been started, they will be executed immediately.
	- The thread manager will cancel the context that is passed into the workload function when the manager is being stopped. If any long-running commands are being run as part of the workload, it is strongly recommended to listen to the context's `Done` channel.
- Use `Start()` to start the thread manager.
	- If any go routines have been added before this is called, they will be started now. New go routines added afterwards will be started immediately.
	- Calling this multiple times doesn't have any effect, unless the manager has already been stopped, in which case `Start()` will panic.
- There are three ways to stop the thread manager again:
	- Use its `Stop()` method.
		- This is a blocking method that waits for all remaining go routines to finish. Their context is cancelled to notify them of the manager being stopped.
	- Cancel the context that was passed into `NewThreadManager` as the first argument.
	- Send a `SIGTERM` or `SIGINT` signal to the process.
- The `TaskManager`'s `Restart`, `RestartOnError`, and `RestartOnSuccess` methods are pre-defined on-finish functions. They are not meant to be used directly, but instead be used as an argument to `Run`. See the example below.

### Examples

```golang
mgr := threads.NewThreadManager(ctx, nil)
mgr.Start()
// do other stuff
// start a go routine that is restarted automatically if it finishes with an error
mgr.Run(myCtx, "myTask", func(ctx context.Context) error {
	// my task coding
}, mgr.RestartOnError)
// do more other stuff
mgr.Stop()
```

