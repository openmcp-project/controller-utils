package externalsecrets

import (
	"context"
	"errors"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/openmcp-project/service-provider-external-secrets/api/v1alpha1"
)

// ErrOrphanCleanup is an user-facing error that indicates orphan cleanup failures
var ErrOrphanCleanup = errors.New("orphan cleanup failed")

var _ OrphanCleaner = &orphanCleaner[*corev1.SecretList]{}

type orphanCleaner[T client.ObjectList] struct {
	cluster     ManagedCluster
	namespace   string
	cleanerType cleanerType[T]
}

type cleanerType[T client.ObjectList] struct {
	ObjectsToKeep []corev1.LocalObjectReference
	EmptyList     func() T
	// PreDeletionSteps is an optional hook that is invoked before OrphanCleaner.Cleanup deletes an object.
	// It returns:
	//   - proceedWithDeletion: true if the object is ready to be deleted.
	//     Returning false prevents deletion for this reconciliation cycle.
	//   - err: an error if the preparation step failed.
	PreDeletionSteps func(context.Context, client.Object) (proceedWithDeletion bool, _ error)
}

// NewOrphanCleaner removes redundant objects in the given target namespace.
func NewOrphanCleaner[T client.ObjectList](cluster ManagedCluster, namespace string, clType cleanerType[T]) OrphanCleaner {
	return &orphanCleaner[T]{
		cluster:     cluster,
		namespace:   namespace,
		cleanerType: clType,
	}
}

func (c *orphanCleaner[T]) items(list T) []client.Object {
	items, _ := meta.ExtractList(list)
	objList := make([]client.Object, 0, len(items))
	for _, item := range items {
		if obj, ok := item.(client.Object); ok {
			objList = append(objList, obj)
		}
	}
	return objList
}

func (c *orphanCleaner[T]) Cleanup(ctx context.Context) ([]Result, error) {
	results := []Result{}
	if c.cleanerType.EmptyList == nil {
		return nil, fmt.Errorf("%w: orphan cleaner is missing empty list definition", ErrOrphanCleanup)
	}
	objList := c.cleanerType.EmptyList()
	cl := c.cluster.GetClient()
	if err := cl.List(ctx, objList,
		client.InNamespace(c.namespace),
		client.MatchingLabels{LabelManagedBy: LabelManagedByValue},
	); err != nil {
		log.FromContext(ctx).Error(err, "failed to list objects for orphan cleanup")
		return nil, ErrOrphanCleanup
	}
	for _, obj := range c.items(objList) {
		if !slices.ContainsFunc(c.cleanerType.ObjectsToKeep, func(ref corev1.LocalObjectReference) bool { return obj.GetName() == ref.Name }) {
			// exec delete preparation steps
			if c.cleanerType.PreDeletionSteps != nil {
				proceedWithDeletion, err := c.cleanerType.PreDeletionSteps(ctx, obj)
				if err != nil {
					results = append(results, c.deletionError(obj, err))
					continue
				}
				if !proceedWithDeletion {
					results = append(results, c.deletionPrepared(obj))
					continue
				}
			}
			// exec delete
			if err := cl.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
				results = append(results, c.deletionError(obj, err))
			}
		}
	}
	return results, nil
}

func (c *orphanCleaner[T]) deletionPrepared(obj client.Object) Result {
	return Result{
		Object: &managedObject{
			object:         obj,
			statusFunc:     cleanupPreparedStatus,
			deletionPolicy: Delete,
		},
		Cluster:         c.cluster,
		OperationResult: OperationResultDeletionRequested,
	}
}

func cleanupPreparedStatus(_ client.Object, rl apiv1alpha1.ResourceLocation) Status {
	return Status{
		Phase:    apiv1alpha1.Terminating,
		Message:  "Orphan cleanup prepared",
		Location: rl,
	}
}

func (c *orphanCleaner[T]) deletionError(obj client.Object, err error) Result {
	return Result{
		Object: &managedObject{
			object:         obj,
			statusFunc:     cleanupErrorStatus,
			deletionPolicy: Delete,
		},
		Cluster:         c.cluster,
		OperationResult: OperationResultDeletionFailed,
		Error:           err,
	}
}

func cleanupErrorStatus(_ client.Object, rl apiv1alpha1.ResourceLocation) Status {
	return Status{
		Phase:    apiv1alpha1.Terminating,
		Message:  "Orphan cleanup failed",
		Location: rl,
	}
}
