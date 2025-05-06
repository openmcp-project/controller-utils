[![REUSE status](https://api.reuse.software/badge/github.com/openmcp-project/controller-utils)](https://api.reuse.software/info/github.com/openmcp-project/controller-utils)

# controller-utils

## About this project

This repository contains the controller-utils library which provides utility functions for Open Managed Control Planes projects. It also contains multiple functions and helper structs which are useful for developing k8s controllers and have been found to be copied around from one controller repository to another.

## Requirements and Setup

```bash
$ go get github.com/openmcp-project/controller-utils@latest
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

## clusters

The `pkg/clusters` package helps with loading kubeconfigs and creating clients for multiple clusters.
```go
foo := clusters.New("foo") // initializes a new cluster with id 'foo'
foo.RegisterConfigPathFlag(cmd.Flags()) // adds a '--foo-cluster' flag to the flag set for passing in a kubeconfig path
foo.InitializeRESTConfig() // loads the kubeconfig using the 'LoadKubeconfig' function from the 'controller' package
foo.InitializeClient(myScheme) // initializes the 'Client' and 'Cluster' interfaces from the controller-runtime
```
You can then use the different getter methods for working with the cluster.

### conditions

The `pkg/conditions` package helps with managing condition lists.

The managed condition implementation must satisfy the `Condition[T comparable]` interface:
```go
type Condition[T comparable] interface {
	GetType() string
	SetType(conType string)
	GetStatus() T
	SetStatus(status T)
	GetLastTransitionTime() time.Time
	SetLastTransitionTime(timestamp time.Time)
	GetReason() string
	SetReason(reason string)
	GetMessage() string
	SetMessage(message string)
}
```

To manage conditions, use the `ConditionUpdater` function and pass in a constructor function for your condition implementation and the old list of conditions. The bool argument determines whether old conditions that are not updated remain in the returned list (`false`) or are removed, so that the returned list contains only the conditions that were touched (`true`).

```go
updater := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, oldCons, false)
```

Note that the `ConditionUpdater` stores the current time upon initialization and will set each updated condition's timestamp to this value, if the status of that condition changed as a result of the update. To use a different timestamp, manually overwrite the `Now` field of the updater.

Use `UpdateCondition` or `UpdateConditionFromTemplate` to update a condition:
```go
updater.UpdateCondition("myCondition", true, "newReason", "newMessage")
```

If all conditions are updated, use the `Conditions` method to generate the new list of conditions. The originally passed in list of conditions is not modified by the updater.
The second return value is `true` if the updated list of conditions differs from the original one.
```go
updatedCons, changed := updater.Conditions()
```

For simplicity, all commands can be chained:
```go
updatedCons, changed := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, oldCons, false).UpdateCondition("myCondition", true, "newReason", "newMessage").Conditions()
```

### controller

The `pkg/controller` package contains useful functions for setting up and running k8s controllers.

#### Noteworthy Functions

- `LoadKubeconfig` creates a REST config for accessing a k8s cluster. It can be used with a path to a kubeconfig file, or a directory containing files for a trust relationship. When called with an empty path, it returns the in-cluster configuration.
	- See also the [`clusters`](#clusters) package, which uses this function internally, but provides some further tooling around it.
- There are some functions useful for working with annotations and labels, e.g. `HasAnnotationWithValue` or `EnsureLabel`.
- There are multiple predefined predicates to help with filtering reconciliation triggers in controllers, e.g. `HasAnnotationPredicate` or `DeletionTimestampChangedPredicate`.
- The `K8sNameHash` function can be used to create a hash that can be used as a name for k8s resources.

#### Status Updater

The status updater gets its own section, because it requires a slightly longer explanation. The idea of it is that many of our resources use a status similar to this:
```go
type MyStatus struct {
	// ObservedGeneration is the generation of this resource that was last reconciled by the controller.
	ObservedGeneration int64 `json:"observedGeneration"`

	// LastReconcileTime is the time when the resource was last reconciled by the controller.
	LastReconcileTime metav1.Time `json:"lastReconcileTime"`

	// Phase is the overall phase of the resource.
	Phase string

	// Reason is expected to contain a CamelCased string that provides further information in a machine-readable format.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message contains further details in a human-readable format.
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions contains the conditions.
	// +optional
	Conditions []MyCondition `json:"conditions,omitempty"`
}
```

The logic for most of these fields is very similar across all of our controllers: `ObservedGeneration` and `LastReconcileTime` should always be updated, `Phase` is usually computed based on the conditions or on whether an error occurred, `Reason`, `Message` and `Conditions` are generated during reconciliation.

To reduce redundant coding and ensure a similar behavior in all controllers, the _status updater_ can be used to update the status. A full example could look something like this:
```go
import (
	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
	v1alpha1 // resource API package
)

