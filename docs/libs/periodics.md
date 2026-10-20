# Periodics

The `pkg/periodic` package contains helper types for constructing simple 'periodics', which can be run by controller-runtime.
A 'periodic' is basically a function that is executed periodically at a set interval.

This was initially implemented as a library for garbage collectors, but has been renamed since it can be used for any logic that should be run periodically.

Use the builder constructor to create a new periodic:
```go
	periodic, err := periodic.NewPeriodicBuilder(func(ctx context.Context, p *periodic.PeriodicReadOnly[any], cfg any) {
		// actual logic here
	}).
		WithInterval(1 * time.Hour).
		Build()
```

The builder constructor takes a function as argument which implements the actual logic. It takes the following arguments:
1. The current `context.Context`. Note that controller-runtime does not inject a logger into the context passed into non-controller runnables.
2. A read-only view on the `Periodic`. It can be used to fetch the periodic's name, logger, last run time, and some other values, but all methods that actually modify the periodic are not available.
3. The config, if one or more config functions were set in the builder. See below for a more detailed explanation.

The periodic implements controller-runtime's `manager.Runnable` interface, which means that it can be passed into the manager to be run automatically:
```go
if err := mgr.Add(gc); err != nil {
  return fmt.Errorf("unable to add periodic to manager: %w", err)
}
```

### Configuration

If the logic requires some configuration which can dynamically change, config functions can be used to fetch it:
```go
// assuming the generic type being 'Config'
builder.WithConfigFuncs(
	func(ctx context.Context, cfg Config) (Config, error) {...},
	// variadic, multiple functions can be passed in
)
```
Each config function is called with the current context and the current configuration. It returns a new configuration (which then becomes the 'current' one) and an error.
A few things to note:
- The first config function is fed the zero value of the `Config` type.
- The return value of each config function is fed into the next config function as argument.
- The return value of the last config function is fed into the run logic.
- If any config function returns an error, the error is logged and the run logic is aborted immediately.

