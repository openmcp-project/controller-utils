package manager

import (
	"context"

	commonapi "github.com/openmcp-project/openmcp-operator/api/common"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeletionPolicy distinguishes between normal deletion and orphaning an object.
type DeletionPolicy string
type InstancePhase string

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
type StatusFunc func(o client.Object, resourceLocation ClusterType) Status

// SimpleStatus indicates whether the given object is in phase terminating, pending or ready.
func SimpleStatus(o client.Object, resourceLocation ClusterType) Status {
	if !o.GetDeletionTimestamp().IsZero() {
		return Status{
			Phase:    commonapi.StatusPhaseTerminating,
			Message:  "Resource is terminating.",
			Location: resourceLocation,
		}
	}
	if o.GetUID() == "" {
		return Status{
			Phase:    commonapi.StatusPhaseProgressing,
			Message:  "Resource has not been created yet.",
			Location: resourceLocation,
		}
	}
	return Status{
		Phase:    commonapi.StatusPhaseReady,
		Message:  "Resource exists.",
		Location: resourceLocation,
	}
}

// Status defines the status attributes of a ManagedObject.
type Status struct {
	Phase    InstancePhase
	Message  string
	Location ClusterType
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
		managedBy:      moc.ManagedBy,
	}
}

// ManagedObjectContext holds the data to manage a client.Object.
type ManagedObjectContext struct {
	ReconcileFunc  ReconcileFunc
	DependsOn      []ManagedObject
	DeletionPolicy DeletionPolicy
	StatusFunc     StatusFunc
	ManagedBy       string
}

// ManagedObject represents an object managed by a Mana^ger.
type ManagedObject interface {
	GetObject() client.Object
	Reconcile(ctx context.Context) error
	GetDependencies() []ManagedObject
	GetDeletionPolicy() DeletionPolicy
	GetStatus(resourceLocation ClusterType) Status
	Label() string
}

var _ ManagedObject = &managedObject{}

type managedObject struct {
	object         client.Object
	reconcileFunc  ReconcileFunc
	statusFunc     StatusFunc
	dependencies   []ManagedObject
	deletionPolicy DeletionPolicy
	managedBy      string
}

// GetStatus implements ManagedObject.
func (m *managedObject) GetStatus(resourceLocation ClusterType) Status {
	if m.statusFunc != nil {
		return m.statusFunc(m.object, resourceLocation)
	}
	return Status{
		Phase:    "Unknown",
		Message:  "No status function defined.",
		Location: resourceLocation,
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

// Label implements ManagedObject.
func (m *managedObject) Label() string {
	return m.managedBy
}
