package manager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// OperationResultDeletionFailed indicates failed to be deleted
	OperationResultDeletionFailed controllerutil.OperationResult = "deletionFailed"
	// OperationResultDeletionRequested indicates that an object has been marked for deletion
	OperationResultDeletionRequested controllerutil.OperationResult = "deletionRequested"
	// OperationResultDeleted indicates that an object has been deleted
	OperationResultDeleted controllerutil.OperationResult = "deleted"
	// OperationResultOrphaned indicates that an object has been orphaned
	OperationResultOrphaned controllerutil.OperationResult = OperationResultDeleted
)

type dependents map[ManagedObject][]dependency

// Manager manages the objects of an arbitrary number of clusters
type Manager interface {
	AddCluster(mc ManagedCluster)
	AddCleaner(oc OrphanCleaner)
	Apply(context.Context) (_ []ManagedResource, cleanup error)
	Delete(context.Context) (_ []ManagedResource, cleanup error)
}

// OrphanCleaner removes any previously managed objects that are no longer part of the desired state.
type OrphanCleaner interface {
	// []Result contains cleanup errors that can be mapped to a managed object.
	// error represents cleanup errors that cannot be mapped to a managed object.
	Cleanup(ctx context.Context) ([]Result, error)
}

// NewManager creates a new Manager instance.
func NewManager(serviceProvider string) Manager {
	return &managerImpl{
		serviceProvider: serviceProvider,
		clusters:       []ManagedCluster{},
		cleaners:       []OrphanCleaner{},
	}
}

// managerImpl manages clusters and invokes reconciliation of ManagedObjects.
type managerImpl struct {
	serviceProvider string
	clusters []ManagedCluster
	cleaners []OrphanCleaner
}

// AddCluster adds a cluster to a Manager.
func (m *managerImpl) AddCluster(mc ManagedCluster) {
	m.clusters = append(m.clusters, mc)
}

// Apply reconciles all managed objects in registration order.
// DependsOn is NOT used for apply ordering; it only affects deletion sequencing.
// Callers are responsible for registering objects in dependency order.
func (m *managerImpl) Apply(ctx context.Context) ([]ManagedResource, error) {
	results, err := m.reconcileObjects(ctx, false)
	if err != nil {
		return []ManagedResource{}, err
	}
	managedResources, resultContainsError := resultsToResources(ctx, results)
	if resultContainsError {
		return managedResources, err
	}
	if allResourcesReady(results) {
		return managedResources, nil
	}
	return managedResources, nil
}

// Delete invokes deletion of all ManagedObjects.
func (m *managerImpl) Delete(ctx context.Context) ([]ManagedResource, error) {
	results, err := m.reconcileObjects(ctx, true)
	if err != nil {
		return []ManagedResource{}, err
	}
	managedResources, resultContainsError := resultsToResources(ctx, results)
	if resultContainsError {
		return managedResources, err
	}
	if allDeleted(results) {
		return managedResources, nil
	}
	return managedResources, nil
}

// AddCleaner adds a cleaner to a Manager.
func (m *managerImpl) AddCleaner(cleaner OrphanCleaner) {
	m.cleaners = append(m.cleaners, cleaner)
}

func (m *managerImpl) reconcileObjects(ctx context.Context, isDeletion bool) ([]Result, error) {
	dependents := m.getDependents()

	// Apply or delete objects from each cluster.
	results := []Result{}
	for _, mc := range m.clusters {
		for _, mo := range mc.GetObjects() {
			result := m.reconcileObject(ctx, mc, mo, dependents, isDeletion)
			results = append(results, result)
		}
	}

	// remove any redundant resources like secret copies that are no longer part of the desired state.
	for _, c := range m.cleaners {
		result, err := c.Cleanup(ctx)
		if err != nil {
			return results, err
		}
		results = slices.Concat(results, result)
	}

	return results, nil
}

