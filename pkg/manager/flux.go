package externalsecrets

import (
	"context"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluxcd/pkg/runtime/conditions"

	"github.com/openmcp-project/opencontrolplane-runtime/pkg/serviceprovider/clusteraccess"

	apiv1alpha1 "github.com/openmcp-project/service-provider-external-secrets/api/v1alpha1"
)

const (
	// DefaultNamespace is the default namespace where External Secrets Operator components are deployed on the ManagedControlPlane
	DefaultNamespace = "external-secrets"
	// OCIRepositoryName is the name of the External Secrets Operator OCIRepository resource
	OCIRepositoryName = "external-secrets"
	// HelmReleaseName is the name of the External Secrets Operator HelmRelease resource
	HelmReleaseName = "external-secrets"
)

// ManageFluxResourcesParams groups all parameters to create the required manage flux resources
type ManageFluxResourcesParams struct {
	// Cluster defines where the resources will be created
	Cluster ManagedCluster
	// MCPNamespace defines the namespace name that deploy ESO
	MCPNamespace string
	// ChartPullSecretName defines the name of the secret copy that will be placed in the Cluster namespace
	ChartPullSecretName string
	// Obj is the tenant API object that is being reconciled
	Obj *apiv1alpha1.ExternalSecretsOperator
	// Interval defines OCIRepository and HelmRelease reconcile intervals
	Interval time.Duration
	// ClusterContext of the current reconciliation context
	ClusterContext clusteraccess.ClusterContext
	// RequestedVersion is the version of External Secrets Operator that a user requested through the onboarding API
	RequestedVersion apiv1alpha1.ExternalSecretsVersion
}

// ManageFluxResources configures OCIRepo and HelmRelease
func ManageFluxResources(p ManageFluxResourcesParams) {
	ociRepo := NewManagedObject(&sourcev1.OCIRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OCIRepositoryName,
			Namespace: p.Cluster.GetDefaultNamespace(),
		},
	}, ManagedObjectContext{
		ReconcileFunc: func(_ context.Context, o client.Object) error {
			ociRepo, ok := o.(*sourcev1.OCIRepository)
			if !ok {
				return fmt.Errorf("expected *sourcev1.OCIRepository, got %T", o)
			}
			if p.RequestedVersion.ChartURL == nil {
				// this should never happen as long as defaulting works properly
				return fmt.Errorf("missing ChartURL definition for Flux version %s", p.RequestedVersion.Version)
			}
			ociRepo.Spec = sourcev1.OCIRepositorySpec{
				Interval: metav1.Duration{Duration: p.Interval},
				URL:      *p.RequestedVersion.ChartURL,
				Reference: &sourcev1.OCIRepositoryRef{
					Tag: p.RequestedVersion.ChartVersion,
				},
				// required to always select the correct OCI layer
				// this mitigates non-deterministic layer ordering across different eso versions
				// that prevented the OCIRepository from getting ready for some eso versions
				// https://fluxcd.io/flux/components/source/ocirepositories/#layer-selector
				LayerSelector: &sourcev1.OCILayerSelector{
					MediaType: "application/vnd.cncf.helm.chart.content.v1.tar+gzip",
					Operation: "extract",
				},
			}
			if p.ChartPullSecretName != "" {
				ociRepo.Spec.SecretRef = &meta.LocalObjectReference{
					Name: p.ChartPullSecretName,
				}
			}
			return nil
		},
		DependsOn:      []ManagedObject{},
		DeletionPolicy: Delete,
		StatusFunc:     FluxStatus,
	})
	p.Cluster.AddObject(ociRepo)

	helmRelease := NewManagedObject(&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HelmReleaseName,
			Namespace: p.Cluster.GetDefaultNamespace(),
		},
	}, ManagedObjectContext{
		ReconcileFunc: func(_ context.Context, o client.Object) error {
			helmRelease, ok := o.(*helmv2.HelmRelease)
			if !ok {
				return fmt.Errorf("expected *helmv2.HelmRelease, got %T", o)
			}
			helmRelease.Spec = helmv2.HelmReleaseSpec{
				Interval: metav1.Duration{Duration: p.Interval},
				ChartRef: &helmv2.CrossNamespaceSourceReference{
					Kind:      "OCIRepository",
					Name:      OCIRepositoryName,
					Namespace: p.Cluster.GetDefaultNamespace(),
				},
				KubeConfig: &meta.KubeConfigReference{
					SecretRef: &meta.SecretKeyReference{
						Name: p.ClusterContext.MCPAccessSecretKey.Name,
						Key:  "kubeconfig",
					},
				},
				Install: &helmv2.Install{
					Remediation: &helmv2.InstallRemediation{
						Retries: 3,
					},
					CreateNamespace: true,
				},
				DriftDetection: &helmv2.DriftDetection{
					Mode: helmv2.DriftDetectionEnabled,
				},
				Values:           p.RequestedVersion.HelmValues,
				TargetNamespace:  p.MCPNamespace,
				StorageNamespace: p.MCPNamespace,
			}
			return nil
		},
		DependsOn:      []ManagedObject{ociRepo},
		DeletionPolicy: Delete,
		StatusFunc:     FluxStatus,
	})
	p.Cluster.AddObject(helmRelease)
}

// FluxStatus indicates whether the given object is in phase terminating, pending or ready.
func FluxStatus(o client.Object, rl apiv1alpha1.ResourceLocation) Status {
	fluxObject := o.(conditions.Getter)
	if !o.GetDeletionTimestamp().IsZero() {
		return Status{
			Phase:    apiv1alpha1.Terminating,
			Message:  "Resource is terminating.",
			Location: rl,
		}
	}
	if conditions.IsTrue(fluxObject, meta.ReadyCondition) {
		return Status{
			Phase:    apiv1alpha1.Ready,
			Message:  "Resource is ready",
			Location: rl,
		}
	}
	return Status{
		Phase:    apiv1alpha1.Pending,
		Message:  "Resource is not ready",
		Location: rl,
	}
}
