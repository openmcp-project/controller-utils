# Smart Requeuing of Kubernetes Resources

The `pkg/controller/smartrequeue` package provides flexible requeue timing for Kubernetes controllers. It tracks requeue state **per-object** and lets you make decisions based on what's actually happening with the resource, rather than applying exponential backoff globally on errors.

**The Problem**: controller-runtime's default approach treats all errors the same and applies exponential backoff universally, without accounting for the actual resource state.

**The Solution**: SmartRequeue lets you:
- Track requeue state per-object (not just globally)
- Make decisions based on resource state (phase, readiness, conditions, etc.)
- Requeue with appropriate delays for each scenario

## When to Use SmartRequeue

- **Resource provisioning**: Exponential backoff while waiting for external systems
- **Configuration synchronization**: Reset delays when changes are detected
- **Status monitoring**: Periodic checks without aggressive retries
- **Error recovery**: Custom backoff independent of error types

## Decision Tree

```
┌─ Reconciliation Executes ─┐
│
├─ Error occurred?
│  └─→ ReturnError(err)  [Reset backoff, let controller-runtime handle retry]
│
├─ Resource progressing?
│  └─→ IsProgressing()   [Reset interval to min, check again soon]
│
├─ Resource stable?
│  └─→ IsStable()        [Periodic monitoring, increase interval]
│
└─ Fully reconciled?
   └─→ StopRequeue()     [Stop requeuing, delete entry]
```

## Setup

Create a `Store` once in your reconciler constructor and keep it across reconciliations:

```go
store := smartrequeue.NewStore(
    5*time.Second,   // minInterval: shortest requeue delay
    10*time.Minute,  // maxInterval: longest requeue delay
    2.0,             // multiplier: exponential growth factor
)
```

### Configuration Parameters

| Parameter | Purpose |
|-----------|---------|
| `minInterval` | Initial requeue delay; used when backoff resets |
| `maxInterval` | Maximum cap on requeue delay |
| `multiplier` | Exponential growth factor (`interval × multiplier` each call) |

If you pass invalid values, safe defaults are applied automatically:
- `minInterval ≤ 0` → 1 second
- `maxInterval < minInterval` → 60 × minInterval
- `multiplier ≤ 1.0` → 2.0

### Configuration Strategies

| Scenario | Min | Max | Multiplier |
|----------|-----|-----|------------|
| **High Priority** (urgent retries) | 100ms | 10s | 2.0 |
| **Balanced** (most common) | 5s | 10m | 2.0 |
| **Low Priority** (periodic checks) | 1m | 1h | 1.5 |

## Entry Methods

During reconciliation, call `store.For(obj)` to get the `Entry` for the object being reconciled, then use one of these methods to produce the `(ctrl.Result, error)` return value:

| Method | Resource state | Behavior |
|--------|---------------|----------|
| `ReturnError(err)` | Error occurred | Resets backoff to `minInterval`; returns the error to let controller-runtime handle requeue |
| `IsProgressing()` | Actively changing | Resets backoff to `minInterval`; requeues after that interval |
| `IsStable()` | Desired state reached | Requeues after the current interval; increases the interval for the next call (exponential backoff) |
| `StopRequeue()` | No further polling needed | Removes the entry from the store; returns an empty result with no requeue |

### Backoff Sequence Example

With `minInterval=5s`, `maxInterval=10m`, `multiplier=2.0`, consecutive `IsStable()` calls produce:

```
5s → 10s → 20s → 40s → 1m20s → 2m40s → 5m20s → 10m → 10m → ...
```

Calling `IsProgressing()` or `ReturnError()` at any point resets the sequence back to `5s`.

## Usage in a Reconciler

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    obj := &myv1.MyResource{}
    if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    entry := r.requeueStore.For(obj)

    if err := r.doWork(ctx, obj); err != nil {
        return entry.ReturnError(err)
    }

    switch obj.Status.Phase {
    case "Provisioning":
        return entry.IsProgressing() // Check again soon
    case "Ready":
        return entry.IsStable()      // Periodic health checks
    case "Complete":
        return entry.StopRequeue()   // All done
    default:
        return entry.IsStable()
    }
}
```

## Context Helpers

The package provides helpers for passing an `Entry` through the context, useful when the requeue decision is made deep in the call stack:

- `NewContext(ctx, entry)` returns a new context carrying the given `*Entry`
- `FromContext(ctx)` retrieves the `*Entry` from the context (returns `nil` if absent)

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    obj := &myv1.MyResource{}
    if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    entry := r.requeueStore.For(obj)
    ctx = smartrequeue.NewContext(ctx, entry)

    if err := r.doWork(ctx, obj); err != nil {
        return entry.ReturnError(err)
    }
    return entry.IsStable()
}

// Deep in the call stack, retrieve the entry from context:
func (r *MyReconciler) nestedHelper(ctx context.Context) {
    if entry := smartrequeue.FromContext(ctx); entry != nil {
        // Use entry to make requeue decisions
    }
}
```

