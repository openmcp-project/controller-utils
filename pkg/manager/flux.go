package manager

import (
	"context"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/opencontrolplane-runtime/pkg/serviceprovider/clusteraccess"
)

// ManageFluxResourcesParams groups all parameters to create the required Flux resources.
type ManageFluxResourcesParams struct {
	// Cluster defines where the resources will be created.
	Cluster ManagedCluster
	// MCPNamespace is the target namespace for the Helm release.
	MCPNamespace string
	// ChartPullSecretName is the name of the image-pull secret placed in the cluster namespace.
	ChartPullSecretName string
	// Interval defines the OCIRepository and HelmRelease reconcile intervals.
	Interval time.Duration
	// ClusterContext of the current reconciliation context
	ClusterContext clusteraccess.ClusterContext
	// RequestedVersion is the version of External Secrets Operator that a user requested through the onboarding API
	RequestedVersion RequestedVersion
	// OCIRepositoryName is the name of the OCIRepository resource.
	OCIRepositoryName string
	// HelmReleaseName is the name of the HelmRelease resource.
	HelmReleaseName string
}

type RequestedVersion struct {
	Version string
	ChartVersion string
	ChartURL string
	ChartPullSecret string
	HelmValues *apiextensionsv1.JSON
}

// ManageFluxResources registers an OCIRepository and a HelmRelease as managed objects on the
// cluster defined in p.Cluster. The objects are not reconciled until the caller invokes
// Manager.Apply.
func ManageFluxResources(p ManageFluxResourcesParams) {
	ociDependants := []ManagedObject{}
	ociRepo := manageOCIRepo(p, ociDependants)
	p.Cluster.AddObject(ociRepo)

	helmDependants := []ManagedObject{ociRepo}
	helmRelease := manageHelmRelease(p, helmDependants)
	p.Cluster.AddObject(helmRelease)
}

func manageOCIRepo(p ManageFluxResourcesParams, dependants []ManagedObject) ManagedObject {
	ociRepo := NewManagedObject(&sourcev1.OCIRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.OCIRepositoryName,
			Namespace: p.Cluster.GetDefaultNamespace(),
		},
	}, ManagedObjectContext{
		ReconcileFunc: func(_ context.Context, o client.Object) error {
			ociRepo, ok := o.(*sourcev1.OCIRepository)
			if !ok {
				return fmt.Errorf("expected *sourcev1.OCIRepository, got %T", o)
			}
			if p.RequestedVersion.ChartURL == "" {
				// this should never happen as long as defaulting works properly
				return fmt.Errorf("missing ChartURL definition for version %s", p.RequestedVersion.Version)
			}
			ociRepo.Spec = sourcev1.OCIRepositorySpec{
				Interval: metav1.Duration{Duration: p.Interval},
				URL:      p.RequestedVersion.ChartURL,
				Reference: &sourcev1.OCIRepositoryRef{
					Tag: p.RequestedVersion.ChartVersion,
				},
				// required to always select the correct OCI layer
				// this mitigates non-deterministic layer ordering across chart versions
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
		DependsOn:      dependants,
		DeletionPolicy: Delete,
		StatusFunc:     FluxStatus,
	})
	return ociRepo
}

func manageHelmRelease(p ManageFluxResourcesParams, dependants []ManagedObject) ManagedObject {
	helmRelease := NewManagedObject(&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.HelmReleaseName,
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
					Name:      p.OCIRepositoryName,
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
		DependsOn:      dependants,
		DeletionPolicy: Delete,
		StatusFunc:     FluxStatus,
	})
	return helmRelease
}

// FluxStatus indicates whether the given object is in phase terminating, pending or ready.
func FluxStatus(o client.Object) Status {
	fluxObject := o.(conditions.Getter)
	if !o.GetDeletionTimestamp().IsZero() {
		return Status{
			Phase:    StatusPhaseTerminating,
			Message:  "Resource is terminating.",
		}
	}
	if conditions.IsTrue(fluxObject, meta.ReadyCondition) {
		return Status{
			Phase:    StatusPhaseReady,
			Message:  "Resource is ready",
		}
	}
	return Status{
		Phase:    StatusPhaseProgressing,
		Message:  "Resource is not ready",
	}
}
