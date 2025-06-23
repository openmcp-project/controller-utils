# Retrying k8s Operations

The `pkg/retry` package contains a `Client` that wraps a `client.Client` while implementing the interface itself and retries any failed (= the returned error is not `nil`) operation.
Methods that don't return an error are simply forwarded to the internal client.

In addition to the `client.Client` interface's methods, the `retry.Client` also has `CreateOrUpdate` and `CreateOrPatch` methods, which use the corresponding controller-runtime implementations internally.

The default retry parameters are:
- retry every 100 milliseconds
- don't increase retry interval
- no maximum number of attempts
- timeout after 1 second

The `retry.Client` struct has builder-style methods to configure the parameters:
```golang
retryingClient := retry.NewRetryingClient(myClient).
  WithTimeout(10 * time.Second). // try for at max 10 seconds
  WithInterval(500 * time.Millisecond). // try every 500 milliseconds, but ...
  WithBackoffMultiplier(2.0) // ... double the interval after each retry
```

For convenience, the `clusters.Cluster` type can return a retrying client for its internal client:
```golang
// cluster is of type *clusters.Cluster
err := cluster.Retry().WithMaxAttempts(3).Get(...)
```