func (m *managerImpl) reconcileObject(ctx context.Context, mc ManagedCluster, mo ManagedObject, dependents dependents, isDeletion bool) Result {
	client := mc.GetClient()
	obj := mo.GetObject()

	if isDeletion {
		if err := m.checkForDependents(ctx, dependents[mo]); err != nil {
			return Result{
				Object:          mo,
				Cluster:         mc,
				OperationResult: controllerutil.OperationResultNone,
				Error:           err,
			}
		}

		if mo.GetDeletionPolicy() == Orphan {
			return Result{
				Object:          mo,
				Cluster:         mc,
				OperationResult: OperationResultOrphaned,
				Error:           nil,
			}
		}

		err := client.Delete(ctx, obj)
		if apierrors.IsNotFound(err) {
			return Result{
				Object:          mo,
				Cluster:         mc,
				OperationResult: OperationResultDeleted,
				Error:           nil,
			}
		}
		return Result{
			Object:          mo,
			Cluster:         mc,
			OperationResult: OperationResultDeletionRequested,
			Error:           err,
		}
	}

	opResult, err := controllerutil.CreateOrUpdate(ctx, client, obj, func() error {
		SetManagedBy(obj , m.serviceProvider)
		return mo.Reconcile(ctx)
	})
	return Result{
		Object:          mo,
		Cluster:         mc,
		OperationResult: opResult,
		Error:           err,
	}
}

func (m *managerImpl) checkForDependents(ctx context.Context, deps []dependency) error {
	errs := []error{}
	for _, dep := range deps {
		obj := dep.Object.GetObject()
		err := dep.Cluster.GetClient().Get(ctx, client.ObjectKeyFromObject(obj), obj)
		if apierrors.IsNotFound(err) {
			// "Not found" is the success case: The object which depends on us does not exist anymore.
			continue
		}
		if err != nil {
			// Some unexpected error occurred.
			errs = append(errs, err)
			continue
		}
		// No error occurred, the GET request has been successful.
		// The object still exists and depends on us.
		errs = append(errs, fmt.Errorf("dependent object still exists: %s", ObjectID(obj)))
	}
	return errors.Join(errs...)
}

func (m *managerImpl) getDependents() dependents {
	deps := dependents{}
	for _, mc := range m.clusters {
		for _, mo := range mc.GetObjects() {
			for _, dep := range mo.GetDependencies() {
				if deps[dep] == nil {
					deps[dep] = []dependency{}
				}
				deps[dep] = append(deps[dep], dependency{
					Object:  mo,
					Cluster: mc,
				})
			}
		}
	}
	return deps
}

// Result summarizes a reconciliation result.
type Result struct {
	Object          ManagedObject
	Cluster         ManagedCluster
	OperationResult controllerutil.OperationResult
	Error           error
}

type dependency struct {
	Object  ManagedObject
	Cluster ManagedCluster
}

// AllDeleted returns true if every item's operation result is OperationResultDeleted.
func allDeleted(results []Result) bool {
	for _, r := range results {
		if r.OperationResult != OperationResultDeleted {
			return false
		}
	}
	return true
}

func allResourcesReady(results []Result) bool {
	for _, r := range results {
		if r.OperationResult != StatusPhaseReady {
			return false
		}
	}
	return true
}

func resultsToResources(ctx context.Context, results []Result) ([]ManagedResource, bool) {
	l := log.FromContext(ctx)
	containsError := false
	resources := make([]ManagedResource, 0, len(results))
	for _, res := range results {
		obj := res.Object.GetObject()
		resources = append(resources, ManagedResource{
			TypedObjectReference: corev1.TypedObjectReference{
				Kind:      reflect.TypeOf(obj).Elem().Name(),
				Name:      obj.GetName(),
				Namespace: nilIfEmptyString(obj.GetNamespace()),
			},
			Status: res.Object.GetStatus(),
			Location: res.Object.GetLocation(),
		})
		if res.Error != nil {
			containsError = true
			l.Error(res.Error, "objectID", ObjectID(obj))
		}
	}
	return resources, containsError
}

func nilIfEmptyString(str string) *string {
	if str == "" {
		return nil
	}
	return ptr.To(str)
}
