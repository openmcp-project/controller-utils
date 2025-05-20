package clusteraccess

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FailIfNotManaged takes an object and a list of expected labels.
// It returns an ResourceNotManagedError, if any of the expected labels is missing on the object or has a different value.
// If the object is nil or the expected labels are empty, it returns nil.
func FailIfNotManaged(obj client.Object, expectedLabels ...Label) error {
	if obj == nil || len(expectedLabels) == 0 {
		return nil
	}
	actualLabels := obj.GetLabels()
	if len(actualLabels) == 0 {
		return NewResourceNotManagedError(obj, expectedLabels...)
	}
	for _, expected := range expectedLabels {
		if actual, ok := actualLabels[expected.Key]; !ok || actual != expected.Value {
			return NewResourceNotManagedError(obj, expectedLabels...)
		}
	}
	return nil
}

// NewResourceNotManagedError creates a new ResourceNotManagedError.
func NewResourceNotManagedError(obj client.Object, expectedLabels ...Label) *ResourceNotManagedError {
	SortLabels(expectedLabels)
	return &ResourceNotManagedError{
		Obj:            obj,
		ExpectedLabels: expectedLabels,
	}
}

type ResourceNotManagedError struct {
	Obj            client.Object
	ExpectedLabels []Label
}

var _ error = &ResourceNotManagedError{}

func (e *ResourceNotManagedError) Error() string {
	kind := e.Obj.GetObjectKind().GroupVersionKind().Kind
	if kind == "" {
		kind = "resource"
	}
	nsMod := ""
	if e.Obj.GetNamespace() != "" {
		nsMod = e.Obj.GetNamespace() + "/"
	}
	actualLabels := make([]Label, 0, len(e.Obj.GetLabels()))
	for k, v := range e.Obj.GetLabels() {
		actualLabels = append(actualLabels, Label{Key: k, Value: v})
	}
	// Sort the labels for consistent error messages
	SortLabels(actualLabels)
	return fmt.Sprintf("%s '%s%s' exists but does not contain the expected management labels %v, its actual labels are %v", kind, nsMod, e.Obj.GetName(), e.ExpectedLabels, actualLabels)
}

// IsResourceNotManagedError returns true if the error is non-nil and of type *ResourceNotManagedError.
func IsResourceNotManagedError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*ResourceNotManagedError); ok {
		return true
	}
	return false
}
