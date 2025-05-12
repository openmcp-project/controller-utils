# Testing

The `pkg/testing` package contains useful functionality to aid with writing tests.

Most notably, the `Environment` provides a `context.Context` containing a logger, a k8s fake client, and has helper methods to allow for easy tests of `Reconcile` methods.

### Noteworthy Functions

- Use `NewEnvironmentBuilder` to construct a simple test environment.
- `Environment` is a simplicity wrapper around `ComplexEnvironment`, which can be used for more complex test scenarios which involve more than one cluster and/or reconciler. Use `NewComplexEnvironmentBuilder` to construct a new `ComplexEnvironment`.

### Examples

Initialize a `Environment` and use it to check if an object is reconciled successfully:
```golang
env := testing.NewEnvironmentBuilder().
  WithFakeClient(nil). // insert your scheme if it differs from default k8s scheme
  WithInitObjectPath("testdata", "test-01").
  WithReconcilerConstructor(func(c client.Client) reconcile.Reconciler {
    return &MyReonciler{
      Client: c,
    }
  }).
  Build()

env.ShouldReconcile(testing.RequestFromStrings("testresource"))
```
