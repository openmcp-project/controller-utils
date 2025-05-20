package clusteraccess

import (
	"context"
	"fmt"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
)

// GetTokenBasedAccess is a convenience function that wraps the flow of ensuring namespace, serviceaccount, (cluster)role(binding), and creating the token.
// It returns a kubeconfig, the token with expiration timestamp, and an error if any of the steps fail.
// The name will be used for all resources except the namespace (serviceaccount, (cluster)role, (cluster)rolebinding), with anything role-related additionally being prefixed with rolePrefix.
// The namespace holds the serviceaccount and, if namespaceScoped is true, the role and rolebinding.
// If namespaceScoped is false, clusterrole and clusterrolebinding are used.
func GetTokenBasedAccess(ctx context.Context, c client.Client, restCfg *rest.Config, name, namespace string, namespaceScoped bool, rolePrefix string, rules []rbacv1.PolicyRule, expectedLabels ...Label) ([]byte, *ServiceAccountToken, error) {
	if namespace == "" {
		return nil, nil, fmt.Errorf("no namespace provided for ServiceAccount")
	}

	_, err := EnsureNamespace(ctx, c, namespace, expectedLabels...)
	if err != nil {
		return nil, nil, err
	}

	sa, err := EnsureServiceAccount(ctx, c, name, namespace, expectedLabels...)
	if err != nil {
		return nil, nil, err
	}

	subjects := []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: name, Namespace: namespace}}
	if namespaceScoped {
		_, _, err = EnsureRoleAndBinding(ctx, c, rolePrefix+name, namespace, subjects, rules, expectedLabels...)
		if err != nil {
			return nil, nil, err
		}
	} else {
		_, _, err = EnsureClusterRoleAndBinding(ctx, c, rolePrefix+name, subjects, rules, expectedLabels...)
		if err != nil {
			return nil, nil, err
		}
	}

	sat, err := CreateTokenForServiceAccount(ctx, c, sa, nil)
	if err != nil {
		return nil, nil, err
	}

	kcfg, err := CreateTokenKubeconfig(name, restCfg.Host, restCfg.CAData, sat.Token)
	if err != nil {
		return nil, nil, err
	}

	return kcfg, sat, nil
}

// EnsureNamespace ensures that the specified Namespace exists.
// If it doesn't exist, it is created with the expected labels.
// If it exists, but does not have the expected labels, a ResourceNotManagedError is returned.
// The namespace is returned.
func EnsureNamespace(ctx context.Context, c client.Client, nsName string, expectedLabels ...Label) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{}
	ns.SetName(nsName)
	found := true
	if err := c.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting Namespace '%s': %w", ns.Name, err)
		}
		found = false
	}
	if found {
		if err := FailIfNotManaged(ns, expectedLabels...); err != nil {
			return nil, err
		}
		// a namespace does not have any spec, so we don't have to do anything, if it was found
		return ns, nil
	}
	ns.SetLabels(LabelListToMap(expectedLabels))
	if err := c.Create(ctx, ns); err != nil {
		return nil, fmt.Errorf("error creating Namespace '%s': %w", ns.Name, err)
	}

	return ns, nil
}

// EnsureServiceAccount ensures that the specified ServiceAccount exists.
// If it doesn't exist, it is created with the expected labels (the namespace has to exist).
// If it exists, but does not have the expected labels, a ResourceNotManagedError is returned.
// The ServiceAccount is returned.
func EnsureServiceAccount(ctx context.Context, c client.Client, saName, saNamespace string, expectedLabels ...Label) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	sa.SetName(saName)
	sa.SetNamespace(saNamespace)
	found := true
	if err := c.Get(ctx, client.ObjectKeyFromObject(sa), sa); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ServiceAccount '%s/%s': %w", sa.Namespace, sa.Name, err)
		}
		found = false
	}
	if found {
		if err := FailIfNotManaged(sa, expectedLabels...); err != nil {
			return nil, err
		}
		// a serviceaccount does not have any relevant spec, so we don't have to do anything, if it was found
		return sa, nil
	}
	sa.SetLabels(LabelListToMap(expectedLabels))
	if err := c.Create(ctx, sa); err != nil {
		return nil, fmt.Errorf("error creating ServiceAccount '%s': %w", sa.Name, err)
	}

	return sa, nil
}

