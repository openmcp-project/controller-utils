package manager

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// StatusPhaseReady indicates that the resource is ready. All conditions are met and are in status "True".
	StatusPhaseReady = "Ready"
	// StatusPhaseProgressing indicates that the resource is not ready and being created or updated.
	StatusPhaseProgressing = "Progressing"
	// StatusPhaseTerminating indicates that the resource is not ready and in deletion.
	StatusPhaseTerminating = "Terminating"
	// StatusUnknown indicates that no StatusFunc is defined
	StatusPhaseUnkown = "Unkown"
)

// Status defines the status attributes of a ManagedObject.
type Status struct {
	Phase    string
	Message  string
}

// SimpleStatus indicates whether the given object is in phase terminating, pending or ready.
func SimpleStatus(o client.Object) Status {
	if !o.GetDeletionTimestamp().IsZero() {
		return Status{
			Phase:    StatusPhaseTerminating,
			Message:  "Resource is terminating.",
		}
	}
	if o.GetUID() == "" {
		return Status{
			Phase:    StatusPhaseProgressing,
			Message:  "Resource has not been created yet.",
		}
	}
	return Status{
		Phase:    StatusPhaseReady,
		Message:  "Resource exists.",
	}
}