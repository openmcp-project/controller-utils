package manager

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"github.com/openmcp-project/opencontrolplane-runtime/testdata/api/v1alpha1"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// CreateFakeCluster sets up a cluster with a fake client
func CreateFakeCluster(t *testing.T, id string, clusterObjects ...client.Object) *clusters.Cluster {
	t.Helper()
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextv1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = clustersv1alpha1.AddToScheme(scheme)
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)

	// init cluster with objects
	fakeClient := fake.NewClientBuilder().WithObjects(clusterObjects...).WithScheme(scheme).Build()
	return clusters.NewTestClusterFromClient(id, fakeClient)
}

// ExecApply sets up a manager for the provided clusters and invokes reconciliation of all managed objects.
// wantDone asserts the expected value of the done return value from Apply.
func ExecApply(t *testing.T, clusters []ManagedCluster, expectedManagedObjects int, wantDone bool) []ManagedResource {
	t.Helper()
	mgr := NewManager("serviceprovider-test")
	for _, cluster := range clusters {
		mgr.AddCluster(cluster)
	}
	managedResources, done, err := mgr.Apply(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, wantDone, done, "unexpected done value from Apply")
	return assertResult(t, managedResources, expectedManagedObjects)
}

func assertResult(t *testing.T, managedResources []ManagedResource, expectedManagedObjects int) []ManagedResource {
	t.Helper()
	assert.Len(t, managedResources, expectedManagedObjects, "expected %d managed object(s), got %d managed object(s)")
	return managedResources
}

func TestApply_Done(t *testing.T) {
	const ns = "default"

	t.Run("done=false when StatusFunc returns Progressing", func(t *testing.T) {
		cluster := CreateFakeCluster(t, "platform")
		mc := NewManagedCluster(cluster, &rest.Config{}, ns, PlatformCluster)
		obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: ns}}
		mo := NewManagedObject(obj, ManagedObjectContext{
			ReconcileFunc: func(_ context.Context, _ client.Object) error { return nil },
			StatusFunc: func(_ client.Object) ManagedResourceStatus {
				return ManagedResourceStatus{Phase: StatusPhaseProgressing, Message: "always progressing"}
			},
		})
		mc.AddObject(mo)
		mgr := NewManager("test-sp")
		mgr.AddCluster(mc)

		_, done, err := mgr.Apply(context.TODO())
		require.NoError(t, err)
		assert.False(t, done, "expected done=false when resource reports Progressing")
	})

	t.Run("done=true when all resources report Ready via SimpleStatus", func(t *testing.T) {
		// Pre-populate the fake store with the object including a UID.
		// CreateOrUpdate will call Get (succeeds, UID propagated to obj), then Update.
		// SimpleStatus checks GetUID() != "" → returns Ready → done=true.
		preExisting := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: ns, UID: "test-uid-1234"},
		}
		cluster := CreateFakeCluster(t, "platform", preExisting)
		mc := NewManagedCluster(cluster, &rest.Config{}, ns, PlatformCluster)
		obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: ns}}
		mo := NewManagedObject(obj, ManagedObjectContext{
			ReconcileFunc: func(_ context.Context, _ client.Object) error { return nil },
			StatusFunc:    SimpleStatus,
		})
		mc.AddObject(mo)
		mgr := NewManager("test-sp")
		mgr.AddCluster(mc)

		_, done, err := mgr.Apply(context.TODO())
		require.NoError(t, err)
		assert.True(t, done, "expected done=true when all resources report Ready")
	})

	t.Run("done=false when resource is Terminating via SimpleStatus", func(t *testing.T) {
		// DeletionTimestamp must be on the server-side object (preExisting), not on the
		// local obj. controllerutil.CreateOrUpdate calls cl.Get which overwrites obj with
		// the server state, so any DeletionTimestamp set only on obj is lost before
		// SimpleStatus is called. A finalizer is required to keep the object in the store
		// while it has a DeletionTimestamp.
		now := metav1.Now()
		preExisting := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-secret",
				Namespace:         ns,
				UID:               "test-uid-1234",
				DeletionTimestamp: &now,
				Finalizers:        []string{"test-finalizer"},
			},
		}
		cluster := CreateFakeCluster(t, "platform", preExisting)
		mc := NewManagedCluster(cluster, &rest.Config{}, ns, PlatformCluster)
		obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: ns}}
		mo := NewManagedObject(obj, ManagedObjectContext{
			ReconcileFunc: func(_ context.Context, _ client.Object) error { return nil },
			StatusFunc:    SimpleStatus,
		})
		mc.AddObject(mo)
		mgr := NewManager("test-sp")
		mgr.AddCluster(mc)

		resources, done, err := mgr.Apply(context.TODO())
		require.NoError(t, err)
		require.Len(t, resources, 1)
		assert.False(t, done, "expected done=false when resource is Terminating")
		assert.Equal(t, StatusPhaseTerminating, resources[0].Status.Phase)
	})
}