// optional, using a type alias removes the need to specify the type arguments every time
type ReconcileResult = ctrlutils.ReconcileResult[*v1alpha1.MyResource, v1alpha1.ConditionStatus]

// this is the method called by the controller-runtime
func (r *GardenerMyResourceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rr := r.reconcile(ctx, req)
	// status update
	return ctrlutils.NewStatusUpdaterBuilder[*v1alpha1.MyResource, v1alpha1.MyResourcePhase, v1alpha1.ConditionStatus]().
		WithPhaseUpdateFunc(func(obj *v1alpha1.MyResource, rr ctrlutils.ReconcileResult[*v1alpha1.MyResource, v1alpha1.ConditionStatus]) (v1alpha1.MyResourcePhase, error) {
			if rr.ReconcileError != nil {
				return v1alpha1.PROVIDER_CONFIG_PHASE_FAILED, nil
			}
			if len(rr.Conditions) > 0 {
				for _, con := range rr.Conditions {
					if con.GetStatus() != v1alpha1.CONDITION_TRUE {
						return v1alpha1.PROVIDER_CONFIG_PHASE_FAILED, nil
					}
				}
			}
			return v1alpha1.PROVIDER_CONFIG_PHASE_SUCCEEDED, nil
		}).
		WithConditionUpdater(func() conditions.Condition[v1alpha1.ConditionStatus] {
			return &v1alpha1.Condition{}
		}, true).
		Build().
		UpdateStatus(ctx, r.PlatformCluster.Client(), rr)
}

