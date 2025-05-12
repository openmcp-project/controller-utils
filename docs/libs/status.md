# Kubernetes Resource Status Updating

## Conditions

The `pkg/conditions` package helps with managing condition lists which can often be found in the status of k8s resources.

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

## Status Updater

The status updater is based on the idea that many of our resources use a status similar to this:
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


### How to use the status updater

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

#### Some more details

- The status updater uses reflection to modifiy the status' fields. This requires it to know the field names (the ones in go, not the ones in the YAML representation). By default, it expects them to be `Status` for the status itself and `Phase`, `ObservedGeneration`, `LastReconcileTime`, `Reason`, `Message`, and `Conditions` for the respective fields within the status.
	- To use a different field name, overwrite it by using either `WithFieldOverride` or `WithFieldOverrides`.
	- If any of the fields is not contained top-level in the status but within a nested struct, the names of these fields must be prefixed with the names of the corresponding structs, separated by a `.`. The `WithNestedStruct` method can be used to set such a prefix quickly for one or more fields.
	- To disable the update of a specific field altogether, set its name to the empty string. This can be done via the aforementioned `WithFieldOverride`/`WithFieldOverrides` methods, or simpler via `WithoutFields`.
		- Doing this for the status field itself turns the status update into a no-op.
	- The package contains constants with the field keys that are required by most of these methods. `STATUS_FIELD` refers to the `Status` field itself, the other field keys are prefixed with `STATUS_FIELD_`.
		- The `AllStatusFields()` function returns a list containing all status field keys, _except the one for the status field itself_, for convenience.
- The `WithCustomUpdateFunc` method can be used to inject a function that performs custom logic on the resource's status. Note that while the function gets the complete object as an argument, only changes to its status will be updated by the status updater.

### The ReconcileResult

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