// EnsureClusterRoleAndBinding combines EnsureClusterRole and EnsureClusterRoleBinding.
// The name is used for both the ClusterRole and ClusterRoleBinding.
func EnsureClusterRoleAndBinding(ctx context.Context, c client.Client, name string, subjects []rbacv1.Subject, rules []rbacv1.PolicyRule, expectedLabels ...Label) (*rbacv1.ClusterRoleBinding, *rbacv1.ClusterRole, error) {
	cr, err := EnsureClusterRole(ctx, c, name, rules, expectedLabels...)
	if err != nil {
		return nil, nil, err
	}
	crb, err := EnsureClusterRoleBinding(ctx, c, name, cr.Name, subjects, expectedLabels...)
	if err != nil {
		return nil, cr, err
	}
	return crb, cr, nil
}

// EnsureClusterRole ensures that the specified ClusterRole exists with the specified rules.
// If it doesn't exist, it is created with the expected labels.
// If it exists, but does not have the expected labels, a ResourceNotManagedError is returned.
// The ClusterRole is returned.
func EnsureClusterRole(ctx context.Context, c client.Client, name string, rules []rbacv1.PolicyRule, expectedLabels ...Label) (*rbacv1.ClusterRole, error) {
	crm := resources.NewClusterRoleMutator(name, rules)
	crm.MetadataMutator().WithLabels(LabelListToMap(expectedLabels))
	cr := crm.Empty()
	found := true
	if err := c.Get(ctx, client.ObjectKeyFromObject(cr), cr); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ClusterRole '%s': %w", cr.Name, err)
		}
		found = false
	}
	if found {
		if err := FailIfNotManaged(cr, expectedLabels...); err != nil {
			return nil, err
		}
	}
	if err := resources.CreateOrUpdateResource(ctx, c, crm); err != nil {
		return nil, fmt.Errorf("error creating/updating ClusterRole '%s': %w", cr.Name, err)
	}
	return cr, nil
}

// EnsureClusterRoleBinding ensures that the specified ClusterRoleBinding exists with the specified subjects.
// If it doesn't exist, it is created with the expected labels.
// If it exists, but does not have the expected labels, a ResourceNotManagedError is returned.
// The ClusterRoleBinding is returned.
func EnsureClusterRoleBinding(ctx context.Context, c client.Client, name, clusterRoleName string, subjects []rbacv1.Subject, expectedLabels ...Label) (*rbacv1.ClusterRoleBinding, error) {
	crbm := resources.NewClusterRoleBindingMutator(name, subjects, resources.NewClusterRoleRef(clusterRoleName))
	crbm.MetadataMutator().WithLabels(LabelListToMap(expectedLabels))
	crb := crbm.Empty()
	found := true
	if err := c.Get(ctx, client.ObjectKeyFromObject(crb), crb); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ClusterRoleBinding '%s': %w", crb.Name, err)
		}
		found = false
	}
	if found {
		if err := FailIfNotManaged(crb, expectedLabels...); err != nil {
			return nil, err
		}
	}
	if err := resources.CreateOrUpdateResource(ctx, c, crbm); err != nil {
		return nil, fmt.Errorf("error creating/updating ClusterRole '%s': %w", crb.Name, err)
	}
	return crb, nil
}

// EnsureRoleAndBinding combines EnsureRole and EnsureRoleBinding.
// The name is used for both the Role and RoleBinding.
func EnsureRoleAndBinding(ctx context.Context, c client.Client, name, namespace string, subjects []rbacv1.Subject, rules []rbacv1.PolicyRule, expectedLabels ...Label) (*rbacv1.RoleBinding, *rbacv1.Role, error) {
	r, err := EnsureRole(ctx, c, name, namespace, rules, expectedLabels...)
	if err != nil {
		return nil, nil, err
	}
	rb, err := EnsureRoleBinding(ctx, c, name, namespace, r.Name, subjects, expectedLabels...)
	if err != nil {
		return nil, r, err
	}
	return rb, r, nil
}

// EnsureRole ensures that the specified Role exists with the specified rules.
// If it doesn't exist, it is created with the expected labels.
// If it exists, but does not have the expected labels, a ResourceNotManagedError is returned.
// The Role is returned.
func EnsureRole(ctx context.Context, c client.Client, name, namespace string, rules []rbacv1.PolicyRule, expectedLabels ...Label) (*rbacv1.Role, error) {
	rm := resources.NewRoleMutator(name, namespace, rules)
	rm.MetadataMutator().WithLabels(LabelListToMap(expectedLabels))
	r := rm.Empty()
	found := true
	if err := c.Get(ctx, client.ObjectKeyFromObject(r), r); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting Role '%s/%s': %w", r.Namespace, r.Name, err)
		}
		found = false
	}
	if found {
		if err := FailIfNotManaged(r, expectedLabels...); err != nil {
			return nil, err
		}
	}
	if err := resources.CreateOrUpdateResource(ctx, c, rm); err != nil {
		return nil, fmt.Errorf("error creating/updating Role '%s/%s': %w", r.Namespace, r.Name, err)
	}
	return r, nil
}

