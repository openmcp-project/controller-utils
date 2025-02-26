[![REUSE status](https://api.reuse.software/badge/github.com/openmcp-project/controller-utils)](https://api.reuse.software/info/github.com/openmcp-project/controller-utils)

# controller-utils

## About this project

This repository contains the controller-utils library which provides utility functions for cloud-orchestration projects. It also contains multiple functions and helper structs which are useful for developing k8s controllers and have been found to be copied around from one controller repository to another.

## Requirements and Setup

```bash
$ go get  github.com/openmcp-project/controller-utils@latest
```


## Packages

### Clientconfig

The `pkg/clientconfig` package provides helper functions for creating Kubernetes clients using multiple connection methods. It defines a `Config` struct that encapsulates a Kubernetes API target and supports various authentication methods like kubeconfig file and a Service Account.

#### Noteworthy Functions

- `GetRESTConfig` generates a `*rest.Config` for interacting with the Kubernetes API. It supports using a kubeconfig string, a kubeconfig file path, a secret reference that contains a kubeconfig file or a Service Account. 
- `GetClient` creates a client.Client for managing Kubernetes resources.

### Webhooks
The `pkg/init/webhooks` provides easy tools to deploy webhook configuration and certificates on a target cluster.

#### Noteworthy Functions
- `GenerateCertificate` generates and deploy webhook certificates to the target cluster.
- `Install` deploys mutating/validating webhook configuration on a target cluster.

### CRDs
The `pkg/init/crds` package allows user to deploy CRDs from yaml files to a target cluster. It uses `embed.FS` to provide the files for deployment.


### collections

The `pkg/collections` package contains multiple interfaces for collections, modelled after the Java Collections Framework. The only actual implementation currently contained is a `LinkedList`, which fulfills the `List` and `Queue` interfaces.

The package also contains further packages that contain some auxiliary functions for working with slices and maps in golang, e.g. for filtering.

### controller

The `pkg/controller` package contains useful functions for setting up and running k8s controllers.

#### Noteworthy Functions

- `LoadKubeconfig` creates a REST config for accessing a k8s cluster. It can be used with a path to a kubeconfig file, or a directory containing files for a trust relationship. When called with an empty path, it returns the in-cluster configuration.

### logging


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

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/openmcp-project/controller-utils/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Security / Disclosure
If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/openmcp-project/controller-utils/security/policy) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2025 SAP SE or an SAP affiliate company and controller-utils contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/openmcp-project/controller-utils).