## Status Updater Integration

SmartRequeue integrates with the [status updater](./status.md) via `WithSmartRequeue`. Set `rr.SmartRequeue` in your `ReconcileResult` to one of:

| Constant | Calls |
|----------|-------|
| `SR_BACKOFF` | `IsStable()` |
| `SR_RESET` | `IsProgressing()` |
| `SR_NO_REQUEUE` | `StopRequeue()` |

If a reconciliation error is present, the status updater automatically calls `ReturnError` regardless of the `SmartRequeue` field.

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    rr := r.reconcile(ctx, req)

    return ctrlutils.NewStatusUpdaterBuilder[*myv1.MyResource]().
        WithPhaseUpdateFunc(r.computePhase).
        WithConditionUpdater(true).
        WithSmartRequeue(r.requeueStore).
        Build().
        UpdateStatus(ctx, r.client, rr)
}

func (r *MyReconciler) reconcile(ctx context.Context, req reconcile.Request) ctrlutils.ReconcileResult[*myv1.MyResource] {
    obj := &myv1.MyResource{}
    if err := r.client.Get(ctx, req.NamespacedName, obj); err != nil {
        return ctrlutils.ReconcileResult[*myv1.MyResource]{ReconcileError: client.IgnoreNotFound(err)}
    }

    rr := ctrlutils.ReconcileResult[*myv1.MyResource]{Object: obj}

    if err := r.syncExternalResource(ctx, obj); err != nil {
        rr.ReconcileError = err
        return rr
    }

    if !obj.Status.IsReady {
        rr.SmartRequeue = ctrlutils.SR_RESET   // Still provisioning
    } else {
        rr.SmartRequeue = ctrlutils.SR_BACKOFF  // Periodic checks
    }
    return rr
}
```

### SmartRequeue Conditionals

Pass `SmartRequeueConditional` callbacks to `WithSmartRequeue` for decisions that depend on the final status (after conditions have been updated). Conditionals run **after** `UpdateStatus()` completes and can override the `SmartRequeue` field:

```go
return ctrlutils.NewStatusUpdaterBuilder[*myv1.MyResource]().
    WithSmartRequeue(r.requeueStore, func(rr ctrlutils.ReconcileResult[*myv1.MyResource]) ctrlutils.SmartRequeueAction {
        if rr.ReconcileError != nil {
            return ctrlutils.SR_BACKOFF
        }
        for _, cond := range rr.Object.Status.Conditions {
            if cond.Type == "Ready" && cond.Status == "False" {
                return ctrlutils.SR_RESET // Not ready; check again soon
            }
        }
        return ctrlutils.SR_BACKOFF
    }).
    Build().
    UpdateStatus(ctx, r.client, rr)
```

### Explicit RequeueAfter Override

If `ReconcileResult.Result.RequeueAfter` is set, it takes precedence when it's **earlier** than the SmartRequeue-computed interval.

## Common Patterns

### Provision-Then-Monitor
The most common pattern for resources with external dependencies (cloud resources, clusters, etc.):

```go
switch obj.Status.Phase {
case "Provisioning":
    return entry.IsProgressing()  // Check again soon
case "Ready":
    return entry.IsStable()       // Periodic health checks
case "Error":
    return entry.IsProgressing()  // Retry after error cleared
}
```

### Event-Driven with Fallback

Combine watches with SmartRequeue as a safety net for missed events:

```go
if reconcileNeeded(obj) {
    if err := r.sync(ctx, obj); err != nil {
        return entry.ReturnError(err)
    }
    return entry.IsProgressing()  // Immediate follow-up
}
// Fallback periodic check even without events
return entry.IsStable()
```

## Per-Object State Tracking

SmartRequeue tracks state per-object using `(Kind, Name, Namespace)` as the key. Each object gets its own requeue schedule and objects never interfere with each other.

## Thread Safety

The store is thread-safe and uses `sync.RWMutex` internally. No additional synchronization is needed.
