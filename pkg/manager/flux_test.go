package manager

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
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/opencontrolplane-runtime/pkg/serviceprovider/clusteraccess"
)

const (
	testNamespace       = "test"
	testChartURL        = "oci://registry.example.com/charts/my-app"
	testChartPullSecret = "chart-pull-secret"
	testKubeconfigKey   = "kubeconfig-key"
)

func TestManageFluxResources(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		params ManageFluxResourcesParams
	}{
		{
			name: "OCIRepository and HelmRelease are created with the correct values",
			params: ManageFluxResourcesParams{
				Cluster:             NewManagedCluster(CreateFakeCluster(t, "platform"), &rest.Config{}, testNamespace, PlatformCluster),
				MCPNamespace:        "my-service",
				ChartPullSecretName: "secret-copy",
				Interval:            time.Hour,
				ClusterContext: clusteraccess.ClusterContext{
					MCPAccessSecretKey: client.ObjectKey{
						Namespace: testNamespace,
						Name:      testKubeconfigKey,
					},
				},
				RequestedVersion: RequestedVersion{
					Version:      "v1.0.0",
					ChartURL:     testChartURL,
					ChartVersion: "1.0.0",
					HelmValues: &apiextensionv1.JSON{
						Raw: []byte(`{"foo":"bar"}`),
					},
				},
				OCIRepositoryName: "test-repo",
				HelmReleaseName:   "test-release",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ManageFluxResources(tt.params)
			// expect oci repo and helm release without errors
			ExecApply(t, []ManagedCluster{tt.params.Cluster}, 2, []string{})

			// assert oci repo
			ociRepo := &sourcev1.OCIRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-repo",
					Namespace: testNamespace,
				},
			}
			require.NoError(t, tt.params.Cluster.GetClient().Get(context.TODO(), client.ObjectKeyFromObject(ociRepo), ociRepo))
			assert.Equal(t, tt.params.RequestedVersion.ChartURL, ociRepo.Spec.URL)
			assert.Equal(t, tt.params.ChartPullSecretName, ociRepo.Spec.SecretRef.Name)
			assert.Equal(t, tt.params.RequestedVersion.ChartVersion, ociRepo.Spec.Reference.Tag)
			assert.Equal(t, tt.params.Interval, ociRepo.Spec.Interval.Duration)

			// assert helm release
			helmRelease := &helmv2.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-release",
					Namespace: testNamespace,
				},
			}
			require.NoError(t, tt.params.Cluster.GetClient().Get(context.TODO(), client.ObjectKeyFromObject(helmRelease), helmRelease))
			assert.Equal(t, tt.params.RequestedVersion.HelmValues, helmRelease.Spec.Values)
			assert.Equal(t, tt.params.Interval, helmRelease.Spec.Interval.Duration)
			assert.Equal(t, tt.params.MCPNamespace, helmRelease.Spec.StorageNamespace)
			assert.Equal(t, tt.params.MCPNamespace, helmRelease.Spec.TargetNamespace)
			assert.Equal(t, tt.params.ClusterContext.MCPAccessSecretKey.Name, helmRelease.Spec.KubeConfig.SecretRef.Name)
			assert.Equal(t, "kubeconfig", helmRelease.Spec.KubeConfig.SecretRef.Key)
		})
	}
}
