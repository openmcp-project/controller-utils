// Package flux provides generic helpers for deploying Helm charts via Flux
// (OCIRepository + HelmRelease). Service providers implement FluxResourceVersion
// with their own CRD-embedded concrete type; the framework never owns a public
// concrete struct for version data.
package flux

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

	"github.com/openmcp-project/controller-utils/pkg/manager"
	"github.com/openmcp-project/opencontrolplane-runtime/pkg/serviceprovider/clusteraccess"
)

// FluxResourceVersion defines the version information needed to deploy a
// Flux-managed Helm chart. Service providers implement this interface with
// their own CRD-embedded concrete type; the framework never owns a public
// concrete struct for version data.
type FluxResourceVersion interface {
	GetVersion() string
	GetChartVersion() string
	GetChartURL() string
	// GetChartPullSecret returns the name of the secret (already placed in the
	// target namespace) used to pull the Helm chart OCI artifact.
	GetChartPullSecret() string
	GetHelmValues() *apiextensionsv1.JSON
}

// NewFluxResourceVersion creates a FluxResourceVersion backed by a private
// implementation. Useful for tests and for service providers that do not
// define their own CRD-backed version type.
func NewFluxResourceVersion(
	version, chartVersion, chartURL, chartPullSecret string,
	helmValues *apiextensionsv1.JSON,
) FluxResourceVersion {
	return &fluxResourceVersion{
		version:         version,
		chartVersion:    chartVersion,
		chartURL:        chartURL,
		chartPullSecret: chartPullSecret,
		helmValues:      helmValues,
	}
}

// fluxResourceVersion is the package-private concrete implementation of
// FluxResourceVersion.
type fluxResourceVersion struct {
	version         string
	chartVersion    string
	chartURL        string
	chartPullSecret string
	helmValues      *apiextensionsv1.JSON
}

func (r *fluxResourceVersion) GetVersion() string                   { return r.version }
func (r *fluxResourceVersion) GetChartVersion() string              { return r.chartVersion }
func (r *fluxResourceVersion) GetChartURL() string                  { return r.chartURL }
func (r *fluxResourceVersion) GetChartPullSecret() string           { return r.chartPullSecret }
func (r *fluxResourceVersion) GetHelmValues() *apiextensionsv1.JSON { return r.helmValues }

// ManageFluxResourcesParams groups all parameters needed to register the
// Flux resources (OCIRepository + HelmRelease) on a cluster.
type ManageFluxResourcesParams struct {
	// Cluster defines where the resources will be created.
	Cluster manager.ManagedCluster
	// MCPNamespace is the target namespace for the Helm release.
	MCPNamespace string
	// Interval defines the OCIRepository and HelmRelease reconcile intervals.
	Interval time.Duration
	// ClusterContext of the current reconciliation context.
	ClusterContext clusteraccess.ClusterContext
	// RequestedVersion carries version and chart metadata. The chart pull
	// secret name is sourced from RequestedVersion.GetChartPullSecret().
	RequestedVersion FluxResourceVersion
	// OCIRepositoryName is the name of the OCIRepository resource.
	OCIRepositoryName string
	// HelmReleaseName is the name of the HelmRelease resource.
	HelmReleaseName string
}

// ManageFluxResources registers an OCIRepository and a HelmRelease as managed
// objects on the cluster defined in p.Cluster. The objects are not reconciled
// until the caller invokes Manager.Apply.
func ManageFluxResources(p ManageFluxResourcesParams) error {
	if err := validateFluxResourcesParams(p); err != nil {
		return err
	}
	ociRepo := manageOCIRepo(p)
	p.Cluster.AddObject(ociRepo)

	helmRelease := manageHelmRelease(p, []manager.ManagedObject{ociRepo})
	p.Cluster.AddObject(helmRelease)
	return nil
}

