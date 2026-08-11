package manager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// ErrManagedResourcesFailed is returned by Apply/Delete when one or more
// managed resources encountered a reconcile error.
var ErrManagedResourcesFailed = errors.New("one or more managed resources failed")

type dependents map[ManagedObject][]dependency

// Manager manages the objects of an arbitrary number of clusters.
type Manager interface {
	AddCluster(mc ManagedCluster)
	AddCleaner(oc OrphanCleaner)
	// Apply reconciles all managed objects in registration order.
	// DependsOn is NOT used for apply ordering; it only affects deletion sequencing.
	// Callers are responsible for registering objects in dependency order.
	// done=true means every resource reached StatusPhaseReady.
	// err wraps ErrManagedResourcesFailed when any individual reconcile failed.
	Apply(ctx context.Context) (resources []ManagedResource, done bool, err error)
	// Delete deletes all managed objects.
	// done=true means every resource has been deleted or orphaned.
	// err wraps ErrManagedResourcesFailed when any individual deletion failed.
	Delete(ctx context.Context) (resources []ManagedResource, done bool, err error)
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
		clusters:        []ManagedCluster{},
		cleaners:        []OrphanCleaner{},
	}
}

// managerImpl manages clusters and invokes reconciliation of ManagedObjects.
type managerImpl struct {
	serviceProvider string
	clusters        []ManagedCluster
	cleaners        []OrphanCleaner
}

// AddCluster adds a cluster to a Manager.
func (m *managerImpl) AddCluster(mc ManagedCluster) {
	m.clusters = append(m.clusters, mc)
}

// Apply implements Manager.
func (m *managerImpl) Apply(ctx context.Context) ([]ManagedResource, bool, error) {
	results, err := m.reconcileObjects(ctx, false)
	if err != nil {
		return nil, false, err
	}
	resources, errs := resultsToResources(ctx, results)
	if len(errs) > 0 {
		return resources, false, fmt.Errorf("%w: %w", ErrManagedResourcesFailed, errors.Join(errs...))
	}
	return resources, allResourcesReady(results), nil
}

// Delete implements Manager.
func (m *managerImpl) Delete(ctx context.Context) ([]ManagedResource, bool, error) {
	results, err := m.reconcileObjects(ctx, true)
	if err != nil {
		return nil, false, err
	}
	resources, errs := resultsToResources(ctx, results)
	if len(errs) > 0 {
		return resources, false, fmt.Errorf("%w: %w", ErrManagedResourcesFailed, errors.Join(errs...))
	}
	return resources, allDeleted(results), nil
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

	// Remove any redundant resources like secret copies that are no longer part of the desired state.
	for _, c := range m.cleaners {
		result, err := c.Cleanup(ctx)
		if err != nil {
			return results, err
		}
		results = slices.Concat(results, result)
	}

	if len(results) == 0 {
		log.FromContext(ctx).V(1).Info("manager reconciled zero objects; no clusters or managed objects have been registered")
	}

	return results, nil
}

func (m *managerImpl) reconcileObject(ctx context.Context, mc ManagedCluster, mo ManagedObject, dependents dependents, isDeletion bool) Result {
	cl := mc.GetClient()
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
			}
		}

		err := cl.Delete(ctx, obj)
		if apierrors.IsNotFound(err) {
			return Result{
				Object:          mo,
				Cluster:         mc,
				OperationResult: OperationResultDeleted,
			}
		}
		if err != nil {
			return Result{
				Object:          mo,
				Cluster:         mc,
				OperationResult: OperationResultDeletionFailed,
				Error:           err,
			}
		}
		return Result{
			Object:          mo,
			Cluster:         mc,
			OperationResult: OperationResultDeletionRequested,
		}
	}

	opResult, err := controllerutil.CreateOrUpdate(ctx, cl, obj, func() error {
		SetManagedBy(obj, m.serviceProvider)
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
	var errs []error
	for _, dep := range deps {
		obj := dep.Object.GetObject()
		err := dep.Cluster.GetClient().Get(ctx, client.ObjectKeyFromObject(obj), obj)
		if apierrors.IsNotFound(err) {
			// "Not found" is the success case: the dependent no longer exists.
			continue
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// No error: the GET succeeded, meaning the dependent still exists.
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

// allDeleted returns true if every result is OperationResultDeleted or OperationResultOrphaned.
// It uses OperationResult because deletion completion is an immediate API fact,
// not a condition that needs to be polled from the object's status.
func allDeleted(results []Result) bool {
	for _, r := range results {
		if r.OperationResult != OperationResultDeleted && r.OperationResult != OperationResultOrphaned {
			return false
		}
	}
	return true
}

// allResourcesReady returns true if every managed object has reached StatusPhaseReady.
// It uses the object's StatusFunc output rather than OperationResult because convergence
// (e.g. a Flux HelmRelease becoming ready) is asynchronous: OperationResult reflects
// what the API server did ("created"/"updated"), not whether the controller has reconciled.
func allResourcesReady(results []Result) bool {
	for _, r := range results {
		if r.Object.GetStatus().Phase != StatusPhaseReady {
			return false
		}
	}
	return true
}

func resultsToResources(ctx context.Context, results []Result) ([]ManagedResource, []error) {
	l := log.FromContext(ctx)
	var errs []error
	resources := make([]ManagedResource, 0, len(results))
	for _, res := range results {
		obj := res.Object.GetObject()

		apiGroup := ""
		kind := reflect.TypeOf(obj).Elem().Name() // fallback: Go type name == Kind by convention
		if gvk, err := res.Cluster.GetClient().GroupVersionKindFor(obj); err == nil {
			apiGroup = gvk.Group
			kind = gvk.Kind
		} else {
			l.Error(err, "cannot determine GVK for managed object; apiGroup will be empty in status", "objectID", ObjectID(obj))
		}

		resources = append(resources, &managedResourceResult{
			apiGroup:   apiGroup,
			kind:       kind,
			name:       obj.GetName(),
			namespace:  obj.GetNamespace(),
			generation: obj.GetGeneration(),
			status:     res.Object.GetStatus(),
			location:   res.Cluster.GetClusterType(),
		})
		if res.Error != nil {
			l.Error(res.Error, "reconcile error", "objectID", ObjectID(obj))
			errs = append(errs, res.Error)
		}
	}
	return resources, errs
}
