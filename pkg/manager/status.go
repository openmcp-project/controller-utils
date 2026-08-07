package manager

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// StatusPhaseReady indicates that the resource is ready.
	StatusPhaseReady = "Ready"
	// StatusPhaseProgressing indicates that the resource is not ready and being created or updated.
	StatusPhaseProgressing = "Progressing"
	// StatusPhaseTerminating indicates that the resource is not ready and in deletion.
	StatusPhaseTerminating = "Terminating"
	// StatusPhaseUnknown indicates that no StatusFunc is defined.
	StatusPhaseUnknown = "Unknown"
)

// ManagedResourceStatus defines the status attributes of a ManagedObject.
type ManagedResourceStatus struct {
	Phase   string
	Message string
}

// SimpleStatus indicates whether the given object is in phase terminating, progressing, or ready.
func SimpleStatus(o client.Object) ManagedResourceStatus {
	if !o.GetDeletionTimestamp().IsZero() {
		return ManagedResourceStatus{
			Phase:   StatusPhaseTerminating,
			Message: "Resource is terminating.",
		}
	}
	if o.GetUID() == "" {
		return ManagedResourceStatus{
			Phase:   StatusPhaseProgressing,
			Message: "Resource has not been created yet.",
		}
	}
	return ManagedResourceStatus{
		Phase:   StatusPhaseReady,
		Message: "Resource exists.",
	}
}
