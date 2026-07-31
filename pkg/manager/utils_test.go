package externalsecrets

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/openmcp-project/controller-utils/pkg/clusters"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openmcp-project/service-provider-external-secrets/api/v1alpha1"
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

// ExecApply sets up a manager for the provided clusters and invokes reconciliation of all managed objects
func ExecApply(t *testing.T, clusters []ManagedCluster, expectedManagedObjects int, wantErrors []string) []Result {
	t.Helper()
	// invoke apply with manager
	mgr := NewManager()
	for _, cluster := range clusters {
		mgr.AddCluster(cluster)
	}
	results, err := mgr.Apply(context.TODO())
	require.NoError(t, err)
	return assertResult(t, results, expectedManagedObjects, wantErrors)
}

func assertResult(t *testing.T, results []Result, expectedManagedObjects int, wantErrors []string) []Result {
	t.Helper()
	assert.Len(t, results, expectedManagedObjects, "expected %d managed object(s), got %d managed object(s)")
	errcount := 0
	for _, r := range results {
		if r.Error != nil {
			// assert that an error is expected
			assert.Contains(t, wantErrors, r.Object.GetObject().GetName(), "unexpected reconcile error of managed object %s", r.Object.GetObject().GetName())
			errcount++
		}
	}
	// assert that the overall number of errors is expected
	assert.Equal(t, len(wantErrors), errcount, "expected %d reconcile error(s), got %d reconcile error(s)", len(wantErrors), errcount)
	return results
}
