package clientconfig

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/api"
)

const (
	pemPrefix = "-----BEGIN"
)

var (
	scheme = runtime.NewScheme()

	ErrInvalidConnectionMethod      = errors.New("exactly one connection method has to be specified")
	ErrServiceAccountNamespaceEmpty = errors.New("service account namespace must be specified")

	reloadNoOp ReloadFunc = func() error { return nil }
)

type ReloadFunc func() error

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func New(target api.Target) *Config {
	return &Config{Target: target}
}

type Config struct {
	api.Target
}

func (c *Config) validate() error {
	methods := 0
	if c.Kubeconfig != nil {
		methods++
	}
	if c.KubeconfigFile != nil {
		methods++
	}
	if c.KubeconfigRef != nil {
		methods++
	}
	if c.ServiceAccount != nil {
		methods++
	}

	if methods != 1 {
		return ErrInvalidConnectionMethod
	}

	return nil
}

// GetRESTConfig creates a *rest.Config for the given API target.
// The second return value is a function which can be used to reload the config.
// This reload func is a no-op for "Kubeconfig" and "ServiceAccount" target types.
func (c *Config) GetRESTConfig() (*rest.Config, ReloadFunc, error) {
	if err := c.validate(); err != nil {
		return nil, nil, err
	}

	if c.Kubeconfig != nil {
		return c.handleKubeconfig()
	}

	if c.KubeconfigFile != nil {
		return c.handleKubeconfigFile()
	}

	if c.KubeconfigRef != nil {
		return c.handleKubeconfigRef()
	}

	if c.ServiceAccount != nil {
		return c.handleServiceAccount()
	}

	return nil, nil, ErrInvalidConnectionMethod
}

func (c *Config) handleKubeconfig() (*rest.Config, ReloadFunc, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(c.Kubeconfig.Raw)
	return config, reloadNoOp, err
}

func (c *Config) handleKubeconfigFile() (*rest.Config, ReloadFunc, error) {
	remoteConfig := &rest.Config{}

	reloadFunc := func() error {
		configBytes, err := os.ReadFile(*c.KubeconfigFile)
		if err != nil {
			return err
		}

		config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
		if err != nil {
			return err
		}

		copyRestConfig(config, remoteConfig)
		return nil
	}

	return remoteConfig, reloadFunc, reloadFunc()
}

func (c *Config) handleKubeconfigRef() (*rest.Config, ReloadFunc, error) {
	inClusterConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	inClusterClient, err := client.New(inClusterConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, nil, err
	}

	remoteConfig := &rest.Config{}

	reloadFunc := func() error {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.KubeconfigRef.Name,
				Namespace: c.KubeconfigRef.Namespace,
			},
		}

		if err := inClusterClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret); err != nil {
			return err
		}

		config, err := clientcmd.RESTConfigFromKubeConfig(secret.Data[c.KubeconfigRef.Key])
		if err != nil {
			return err
		}

		copyRestConfig(config, remoteConfig)
		return nil
	}

	return remoteConfig, reloadFunc, reloadFunc()
}

func (c *Config) handleServiceAccount() (*rest.Config, ReloadFunc, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	if c.ServiceAccount.TokenFile != "" {
		// Disable other authentication types
		cfg = rest.AnonymousClientConfig(cfg)

		cfg.BearerTokenFile = c.ServiceAccount.TokenFile
	}

	if c.ServiceAccount.Name != "" {
		if c.ServiceAccount.Namespace == "" {
			return nil, nil, ErrServiceAccountNamespaceEmpty
		}

		cfg.Impersonate = rest.ImpersonationConfig{
			UserName: fmt.Sprintf("system:serviceaccount:%s:%s", c.ServiceAccount.Name, c.ServiceAccount.Namespace),
		}
	}

	if c.ServiceAccount.Host != "" {
		cfg.APIPath = ""
		cfg.Host = c.ServiceAccount.Host
	}

	if c.ServiceAccount.CAFile != nil {
		cfg.CAFile = *c.ServiceAccount.CAFile
	}

	if c.ServiceAccount.CAData != nil {
		// Check if CAData is raw (not base64-encoded)
		if strings.HasPrefix(strings.TrimSpace(*c.ServiceAccount.CAData), pemPrefix) {
			cfg.CAData = []byte(*c.ServiceAccount.CAData)
		} else {
			decoded, err := base64.StdEncoding.DecodeString(*c.ServiceAccount.CAData)
			if err != nil {
				return nil, nil, err
			}
			cfg.CAData = decoded
		}
	}

	return cfg, reloadNoOp, nil
}

// GetClient creates a client.Client for the given API target.
// The second return value is a function which can be used to reload the config.
// This reload func is a no-op for "Kubeconfig" and "ServiceAccount" target types.
func (c *Config) GetClient(options client.Options) (client.Client, ReloadFunc, error) {
	restConfig, reloadFunc, err := c.GetRESTConfig()
	if err != nil {
		return nil, nil, err
	}

	client, err := client.New(restConfig, options)
	return client, reloadFunc, err
}

// copyRestConfig copies all fields from one *rest.Config to the other.
// rest.CopyConfig was used as a template.
func copyRestConfig(from, to *rest.Config) {
	to.Host = from.Host
	to.APIPath = from.APIPath
	to.ContentConfig = from.ContentConfig
	to.Username = from.Username
	to.Password = from.Password
	to.BearerToken = from.BearerToken
	to.BearerTokenFile = from.BearerTokenFile
	to.Impersonate.UserName = from.Impersonate.UserName
	to.Impersonate.UID = from.Impersonate.UID
	to.Impersonate.Groups = from.Impersonate.Groups
	to.Impersonate.Extra = from.Impersonate.Extra
	to.AuthProvider = from.AuthProvider
	to.AuthConfigPersister = from.AuthConfigPersister
	to.ExecProvider = from.ExecProvider
	to.TLSClientConfig.Insecure = from.TLSClientConfig.Insecure
	to.TLSClientConfig.ServerName = from.TLSClientConfig.ServerName
	to.TLSClientConfig.CertFile = from.TLSClientConfig.CertFile
	to.TLSClientConfig.KeyFile = from.TLSClientConfig.KeyFile
	to.TLSClientConfig.CAFile = from.TLSClientConfig.CAFile
	to.TLSClientConfig.CertData = from.TLSClientConfig.CertData
	to.TLSClientConfig.KeyData = from.TLSClientConfig.KeyData
	to.TLSClientConfig.CAData = from.TLSClientConfig.CAData
	to.TLSClientConfig.NextProtos = from.TLSClientConfig.NextProtos
	to.UserAgent = from.UserAgent
	to.DisableCompression = from.DisableCompression
	to.Transport = from.Transport
	to.WrapTransport = from.WrapTransport
	to.QPS = from.QPS
	to.Burst = from.Burst
	to.RateLimiter = from.RateLimiter
	to.WarningHandler = from.WarningHandler
	to.Timeout = from.Timeout
	to.Dial = from.Dial
	to.Proxy = from.Proxy
}
