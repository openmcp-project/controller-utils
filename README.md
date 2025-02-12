# Controller Utilities Library

This repo is meant to contain multiple functions and helper structs which are useful for developing k8s controllers and have been found to be copied around from one controller repository to another.

## Packages

### collections

The `pkg/collections` package contains multiple interfaces for collections, modelled after the Java Collections Framework. The only actual implementation currently contained is a `LinkedList`, which fulfills the `List` and `Queue` interfaces.

The package also contains further packages that contain some auxiliary functions for working with slices and maps in golang, e.g. for filtering.

### controller

The `pkg/controller` package contains useful functions for setting up and running k8s controllers.

#### Noteworthy Functions

- `LoadKubeconfig` creates a REST config for accessing a k8s cluster. It can be used with a path to a kubeconfig file, or a directory containing files for a trust relationship. When called with an empty path, it returns the in-cluster configuration.

### logging

> Simply copied from the Landscaper repo, we should have a look at possible refactorings at some point.

This package contains the logging library from the [Landscaper controller-utils module](https://github.com/gardener/landscaper/tree/master/controller-utils/pkg/logging).

The library provides a wrapper around `logr.Logger`, exposing additional helper functions. The original `logr.Logger` can be retrieved by using the `Logr()` method. Also, it notices when multiple values are added to the logger with the same key - instead of simply overwriting the previous ones (like `logr.Logger` does it), it appends the key with a `_conflict(x)` suffix, where `x` is the number of times this conflict has occurred.

#### Noteworthy Functions

- `GetLogger()` is a singleton-style getter function for a logger.
- There are several `FromContext...` functions for retrieving a logger from a `context.Context` object.
- `InitFlags(...)` can be used to add the configuration flags for this logger to a cobra `FlagSet`.

### testing

This package contains useful functionality to aid with writing tests.

Most notably, the `Environment` provides a `context.Context` containing a logger, a k8s fake client, and has helper methods to allow for easy tests of `Reconcile` methods.

#### Noteworthy Functions

- Use `NewEnvironmentBuilder` to construct a simple test environment.
- `Environment` is a simplicity wrapper around `ComplexEnvironment`, which can be used for more complex test scenarios which involve more than one cluster and/or reconciler. Use `NewComplexEnvironmentBuilder` to construct a new `ComplexEnvironment`.

#### Examples

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
