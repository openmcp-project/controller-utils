package manager

import (
	corev1 "k8s.io/api/core/v1"
)

type InstancePhase string

// Constants representing the phases of an instance lifecycle.
const (
	Pending     InstancePhase = "Pending"
	Progressing InstancePhase = "Progressing"
	Ready       InstancePhase = "Ready"
	Failed      InstancePhase = "Failed"
	Terminating InstancePhase = "Terminating"
	Unknown     InstancePhase = "Unknown"
)

// ManagedResource defines a kubernetes object with its lifecycle phase
type ManagedResource struct {
	corev1.TypedObjectReference `json:",inline"`

	// +required
	// Phase InstancePhase `json:"phase"`
	// +optional
	// Message string `json:"message,omitempty"`

	// +required
	Status Status `json:"status,omitempty"`
	
	// +optional
	Location ClusterType `json:"location,omitempty"`
}
