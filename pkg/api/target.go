// +kubebuilder:object:generate=true
package api

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type Target struct {
	// Kubeconfig is an inline kubeconfig.
	// +kubebuilder:pruning:PreserveUnknownFields
	Kubeconfig *apiextensionsv1.JSON `json:"kubeconfig,omitempty"`

	// KubeconfigFile is a path to a file containing a kubeconfig.
	KubeconfigFile *string `json:"kubeconfigFile,omitempty"`

	// KubeconfigRef is a reference to a Kubernetes secret that contains a kubeconfig.
	KubeconfigRef *KubeconfigReference `json:"kubeconfigRef,omitempty"`

	// ServiceAccount references a local service account.
	ServiceAccount *ServiceAccountConfig `json:"serviceAccount,omitempty"`
}

type KubeconfigReference struct {
	corev1.SecretReference `json:",inline"`

	// The key of the secret to select from.  Must be a valid secret key.
	// +kubebuilder:default="kubeconfig"
	Key string `json:"key"`
}

type ServiceAccountConfig struct {
	// Name is the name of the service account.
	// This value is optional. If not provided, the pod's service account will be used.
	Name string `json:"name,omitempty"`

	// Namespace is the name of the service account.
	// This value is optional. If not provided, the pod's service account will be used.
	Namespace string `json:"namespace,omitempty"`

	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	// This value is optional. If not provided, the local API server will be used.
	Host string `json:"host,omitempty"`

	// CAFile points to a file containing the root certificates for the API server.
	// This value is optional. If not provided, the value of CAData will be used.
	CAFile *string `json:"caFile,omitempty"`

	// CAData holds (Base64-)PEM-encoded bytes.
	// CAData takes precedence over CAFile.
	// This value is optional. If not provided, the CAData of the in-cluster config will be used.
	// Providing an empty string means that the operating system's defaults root certificates will be used.
	CAData *string `json:"caData,omitempty"`

	// TokenFile points to a file containing a bearer token (e.g. projected service account token (PSAT) with custom audience) to be used for authentication against the API server.
	// If provided, all other authentication methods (Basic, client-side TLS, etc.) will be disabled.
	TokenFile string `json:"tokenFile,omitempty"`
}
