package manager

import "sigs.k8s.io/controller-runtime/pkg/client"

const (
	// LabelManagedBy defines the managed-by label that is added to every managed object.
	LabelManagedBy = "app.kubernetes.io/managed-by"
)

// SetManagedBy sets the managed-by label of the given client.Object.
func SetManagedBy(o client.Object, managedBy string) {
	labels := o.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[LabelManagedBy] = managedBy
	o.SetLabels(labels)
}

// ManagedBy returns a list option filtering objects by managed-by label value.
func ManagedBy(managedBy string) client.ListOption {
	return client.MatchingLabels{
		LabelManagedBy: managedBy,
	}
}