// EnsureRoleBinding ensures that the specified RoleBinding exists with the specified subjects.
// If it doesn't exist, it is created with the expected labels.
// If it exists, but does not have the expected labels, a ResourceNotManagedError is returned.
// The RoleBinding is returned.
func EnsureRoleBinding(ctx context.Context, c client.Client, name, namespace, roleName string, subjects []rbacv1.Subject, expectedLabels ...Label) (*rbacv1.RoleBinding, error) {
	rbm := resources.NewRoleBindingMutator(name, namespace, subjects, resources.NewRoleRef(roleName))
	rbm.MetadataMutator().WithLabels(LabelListToMap(expectedLabels))
	rb := rbm.Empty()
	found := true
	if err := c.Get(ctx, client.ObjectKeyFromObject(rb), rb); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting RoleBinding '%s/%s': %w", rb.Namespace, rb.Name, err)
		}
		found = false
	}
	if found {
		if err := FailIfNotManaged(rb, expectedLabels...); err != nil {
			return nil, err
		}
	}
	if err := resources.CreateOrUpdateResource(ctx, c, rbm); err != nil {
		return nil, fmt.Errorf("error creating/updating RoleBinding '%s/%s': %w", rb.Namespace, rb.Name, err)
	}
	return rb, nil
}

// CreateTokenForServiceAccount generates a token for the given ServiceAccount.
func CreateTokenForServiceAccount(ctx context.Context, c client.Client, sa *corev1.ServiceAccount, desiredDuration *time.Duration) (*ServiceAccountToken, error) {
	tr := &authenticationv1.TokenRequest{}
	if desiredDuration != nil {
		tr.Spec.ExpirationSeconds = ptr.To((int64)(desiredDuration.Seconds()))
	}

	sat := &ServiceAccountToken{
		CreationTimestamp: time.Now(),
	}
	if err := c.SubResource("token").Create(ctx, sa, tr); err != nil {
		return nil, fmt.Errorf("error creating token for ServiceAccount '%s/%s': %w", sa.Namespace, sa.Name, err)
	}
	sat.Token = tr.Status.Token
	sat.ExpirationTimestamp = tr.Status.ExpirationTimestamp.Time

	return sat, nil
}

// ServiceAccountToken is a helper struct that bundles a ServiceAccount token together with its creation and expiration timestamps.
type ServiceAccountToken struct {
	Token               string
	CreationTimestamp   time.Time
	ExpirationTimestamp time.Time
}

// CreateTokenKubeconfig generates a kubeconfig based on the given values.
// The 'user' arg is used as key for the auth configuration and can be chosen freely.
func CreateTokenKubeconfig(user, host string, caData []byte, token string) ([]byte, error) {
	id := "cluster"
	kcfg := clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			id: {
				Server:                   host,
				CertificateAuthorityData: caData,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			id: {
				Cluster:  id,
				AuthInfo: user,
			},
		},
		CurrentContext: id,
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			user: {
				Token: token,
			},
		},
	}

	kcfgBytes, err := clientcmd.Write(kcfg)
	if err != nil {
		return nil, fmt.Errorf("error converting converting generated kubeconfig into yaml: %w", err)
	}
	return kcfgBytes, nil
}

// ComputeTokenRenewalTime computes the time for the renewal of a token, given its creation and expiration time.
// Returns the zero time if either of the given times is zero.
// The returned time is when 80% of the validity duration is reached.
// If another percentage is desired, use ComputeTokenRenewalTimeWithRatio instead.
func ComputeTokenRenewalTime(creationTime, expirationTime time.Time) time.Time {
	return ComputeTokenRenewalTimeWithRatio(creationTime, expirationTime, 0.8)
}

// ComputeTokenRenewalTime computes the time for the renewal of a token, given its creation and expiration time.
// Returns the zero time if either of the given times is zero.
// Ratio must be between 0 and 1. The returned time is when this percentage of the validity duration is reached.
func ComputeTokenRenewalTimeWithRatio(creationTime, expirationTime time.Time, ratio float64) time.Time {
	if creationTime.IsZero() || expirationTime.IsZero() {
		return time.Time{}
	}
	// validity is how long the token was valid in the first place
	validity := expirationTime.Sub(creationTime)
	// renewalAfter is 80% of the validity
	renewalAfter := time.Duration(float64(validity) * ratio)
	// renewalAt is the point in time at which the token should be renewed
	renewalAt := creationTime.Add(renewalAfter)
	return renewalAt
}
