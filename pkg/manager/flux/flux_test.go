package flux_test

import (
	"context"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/runtime/conditions"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
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

// validParams returns a fully populated ManageFluxResourcesParams. Individual
// test cases mutate one field to exercise a specific validation branch.
func validParams(t *testing.T) flux.ManageFluxResourcesParams {
	t.Helper()
	return flux.ManageFluxResourcesParams{
		Cluster:      createFakeCluster(t, "platform"),
		MCPNamespace: "my-ns",
		Interval:     time.Hour,
		ClusterContext: clusteraccess.ClusterContext{
			MCPAccessSecretKey: client.ObjectKey{
				Namespace: testNamespace,
				Name:      testKubeconfigKey,
			},
		},
		RequestedVersion:  flux.NewFluxResourceVersion("v1", "1.0", testChartURL, "", nil),
		OCIRepositoryName: "my-repo",
		HelmReleaseName:   "my-release",
	}
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
			require.NoError(t, flux.ManageFluxResources(tt.params))

			// apply via a manager
			mgr := manager.NewManager("serviceprovider-test")
			mgr.AddCluster(tt.params.Cluster)
			_, done, err := mgr.Apply(context.TODO())
			require.NoError(t, err)
			// The fake client never sets a Flux ReadyCondition, so FluxStatus always returns
			// Progressing. done must be false until Flux itself reconciles the resources.
			assert.False(t, done, "expected done=false: Flux resources are not yet Ready (no ReadyCondition set by fake client)")

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

			// LayerSelector
			require.NotNil(t, ociRepo.Spec.LayerSelector)
			assert.Equal(t, "application/vnd.cncf.helm.chart.content.v1.tar+gzip", ociRepo.Spec.LayerSelector.MediaType)
			assert.Equal(t, "extract", ociRepo.Spec.LayerSelector.Operation)

			// assert HelmRelease
			helmRelease := &helmv2.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-release",
					Namespace: testNamespace,
				},
			}
			require.NoError(t, tt.params.Cluster.GetClient().Get(context.TODO(), client.ObjectKeyFromObject(helmRelease), helmRelease))
			// ChartRef
			assert.Equal(t, "OCIRepository", helmRelease.Spec.ChartRef.Kind)
			assert.Equal(t, tt.params.OCIRepositoryName, helmRelease.Spec.ChartRef.Name)
			assert.Equal(t, tt.params.Cluster.GetDefaultNamespace(), helmRelease.Spec.ChartRef.Namespace)

			// Install
			assert.Equal(t, 3, helmRelease.Spec.Install.Remediation.Retries)
			assert.Equal(t, true, helmRelease.Spec.Install.CreateNamespace)

			// Update
			assert.Equal(t, 3, helmRelease.Spec.Upgrade.Remediation.Retries)

			// DriftDetection
			assert.Equal(t, helmv2.DriftDetectionEnabled, helmRelease.Spec.DriftDetection.Mode)

			assert.Equal(t, tt.params.RequestedVersion.GetHelmValues(), helmRelease.Spec.Values)
			assert.Equal(t, tt.params.Interval, helmRelease.Spec.Interval.Duration)
			assert.Equal(t, tt.params.MCPNamespace, helmRelease.Spec.StorageNamespace)
			assert.Equal(t, tt.params.MCPNamespace, helmRelease.Spec.TargetNamespace)
			assert.Equal(t, tt.params.ClusterContext.MCPAccessSecretKey.Name, helmRelease.Spec.KubeConfig.SecretRef.Name)
			assert.Equal(t, "kubeconfig", helmRelease.Spec.KubeConfig.SecretRef.Key)
		})
	}
}

// TestManageFluxResources_Validation exercises every guard in validateFluxResourcesParams.
// Each case starts from a fully valid params struct and breaks exactly one field.
func TestManageFluxResources_Validation(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(p *flux.ManageFluxResourcesParams)
		wantErrFrag string
	}{
		{
			name:        "nil Cluster",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.Cluster = nil },
			wantErrFrag: "Cluster must not be nil",
		},
		{
			name:        "nil RequestedVersion",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.RequestedVersion = nil },
			wantErrFrag: "RequestedVersion must not be nil",
		},
		{
			name: "empty ChartURL",
			mutate: func(p *flux.ManageFluxResourcesParams) {
				p.RequestedVersion = flux.NewFluxResourceVersion("v1", "1.0", "", "", nil)
			},
			wantErrFrag: "GetChartURL() must not be empty",
		},
		{
			name: "empty MCPAccessSecretKey.Name",
			mutate: func(p *flux.ManageFluxResourcesParams) {
				p.ClusterContext.MCPAccessSecretKey.Name = ""
			},
			wantErrFrag: "MCPAccessSecretKey.Name must not be empty",
		},
		{
			name:        "zero Interval",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.Interval = 0 },
			wantErrFrag: "Interval must be greater than zero",
		},
		{
			name:        "negative Interval",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.Interval = -time.Second },
			wantErrFrag: "Interval must be greater than zero",
		},
		{
			name:        "empty OCIRepositoryName",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.OCIRepositoryName = "" },
			wantErrFrag: "OCIRepositoryName must not be empty",
		},
		{
			name:        "empty HelmReleaseName",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.HelmReleaseName = "" },
			wantErrFrag: "HelmReleaseName must not be empty",
		},
		{
			name:        "empty MCPNamespace",
			mutate:      func(p *flux.ManageFluxResourcesParams) { p.MCPNamespace = "" },
			wantErrFrag: "MCPNamespace must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validParams(t)
			tt.mutate(&p)
			err := flux.ManageFluxResources(p)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrFrag)
		})
	}
}

// TestFluxStatus exercises all four branches of FluxStatus:
// Unknown (non-Flux object), Terminating, Ready, and Progressing (with and without a message).
func TestFluxStatus(t *testing.T) {
	tests := []struct {
		name        string
		object      client.Object
		wantPhase   string
		wantMessage string
	}{
		{
			name:      "Unknown — object does not implement conditions.Getter",
			object:    &corev1.Secret{},
			wantPhase: manager.StatusPhaseUnknown,
		},
		{
			name: "Terminating — DeletionTimestamp is set",
			object: func() client.Object {
				now := metav1.Now()
				o := &sourcev1.OCIRepository{}
				o.DeletionTimestamp = &now
				return o
			}(),
			wantPhase:   manager.StatusPhaseTerminating,
			wantMessage: "Resource is terminating.",
		},
		{
			name: "Ready — ReadyCondition is True",
			object: func() client.Object {
				o := &sourcev1.OCIRepository{}
				conditions.MarkTrue(o, "Ready", "Reconciled", "chart applied successfully")
				return o
			}(),
			wantPhase: manager.StatusPhaseReady,
		},
		{
			name: "Progressing — ReadyCondition is False with a message",
			object: func() client.Object {
				o := &sourcev1.OCIRepository{}
				conditions.MarkFalse(o, "Ready", "ChartNotFound", "chart not found in registry")
				return o
			}(),
			wantPhase:   manager.StatusPhaseProgressing,
			wantMessage: "chart not found in registry",
		},
		{
			name:        "Progressing — no conditions set",
			object:      &sourcev1.OCIRepository{},
			wantPhase:   manager.StatusPhaseProgressing,
			wantMessage: "Resource is not ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := flux.FluxStatus(tt.object)
			assert.Equal(t, tt.wantPhase, status.Phase)
			if tt.wantMessage != "" {
				assert.Equal(t, tt.wantMessage, status.Message)
			}
		})
	}
}
