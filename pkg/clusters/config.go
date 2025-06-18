package clusters

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

// oidcTrustConfig represents the configuration for an OIDC trust relationship.
// It includes the host of the Kubernetes API server, CA data for TLS verification,
// and the audience for the OIDC tokens.
type oidcTrustConfig struct {
	// Host is the URL of the Kubernetes API server.
	Host string `json:"host,omitempty"`
	// CAData is the base64-encoded CA certificate data used to verify the server's TLS certificate.
	CAData []byte `json:"caData,omitempty"`
}

// MarshalOIDCConfig converts the Cluster's RESTConfig to an OIDC trust configuration format.
// It returns an error if the RESTConfig is not set or if marshalling fails.
// When creating a Kubernetes deployment, this configuration is used to set up the trust relationship to
// the target cluster.
// Example:
//
// spec:
//
//	 template:
//	   spec:
//		   volumes:
//		     - name: oidc-trust-config
//		       projected:
//		         sources:
//		           - secret:
//		               name: oidc-trust-config
//		               items:
//		                 - key: host
//		                   path: cluster/host
//		                 - key: caData
//	                       path: cluster/ca.crt
//		           - serviceAccountToken:
//		               audience: target-cluster
//	                   path: cluster/token
//	                   expirationSeconds: 3600
//
//		   volumeMounts:
//		     - name: oidc-trust-config
//		       mountPath: /var/run/secrets/oidc-trust-config
//		       readOnly: true
func (c *Cluster) MarshalOIDCConfig() ([]byte, error) {
	if c.RESTConfig() == nil {
		return nil, fmt.Errorf("cannot marshal OIDC trust config when RESTConfig is nil")
	}

	restConfig := c.RESTConfig()

	oidcConfig := &oidcTrustConfig{
		Host:   restConfig.Host,
		CAData: restConfig.CAData,
	}

	configMarshaled, err := yaml.Marshal(oidcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OIDC trust config: %w", err)
	}

	return configMarshaled, nil
}

// MarshalKubeconfig converts the Cluster to a kubeconfig format.
// It returns an error if the RESTConfig is not set or does not contain authentication information.
// Supported authentication methods are Bearer Token, Username/Password and Client Certificate.
func (c *Cluster) MarshalKubeconfig() ([]byte, error) {
	var contextName string
	var authInfo *clientapi.AuthInfo

	if c.RESTConfig() == nil {
		return nil, fmt.Errorf("cannot marshal to config when RESTConfig is nil")
	}

	restConfig := c.RESTConfig()

	if c.HasID() {
		contextName = c.ID()
	} else {
		contextName = "default"
	}

	type authType string
	const (
		authTypeBearerToken authType = "BearerToken"
		authTypeBasicAuth   authType = "BasicAuth"
		authTypeClientCert  authType = "ClientCert"
	)
	availableAuthTypes := make(map[authType]interface{}, 0)
	if restConfig.BearerToken != "" {
		availableAuthTypes[authTypeBearerToken] = nil
	}

	if restConfig.Username != "" && restConfig.Password != "" {
		availableAuthTypes[authTypeBasicAuth] = nil
	}

	if restConfig.TLSClientConfig.CertData != nil && restConfig.TLSClientConfig.KeyData != nil {
		availableAuthTypes[authTypeClientCert] = nil
	}

	if len(availableAuthTypes) == 0 {
		return nil, fmt.Errorf("cannot marshal to config when RESTConfig does not contain any supported authentication information")
	}

	if _, ok := availableAuthTypes[authTypeBearerToken]; ok {
		authInfo = &clientapi.AuthInfo{
			Token: restConfig.BearerToken,
		}
	}

	if _, ok := availableAuthTypes[authTypeBasicAuth]; ok {
		authInfo = &clientapi.AuthInfo{
			Username: restConfig.Username,
			Password: restConfig.Password,
		}
	}

	if _, ok := availableAuthTypes[authTypeClientCert]; ok {
		authInfo = &clientapi.AuthInfo{
			ClientCertificateData: restConfig.TLSClientConfig.CertData,
			ClientKeyData:         restConfig.TLSClientConfig.KeyData,
		}
	}

	server := restConfig.Host
	if restConfig.APIPath != "" {
		server = fmt.Sprint(server, "/", restConfig.APIPath)
	}

	kubeConfig := clientapi.Config{
		CurrentContext: contextName,
		Contexts: map[string]*clientapi.Context{
			contextName: {
				AuthInfo: contextName,
				Cluster:  contextName,
			},
		},
		Clusters: map[string]*clientapi.Cluster{
			contextName: {
				Server:                   server,
				CertificateAuthorityData: restConfig.CAData,
			},
		},
		AuthInfos: map[string]*clientapi.AuthInfo{
			contextName: authInfo,
		},
	}

	configMarshaled, err := clientcmd.Write(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cluster to config: %w", err)
	}

	return configMarshaled, nil
}