func validateFluxResourcesParams(p ManageFluxResourcesParams) error {
	if p.Cluster == nil {
		return fmt.Errorf("ManageFluxResourcesParams.Cluster must not be nil")
	}
	if p.RequestedVersion == nil {
		return fmt.Errorf("ManageFluxResourcesParams.RequestedVersion must not be nil")
	}
	if p.RequestedVersion.GetChartURL() == "" {
		return fmt.Errorf("ManageFluxResourcesParams.RequestedVersion.GetChartURL() must not be empty")
	}
	if p.ClusterContext.MCPAccessSecretKey.Name == "" {
		return fmt.Errorf("ClusterContext.MCPAccessSecretKey.Name must not be empty")
	}
	if p.Interval <= 0 {
		return fmt.Errorf("Interval must be greater than zero, got %v", p.Interval)
	}
	if p.OCIRepositoryName == "" {
		return fmt.Errorf("ManageFluxResourcesParams.OCIRepositoryName must not be empty")
	}
	if p.HelmReleaseName == "" {
		return fmt.Errorf("ManageFluxResourcesParams.HelmReleaseName must not be empty")
	}
	if p.MCPNamespace == "" {
		return fmt.Errorf("ManageFluxResourcesParams.MCPNamespace must not be empty")
	}
	return nil
}

func manageOCIRepo(p ManageFluxResourcesParams) manager.ManagedObject {
	return manager.NewManagedObject(&sourcev1.OCIRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.OCIRepositoryName,
			Namespace: p.Cluster.GetDefaultNamespace(),
		},
	}, manager.ManagedObjectContext{
		ReconcileFunc: func(_ context.Context, o client.Object) error {
			ociRepo, ok := o.(*sourcev1.OCIRepository)
			if !ok {
				return fmt.Errorf("expected *sourcev1.OCIRepository, got %T", o)
			}
			ociRepo.Spec = sourcev1.OCIRepositorySpec{
				Interval: metav1.Duration{Duration: p.Interval},
				URL:      p.RequestedVersion.GetChartURL(),
				Reference: &sourcev1.OCIRepositoryRef{
					Tag: p.RequestedVersion.GetChartVersion(),
				},
				// required to always select the correct OCI layer
				// this mitigates non-deterministic layer ordering across chart versions
				// https://fluxcd.io/flux/components/source/ocirepositories/#layer-selector
				LayerSelector: &sourcev1.OCILayerSelector{
					MediaType: "application/vnd.cncf.helm.chart.content.v1.tar+gzip",
					Operation: "extract",
				},
			}
			if secret := p.RequestedVersion.GetChartPullSecret(); secret != "" {
				ociRepo.Spec.SecretRef = &meta.LocalObjectReference{
					Name: secret,
				}
			}
			return nil
		},
		DeletionPolicy: manager.Delete,
		StatusFunc:     FluxStatus,
	})
}

func manageHelmRelease(p ManageFluxResourcesParams, dependencies []manager.ManagedObject) manager.ManagedObject {
	return manager.NewManagedObject(&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.HelmReleaseName,
			Namespace: p.Cluster.GetDefaultNamespace(),
		},
	}, manager.ManagedObjectContext{
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
				Upgrade: &helmv2.Upgrade{
					Remediation: &helmv2.UpgradeRemediation{
						Retries: 3,
					},
				},
				DriftDetection: &helmv2.DriftDetection{
					Mode: helmv2.DriftDetectionEnabled,
				},
				Values:           p.RequestedVersion.GetHelmValues(),
				TargetNamespace:  p.MCPNamespace,
				StorageNamespace: p.MCPNamespace,
			}
			return nil
		},
		DependsOn:      dependencies,
		DeletionPolicy: manager.Delete,
		StatusFunc:     FluxStatus,
	})
}

// FluxStatus indicates whether the given object is in phase terminating, pending or ready.
func FluxStatus(o client.Object) manager.ManagedResourceStatus {
	fluxObject, ok := o.(conditions.Getter)
	if !ok {
		return manager.ManagedResourceStatus{
			Phase:   manager.StatusPhaseUnknown,
			Message: fmt.Sprintf("object %T does not implement conditions.Getter", o),
		}
	}
	if !o.GetDeletionTimestamp().IsZero() {
		return manager.ManagedResourceStatus{
			Phase:   manager.StatusPhaseTerminating,
			Message: "Resource is terminating.",
		}
	}
	if conditions.IsTrue(fluxObject, meta.ReadyCondition) {
		return manager.ManagedResourceStatus{
			Phase:   manager.StatusPhaseReady,
			Message: "Resource is ready",
		}
	}
	msg := "Resource is not ready"
	if cond := conditions.Get(fluxObject, meta.ReadyCondition); cond != nil && cond.Message != "" {
		msg = cond.Message
	}
	return manager.ManagedResourceStatus{Phase: manager.StatusPhaseProgressing, Message: msg}
}
