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

// Status defines the status attributes of a ManagedObject.
type Status struct {
	Phase   string
	Message string
}

// SimpleStatus indicates whether the given object is in phase terminating, progressing, or ready.
func SimpleStatus(o client.Object) Status {
	if !o.GetDeletionTimestamp().IsZero() {
		return Status{
			Phase:   StatusPhaseTerminating,
			Message: "Resource is terminating.",
		}
	}
	if o.GetUID() == "" {
		return Status{
			Phase:   StatusPhaseProgressing,
			Message: "Resource has not been created yet.",
		}
	}
	return Status{
		Phase:   StatusPhaseReady,
		Message: "Resource exists.",
	}
}