func (r *GardenerProviderConfigReconciler) reconcile(ctx context.Context, req reconcile.Request) ReconcileResult {
	// actual reconcile logic here
}
```

Some information regarding the example:
- `v1alpha1.MyResource` is the resource type being reconciled in this example.
- `v1alpha1.MyResourcePhase` is the type of the `Phase` field used in the status of `MyResource`.
	- It must be a string-like type, e.g. `type MyResourcePhase string`.
	- If the resource status doesn't have a `Phase` or updating it is not desired, simply set this type argument to `string`.
- `v1alpha1.ConditionStatus` is the type of the `Status` field within the conditions. It must be `comparable`.
	- Usually, this will either be a boolean or a string-like type.
	- If the resource status doesn't have conditions or updating them is not desired, simply set this type argument to `bool`.
- The conditions must be a list of a type `T`, where either `T` or `*T` implements the `conditions.Condition[ConType]` interface.
	- `ConType` is `v1alpha1.ConditionStatus` in this example.


**How to use the status updater**

It is recommended to move the actual reconciliation logic into a helper function (`reconcile` in the example). This makes it easier to ensure that the status updater is always called, no matter where the reconciliation exits, e.g. due to an error. This helper function should then return the `ReconcileResult` required by the status updater.

First, initialize a new `StatusUpdaterBuilder`:
```go
ctrlutils.NewStatusUpdaterBuilder[*v1alpha1.MyResource, v1alpha1.MyResourcePhase, v1alpha1.ConditionStatus]()
```
It takes the type of the reconciled resource, the type of its `Phase` attribute and the type of the `Status` attribute of its conditions as type arguments.

If you want to update the phase, you have to pass in a function that computes the new phase based on the the current state of the object and the returned reconcile result. Note that the function just has to return the phase, not to set it in the object. Failing to provide this function causes the updater to use a dummy implementation that sets the phase to the empty string.
```go
WithPhaseUpdateFunc(func(obj *v1alpha1.MyResource, rr ctrlutils.ReconcileResult[*v1alpha1.MyResource, v1alpha1.ConditionStatus]) (v1alpha1.MyResourcePhase, error) {
	if rr.ReconcileError != nil {
		return v1alpha1.PROVIDER_CONFIG_PHASE_FAILED, nil
	}
	if len(rr.Conditions) > 0 {
		for _, con := range rr.Conditions {
			if con.GetStatus() != v1alpha1.CONDITION_TRUE {
				return v1alpha1.PROVIDER_CONFIG_PHASE_FAILED, nil
			}
		}
	}
	return v1alpha1.PROVIDER_CONFIG_PHASE_SUCCEEDED, nil
})
```

If the conditions should be updated, the `WithConditionUpdater` method must be called. Similarly to the condition updater from the `conditions` package - which is used internally - it requires a constructor function that returns a new, empty instance of the controller-specific `conditions.Condition` implementation. The second argument specifies whether existing conditions that are not part of the updated conditions in the `ReconcileResult` should be removed or kept.

You can then `Build()` the status updater and run `UpdateStatus()` to do the actual status update. The return values of this method are meant to be returned by the `Reconcile` function.

**Some more details**

- The status updater uses reflection to modifiy the status' fields. This requires it to know the field names (the ones in go, not the ones in the YAML representation). By default, it expects them to be `Status` for the status itself and `Phase`, `ObservedGeneration`, `LastReconcileTime`, `Reason`, `Message`, and `Conditions` for the respective fields within the status.
	- To use a different field name, overwrite it by using either `WithFieldOverride` or `WithFieldOverrides`.
	- If any of the fields is not contained top-level in the status but within a nested struct, the names of these fields must be prefixed with the names of the corresponding structs, separated by a `.`. The `WithNestedStruct` method can be used to set such a prefix quickly for one or more fields.
	- To disable the update of a specific field altogether, set its name to the empty string. This can be done via the aforementioned `WithFieldOverride`/`WithFieldOverrides` methods, or simpler via `WithoutFields`.
		- Doing this for the status field itself turns the status update into a no-op.
	- The package contains constants with the field keys that are required by most of these methods. `STATUS_FIELD` refers to the `Status` field itself, the other field keys are prefixed with `STATUS_FIELD_`.
		- The `AllStatusFields()` function returns a list containing all status field keys, _except the one for the status field itself_, for convenience.
- The `WithCustomUpdateFunc` method can be used to inject a function that performs custom logic on the resource's status. Note that while the function gets the complete object as an argument, only changes to its status will be updated by the status updater.

**The ReconcileResult**

The `ReconcileResult` that is passed into the status updater is expected to contain a representation of what happened during the reconciliation. Its fields influence what the updated status will look like.

- `Result` contains the `reconcile.Result` that is expected as a return value from the `reconcile.Reconciler` interface's `Reconcile` method. It is not modified in any way and simply passed through. It does not affect any of the status' fields.
- `ReconcileError` contains any error(s) that occurred during the actual reconciliation. It must be of type `errors.ReasonableError`. This will also be the second return argument from the `UpdateStatus()` method.
- `Reason` and `Message` can be set to set the status' corresponding fields.
	- If either one is nil, but `ReconcileError` is not, it will be filled with a value derived from the error.
- `Conditions` contains the updated conditions. Depending on with which arguments `WithConditionUpdater` was called, the existing conditions will be either updated with these ones (keeping the other ones), or be replaced by them.
- `Object` contains the object to be updated.
	- If `Object` is nil, no status update will be performed.
- `OldObject` holds the version of the object that will be used as a base for constructing the patch during the status update.
	- If this is nil, `Object` will be used instead.
	- If this is non-nil, it must not point to the same instance as `Object` - use the `DeepCopy()` function to create a different instance.
	- All changes to `Object`'s status that are not part to `OldObject`'s status will be included in the patch during the status update. This can be used to inject custom changes to the status into the status update (in addition to the `WithCustomUpdateFunc` mentioned above).

### errors

The `errors` package contains the `ReasonableError` type, which combines a normal `error` with a reason string. This is useful for errors that happen during reconciliation for updating the resource's status with a reason for the error later on.

#### Noteworthy Functions:

- `WithReason(...)` can be used to wrap a standard error together with a reason into a `ReasonableError`.
- `Errorf(...)` can be used to wrap an existing `ReasonableError` together with a new error, similarly to how `fmt.Errorf(...)` does it for standard errors.
- `NewReasonableErrorList(...)` or `Join(...)` can be used to work with lists of errors. `Aggregate()` turns them into a single error again.

### logging

This package contains the logging library from the [Landscaper controller-utils module](https://github.com/gardener/landscaper/tree/master/controller-utils/pkg/logging).

The library provides a wrapper around `logr.Logger`, exposing additional helper functions. The original `logr.Logger` can be retrieved by using the `Logr()` method. Also, it notices when multiple values are added to the logger with the same key - instead of simply overwriting the previous ones (like `logr.Logger` does it), it appends the key with a `_conflict(x)` suffix, where `x` is the number of times this conflict has occurred.

#### Noteworthy Functions

- `GetLogger()` is a singleton-style getter function for a logger.
- There are several `FromContext...` functions for retrieving a logger from a `context.Context` object.
- `InitFlags(...)` can be used to add the configuration flags for this logger to a cobra `FlagSet`.

### threads

The `threads` package provides a simple thread managing library. It can be used to run go routines in a non-blocking manner and provides the possibility to react if the routine has exited.

The most relevant use-case for this library in the context of k8s controllers is to handle dynamic watches on multiple clusters. To start a watch, that cluster's cache's `Start` method has to be used. Because this method is blocking, it has to be executed in a different go routine, and because it can return an error, a simple `go cache.Start(...)` is not enough, because it would hide the error.

#### Noteworthy Functions

- `NewThreadManager` creates a new thread manager.
	- The first argument is a `context.Context` used by the manager itself. Cancelling this context will stop the manager, and if the context contains a `logging.Logger`, the manager will use it for logging.
	- The second argument is a `context.Context` that is used as a base context for the executed go routines.
	- The third argument is an optional function that is executed after any go routine executed with this manager has finished. It is also possible to provide such a function for a specific go routine, instead for all of them, see below.
- Use the `Run` method to start a new go routine.
	- This method also takes an optional function to be executed after the actual workload is done.
		- A on-finish function specified here is executed before the on-finish function of the manager is executed.
	- Note that go routines will wait for the thread manager to be started, if that has not yet happened. If the manager has been started, they will be executed immediately.
	- The thread manager will cancel the context that is passed into the workload function when the manager is being stopped. If any long-running commands are being run as part of the workload, it is recommended to listen to the context's `Done` channel.
- Use `Start()` to start the thread manager.
	- If any go routines have been added before this is called, they will be started now. New go routines added afterwards will be started immediately.
	- Calling this multiple times doesn't have any effect, unless the manager has already been stopped, in which case `Start()` will panic.
- There are three ways to stop the thread manager again:
	- Use its `Stop()` method.
		- This is a blocking method that waits for all remaining go routines to finish. Their context is cancelled to notify them of the manager being stopped.
	- Cancel the context that was passed into `NewThreadManager` as the first argument.
	- Send a `SIGTERM` or `SIGINT` signal to the process.
- The `TaskManager`'s `Restart`, `RestartOnError`, and `RestartOnSuccess` methods are pre-defined on-finish functions. They are not meant to be used directly, but instead be used as an argument to `Run`. See the example below.

#### Examples

```golang
mgr := threads.NewThreadManager(ctx1, ctx2, nil)
mgr.Start()
// do other stuff
// start a go routine that is restarted automatically if it finishes with an error
mgr.Run("myTask", func(ctx context.Context) error {
	// my task coding
}, mgr.RestartOnError)
// do more other stuff
mgr.Stop()
```

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

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/openmcp-project/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2025 SAP SE or an SAP affiliate company and controller-utils contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/openmcp-project/controller-utils).
