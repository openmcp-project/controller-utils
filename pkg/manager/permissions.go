package externalsecrets

import (
	"context"
	"fmt"

	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"

	externalsecretsoperatorsv1alpha1 "github.com/openmcp-project/service-provider-external-secrets/api/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ESONamespace carries the resolved ESO installation namespace between ResolveEsoNamespace and TokenAccessGenerator.
type ESONamespace string

const (
	verbCreate = "create"
	verbDelete = "delete"
	verbGet    = "get"
	verbList   = "list"
	verbPatch  = "patch"
	verbUpdate = "update"
	verbWatch  = "watch"
)

// ResolveEsoNamespace is an AdditionalDataResolver that returns the ESO installation namespace.
func ResolveEsoNamespace(_ context.Context, obj *externalsecretsoperatorsv1alpha1.ExternalSecretsOperator, providerConfig *externalsecretsoperatorsv1alpha1.ProviderConfig) (any, error) {
	if obj == nil {
		return nil, fmt.Errorf("obj must not be nil")
	}
	if providerConfig == nil {
		return nil, fmt.Errorf("providerConfig must not be nil")
	}
	for _, version := range providerConfig.Spec.Versions {
		if version.Version != obj.Spec.Version {
			continue
		}
		helmValues, err := ExtractHelmValues(version.HelmValues)
		if err != nil {
			return nil, fmt.Errorf("failed to extract helm values: %w", err)
		}
		namespace := DefaultNamespace
		if helmValues.NamespaceOverride != "" {
			namespace = helmValues.NamespaceOverride
		}
		return ESONamespace(namespace), nil
	}
	return nil, fmt.Errorf("version %s not found in ProviderConfig", obj.Spec.Version)
}

// TokenAccesGenerator returns a TokenConfig with the RBAC permissions required to install ESO.
func TokenAccesGenerator(_ reconcile.Request, additionalData ...any) (*clustersv1alpha1.TokenConfig, error) {
	namespace := DefaultNamespace
	for _, data := range additionalData {
		ns, ok := data.(ESONamespace)
		if ok && ns != "" {
			namespace = string(ns)
			break
		}
	}
	return &clustersv1alpha1.TokenConfig{
		Permissions: []clustersv1alpha1.PermissionsRequest{
			{
				Namespace: namespace,
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{""},
						Resources: []string{"serviceaccounts", "services"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"apps"},
						Resources: []string{"deployments", "replicasets"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles", "rolebindings"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"policy"},
						Resources: []string{"poddisruptionbudgets"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"cert-manager.io"},
						Resources: []string{"certificates"},
						Verbs:     []string{rbacv1.VerbAll},
					},
				},
			},
			{
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{"apiextensions.k8s.io"},
						Resources: []string{"customresourcedefinitions"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"clusterroles", "clusterrolebindings"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"admissionregistration.k8s.io"},
						Resources: []string{"validatingwebhookconfigurations"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"endpoints"},
						Verbs:     []string{verbGet, verbList, verbWatch},
					},
					{
						APIGroups: []string{"discovery.k8s.io"},
						Resources: []string{"endpointslices"},
						Verbs:     []string{verbGet, verbList, verbWatch},
					},
					{
						APIGroups: []string{"coordination.k8s.io"},
						Resources: []string{"leases"},
						Verbs:     []string{verbGet, verbCreate, verbPatch, verbUpdate},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
						Verbs:     []string{verbGet, verbList, verbWatch, verbCreate, verbPatch, verbUpdate},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"events"},
						Verbs:     []string{verbCreate, verbPatch},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"secrets"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"serviceaccounts"},
						Verbs:     []string{verbGet, verbList, verbWatch},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"serviceaccounts/token"},
						Verbs:     []string{verbCreate},
					},
					{
						APIGroups: []string{"authentication.k8s.io"},
						Resources: []string{"tokenreviews"},
						Verbs:     []string{verbCreate},
					},
					{
						APIGroups: []string{"authorization.k8s.io"},
						Resources: []string{"subjectaccessreviews"},
						Verbs:     []string{verbCreate},
					},
					{
						APIGroups: []string{"monitoring.coreos.com"},
						Resources: []string{"servicemonitors"},
						Verbs:     []string{rbacv1.VerbAll},
					},
					{
						APIGroups: []string{"external-secrets.io"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
					{
						APIGroups: []string{"generators.external-secrets.io"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				},
			},
		},
	}, nil
}
