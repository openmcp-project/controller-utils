package flux_test

import (
	"context"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"github.com/openmcp-project/controller-utils/pkg/manager"
	"github.com/openmcp-project/controller-utils/pkg/manager/flux"
	"github.com/openmcp-project/opencontrolplane-runtime/pkg/serviceprovider/clusteraccess"
	ocprtv1alpha1 "github.com/openmcp-project/opencontrolplane-runtime/testdata/api/v1alpha1"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
)

const (
	testNamespace       = "test"
	testChartURL        = "oci://registry.example.com/charts/my-app"
	testChartPullSecret = "chart-pull-secret"
	testKubeconfigKey   = "kubeconfig-key"
)

// createFakeCluster sets up a ManagedCluster backed by a fake client.
// It mirrors the CreateFakeCluster helper in pkg/manager/utils_test.go but is
// inlined here because that helper is only available inside package manager.
func createFakeCluster(t *testing.T, id string) manager.ManagedCluster {
	t.Helper()
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionv1.AddToScheme(scheme)
	_ = ocprtv1alpha1.AddToScheme(scheme)
	_ = clustersv1alpha1.AddToScheme(scheme)
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	c := clusters.NewTestClusterFromClient(id, fakeClient)
	return manager.NewManagedCluster(c, &rest.Config{}, testNamespace, manager.PlatformCluster)
}

func TestManageFluxResources(t *testing.T) {
	tests := []struct {
		name   string
		params flux.ManageFluxResourcesParams
	}{
		{
			name: "OCIRepository and HelmRelease are created with the correct values",
			params: flux.ManageFluxResourcesParams{
				Cluster:      createFakeCluster(t, "platform"),
				MCPNamespace: "my-service",
				Interval:     time.Hour,
				ClusterContext: clusteraccess.ClusterContext{
					MCPAccessSecretKey: client.ObjectKey{
						Namespace: testNamespace,
						Name:      testKubeconfigKey,
					},
				},
				RequestedVersion: flux.NewFluxResourceVersion(
					"v1.0.0",
					"1.0.0",
					testChartURL,
					testChartPullSecret,
					&apiextensionv1.JSON{Raw: []byte(`{"foo":"bar"}`)},
				),
				OCIRepositoryName: "test-repo",
				HelmReleaseName:   "test-release",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flux.ManageFluxResources(tt.params)

			// apply via a manager
			mgr := manager.NewManager("serviceprovider-test")
			mgr.AddCluster(tt.params.Cluster)
			_, _, err := mgr.Apply(context.TODO())
			require.NoError(t, err)

			// assert OCIRepository
			ociRepo := &sourcev1.OCIRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-repo",
					Namespace: testNamespace,
				},
			}
			require.NoError(t, tt.params.Cluster.GetClient().Get(context.TODO(), client.ObjectKeyFromObject(ociRepo), ociRepo))
			assert.Equal(t, tt.params.RequestedVersion.GetChartURL(), ociRepo.Spec.URL)
			assert.Equal(t, tt.params.RequestedVersion.GetChartPullSecret(), ociRepo.Spec.SecretRef.Name)
			assert.Equal(t, tt.params.RequestedVersion.GetChartVersion(), ociRepo.Spec.Reference.Tag)
			assert.Equal(t, tt.params.Interval, ociRepo.Spec.Interval.Duration)

			// assert HelmRelease
			helmRelease := &helmv2.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-release",
					Namespace: testNamespace,
				},
			}
			require.NoError(t, tt.params.Cluster.GetClient().Get(context.TODO(), client.ObjectKeyFromObject(helmRelease), helmRelease))
			assert.Equal(t, tt.params.RequestedVersion.GetHelmValues(), helmRelease.Spec.Values)
			assert.Equal(t, tt.params.Interval, helmRelease.Spec.Interval.Duration)
			assert.Equal(t, tt.params.MCPNamespace, helmRelease.Spec.StorageNamespace)
			assert.Equal(t, tt.params.MCPNamespace, helmRelease.Spec.TargetNamespace)
			assert.Equal(t, tt.params.ClusterContext.MCPAccessSecretKey.Name, helmRelease.Spec.KubeConfig.SecretRef.Name)
			assert.Equal(t, "kubeconfig", helmRelease.Spec.KubeConfig.SecretRef.Key)
		})
	}
}
