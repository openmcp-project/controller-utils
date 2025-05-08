# Controller Utility Functions

The `pkg/controller` package contains useful functions for setting up and running k8s controllers.

### Noteworthy Functions

- `LoadKubeconfig` creates a REST config for accessing a k8s cluster. It can be used with a path to a kubeconfig file, or a directory containing files for a trust relationship. When called with an empty path, it returns the in-cluster configuration.
  - See also the [`clusters`](#clusters) package, which uses this function internally, but provides some further tooling around it.
- There are some functions useful for working with annotations and labels, e.g. `HasAnnotationWithValue` or `EnsureLabel`.
- There are multiple predefined predicates to help with filtering reconciliation triggers in controllers, e.g. `HasAnnotationPredicate` or `DeletionTimestampChangedPredicate`.
- The `K8sNameHash` function can be used to create a hash that can be used as a name for k8s resources.