func TestDelete_Done(t *testing.T) {
	const ns = "default"
	const secretName = "managed-secret"

	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns},
	}
	cluster := CreateFakeCluster(t, "platform", existingSecret)
	mc := NewManagedCluster(cluster, &rest.Config{}, ns, PlatformCluster)

	obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns}}
	mo := NewManagedObject(obj, ManagedObjectContext{
		ReconcileFunc:  func(_ context.Context, _ client.Object) error { return nil },
		StatusFunc:     SimpleStatus,
		DeletionPolicy: Delete,
	})
	mc.AddObject(mo)
	mgr := NewManager("test-sp")
	mgr.AddCluster(mc)

	// First Delete: fake client removes the object (returns nil, not NotFound).
	// The manager receives DeletionRequested → done=false.
	_, done, err := mgr.Delete(context.TODO())
	require.NoError(t, err)
	assert.False(t, done, "expected done=false: deletion was requested but object not yet confirmed gone")

	// Second Delete: object is already gone, fake client returns NotFound.
	// The manager receives Deleted → done=true.
	_, done, err = mgr.Delete(context.TODO())
	require.NoError(t, err)
	assert.True(t, done, "expected done=true: object confirmed deleted (NotFound)")
}

func TestDelete_OrphanPolicy(t *testing.T) {
	const ns = "default"
	const secretName = "orphaned-secret"

	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns},
	}
	cluster := CreateFakeCluster(t, "platform", existingSecret)
	mc := NewManagedCluster(cluster, &rest.Config{}, ns, PlatformCluster)

	obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns}}
	mo := NewManagedObject(obj, ManagedObjectContext{
		ReconcileFunc:  func(_ context.Context, _ client.Object) error { return nil },
		StatusFunc:     SimpleStatus,
		DeletionPolicy: Orphan,
	})
	mc.AddObject(mo)
	mgr := NewManager("test-sp")
	mgr.AddCluster(mc)

	stored := &corev1.Secret{}
	require.NoError(t, cluster.Client().Get(
		context.TODO(),
		client.ObjectKey{Name: secretName, Namespace: ns},
		stored,
	), "Secret was created")

	// Orphaned objects are considered done immediately — cl.Delete is never called.
	_, done, err := mgr.Delete(context.TODO())
	require.NoError(t, err)
	assert.True(t, done, "expected done=true: orphaned object counts as deleted")

	// The object must still be in the store — it was left alive intentionally.
	require.NoError(t, cluster.Client().Get(
		context.TODO(),
		client.ObjectKey{Name: secretName, Namespace: ns},
		stored,
	), "orphaned object should still exist in the store")
}

func TestDelete_DependsOn(t *testing.T) {
	const ns = "default"
	// secretA is a dependency; secretB depends on A.
	// A cannot be deleted while B still exists.
	const nameA, nameB = "secret-a", "secret-b"

	cluster := CreateFakeCluster(t, "platform",
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: nameA, Namespace: ns}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: nameB, Namespace: ns}},
	)
	mc := NewManagedCluster(cluster, &rest.Config{}, ns, PlatformCluster)

	objA := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: nameA, Namespace: ns}}
	moA := NewManagedObject(objA, ManagedObjectContext{
		ReconcileFunc: NoOp, StatusFunc: SimpleStatus, DeletionPolicy: Delete,
	})

	objB := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: nameB, Namespace: ns}}
	moB := NewManagedObject(objB, ManagedObjectContext{
		ReconcileFunc:  NoOp,
		StatusFunc:     SimpleStatus,
		DeletionPolicy: Delete,
		// B depends on A: the framework blocks A's deletion until B is gone.
		DependsOn: []ManagedObject{moA},
	})

	mc.AddObject(moA)
	mc.AddObject(moB)
	mgr := NewManager("test-sp")
	mgr.AddCluster(mc)

	// Iteration 1: A is blocked (B still exists); B deletion is requested.
	_, done, err := mgr.Delete(context.TODO())
	assert.False(t, done, "iteration 1: done should be false — A is blocked")
	require.Error(t, err, "iteration 1: expected error — A cannot be deleted while B exists")
	assert.ErrorIs(t, err, ErrManagedResourcesFailed)

	// Iteration 2: B is confirmed gone (fake client removes immediately on Delete);
	// A deletion is now requested. done is still false — A is DeletionRequested, not Deleted.
	_, done, err = mgr.Delete(context.TODO())
	require.NoError(t, err)
	assert.False(t, done, "iteration 2: done should be false — A deletion pending")

	// Iteration 3: both A and B return NotFound → done=true.
	_, done, err = mgr.Delete(context.TODO())
	require.NoError(t, err)
	assert.True(t, done, "iteration 3: expected done=true — all objects confirmed deleted")
}
