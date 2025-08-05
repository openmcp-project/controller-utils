# Smart Requeuing of Kubernetes Resources

The `pkg/controller/smartrequeue` package contains the smart requeuing logic that was originally implemented [here](https://github.com/openmcp-project/cluster-provider-kind/tree/v0.0.7/pkg/smartrequeue). It allows to requeue reconcile requests with an increasing backoff, similar to what the controller-runtime does when the `Reconcile` method returns an error.

Use `NewStore` in the constructor of the reconciler. During reconciliation, the store's `For` method can be used to get the entry for the passed-in object, which can then generate the reconcile result:
- `Error` returns an error and lets the controller-runtime handle the backoff
- `Never` does not requeue the object
- `Backoff` requeues the object with an increasing backoff every time it is called on the same object
- `Reset` requeues the object, but resets the duration to its minimal value

There is also an integration into the [status updater](./status.md) for the smart requeuing logic.
