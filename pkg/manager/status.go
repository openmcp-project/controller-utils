package manager

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetCondition synthesizes a Ready condition from the manager status.
// observedGeneration must be set to the current metadata.generation of the
// object being reconciled so consumers can determine whether the condition is stale.
func (s *ManagedResourceStatus) GetCondition(observedGeneration int64) metav1.Condition {
	condStatus := metav1.ConditionFalse
	if s.Phase == StatusPhaseReady {
		condStatus = metav1.ConditionTrue
	}
	reason := s.Phase
	if reason == "" {
		reason = "Unknown"
	}

	return metav1.Condition{
		Type:               "Ready",
		Status:             condStatus,
		Reason:             reason,
		Message:            s.Message,
		ObservedGeneration: observedGeneration,
	}
}
