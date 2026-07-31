package externalsecrets

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/service-provider-external-secrets/api/v1alpha1"
)

// DeletionPolicy distinguishes between normal deletion and orphaning an object.
type DeletionPolicy string

const (
	// Orphan indicates that an object will be orphaned when deletion is requested
	Orphan DeletionPolicy = "orphan"
	// Delete indicates that an object will be deleted when deletion is requested
	Delete DeletionPolicy = "delete"
)

// ReconcileFunc reconciles the given client.Object.
type ReconcileFunc func(ctx context.Context, o client.Object) error

// NoOp does not do anything with the provided object and returns nil.
func NoOp(context.Context, client.Object) error {
	return nil
}

// StatusFunc provides Status information for the given client.Object.
type StatusFunc func(o client.Object, rl v1alpha1.ResourceLocation) Status

// SimpleStatus indicates whether the given object is in phase terminating, pending or ready.
func SimpleStatus(o client.Object, rl v1alpha1.ResourceLocation) Status {
	if !o.GetDeletionTimestamp().IsZero() {
		return Status{
			Phase:    v1alpha1.Terminating,
			Message:  "Resource is terminating.",
			Location: rl,
		}
	}
	if o.GetUID() == "" {
		return Status{
			Phase:    v1alpha1.Pending,
			Message:  "Resource has not been created yet.",
			Location: rl,
		}
	}
	return Status{
		Phase:    v1alpha1.Ready,
		Message:  "Resource exists.",
		Location: rl,
	}
}

// Status defines the status attributes of a ManagedObject.
type Status struct {
	Phase    v1alpha1.InstancePhase
	Message  string
	Location v1alpha1.ResourceLocation
}

// NewManagedObject creates a new ManagedObject instances to manage the given client.Object.
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
	GetStatus(v1alpha1.ResourceLocation) Status
}

var _ ManagedObject = &managedObject{}

type managedObject struct {
	object         client.Object
	reconcileFunc  ReconcileFunc
	statusFunc     StatusFunc
	dependencies   []ManagedObject
	deletionPolicy DeletionPolicy
}

// GetStatus implements ManagedObject.
func (m *managedObject) GetStatus(rl v1alpha1.ResourceLocation) Status {
	if m.statusFunc != nil {
		return m.statusFunc(m.object, rl)
	}
	return Status{
		Phase:    v1alpha1.Unknown,
		Message:  "No status function defined.",
		Location: rl,
	}
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
