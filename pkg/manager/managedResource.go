package manager

// ManagedResource is the read-only view of a reconciled object returned by Apply and Delete.
// It is a plain value struct: Apply/Delete return it by value, so callers receive a copy and
// cannot affect the manager's internal state. It stays decoupled from any CRD schema; consumers
// project it into their own CRD-embedded type via ProjectResources and ResourceStatusWriter.
type ManagedResource struct {
	APIGroup  string
	Kind      string
	Name      string
	Namespace string
	Location  ClusterType
	Status    ManagedResourceStatus
}

// ResourceRef identifies a managed resource. It is passed to a consumer's
// ResourceStatusWriter so the consumer can decide how to store the fields
// (e.g. whether to represent an empty Namespace or APIGroup as nil).
type ResourceRef struct {
	APIGroup  string
	Kind      string
	Name      string
	Namespace string
	Location  string
}

// ResourceStatusWriter is implemented by a consumer's CRD-embedded resource type.
// ProjectResources drives it to populate a consumer-owned status entry from a
// framework ManagedResource, keeping pkg/manager decoupled from any CRD schema.
type ResourceStatusWriter interface {
	// SetReference stores the identity of the managed resource.
	SetReference(ResourceRef)
	// SetPhase stores the lifecycle phase and human-readable message.
	SetPhase(phase, message string)
}

// ProjectResources converts framework results into a slice of a consumer-provided
// type W. The framework owns the loop and allocation; the consumer owns only the
// field mapping via the ResourceStatusWriter setters. newWriter constructs a fresh,
// writable W for each result.
//
// Example:
//
//	obj.Status.Resources = manager.ProjectResources(
//		resources,
//		func() *apiv1alpha1.ManagedResource { return &apiv1alpha1.ManagedResource{} },
//	)
func ProjectResources[W ResourceStatusWriter](results []ManagedResource, newWriter func() W) []W {
	out := make([]W, len(results))
	for i, r := range results {
		w := newWriter()
		w.SetReference(ResourceRef{
			APIGroup:  r.APIGroup,
			Kind:      r.Kind,
			Name:      r.Name,
			Namespace: r.Namespace,
			Location:  string(r.Location),
		})
		w.SetPhase(r.Status.Phase, r.Status.Message)
		out[i] = w
	}
	return out
}
