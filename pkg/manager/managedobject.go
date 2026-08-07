package manager

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeletionPolicy distinguishes between normal deletion and orphaning an object.
type DeletionPolicy string

const (
	// Orphan indicates that an object will be orphaned when deletion is requested.
	Orphan DeletionPolicy = "orphan"
	// Delete indicates that an object will be deleted when deletion is requested.
	Delete DeletionPolicy = "delete"
)

// ReconcileFunc reconciles the given client.Object.
type ReconcileFunc func(ctx context.Context, o client.Object) error

// StatusFunc provides ManagedResourceStatus information for the given client.Object.
type StatusFunc func(client.Object) ManagedResourceStatus

// NoOp does not do anything with the provided object and returns nil.
func NoOp(context.Context, client.Object) error {
	return nil
}

// NewManagedObject creates a new ManagedObject instance to manage the given client.Object.
func NewManagedObject(o client.Object, moc ManagedObjectContext) ManagedObject {
	if moc.DeletionPolicy == "" {
		moc.DeletionPolicy = Delete
	}

	return &managedObject{
		object:         o,
		reconcileFunc:  moc.ReconcileFunc,
		dependencies:   moc.DependsOn,
		deletionPolicy: moc.DeletionPolicy,
		statusFunc:     moc.StatusFunc,
	}
}

// ManagedObjectContext holds the data to manage a client.Object.
type ManagedObjectContext struct {
	ReconcileFunc  ReconcileFunc
	DependsOn      []ManagedObject
	DeletionPolicy DeletionPolicy
	StatusFunc     StatusFunc
}

// ManagedObject represents an object managed by a Manager.
type ManagedObject interface {
	GetObject() client.Object
	Reconcile(ctx context.Context) error
	GetDependencies() []ManagedObject
	GetDeletionPolicy() DeletionPolicy
	GetStatus() ManagedResourceStatus
}

var _ ManagedObject = &managedObject{}

type managedObject struct {
	object         client.Object
	reconcileFunc  ReconcileFunc
	dependencies   []ManagedObject
	deletionPolicy DeletionPolicy
	statusFunc     StatusFunc
}

// GetDeletionPolicy implements ManagedObject.
func (m *managedObject) GetDeletionPolicy() DeletionPolicy {
	return m.deletionPolicy
}

// GetDependencies implements ManagedObject.
func (m *managedObject) GetDependencies() []ManagedObject {
	return m.dependencies
}

// Reconcile implements ManagedObject.
func (m *managedObject) Reconcile(ctx context.Context) error {
	if m.reconcileFunc != nil {
		return m.reconcileFunc(ctx, m.object)
	}
	return nil
}

// GetObject implements ManagedObject.
func (m *managedObject) GetObject() client.Object {
	return m.object
}

// GetStatus implements ManagedObject.
func (m *managedObject) GetStatus() ManagedResourceStatus {
	if m.statusFunc != nil {
		return m.statusFunc(m.object)
	}
	return ManagedResourceStatus{
		Phase:   StatusPhaseUnknown,
		Message: "No status function defined.",
	}
}
