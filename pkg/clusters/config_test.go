package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
)

var _ = Describe("Clusters Suite", func() {
	It("should create an OIDC config", func() {
		c := clusters.New("test").WithConfigPath("./testdata/config/kubeconfig-token.yaml")
		Expect(c.InitializeRESTConfig()).ToNot(HaveOccurred())

		oidcConfigRaw, err := c.MarshalOIDCConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(oidcConfigRaw).ToNot(BeEmpty())

		var oidcConfig map[string]string
		Expect(yaml.Unmarshal(oidcConfigRaw, &oidcConfig)).ToNot(HaveOccurred())
		Expect(oidcConfig).To(HaveKeyWithValue("host", "https://test-server"))
		Expect(oidcConfig)
	})

	It("should create a kubeconfig with token", func() {
		c := clusters.New("test").WithConfigPath("./testdata/config/kubeconfig-token.yaml")
		Expect(c.InitializeRESTConfig()).ToNot(HaveOccurred())

		kubeconfigRaw, err := c.MarshalKubeconfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeconfigRaw).ToNot(BeEmpty())

		config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigRaw)
		Expect(err).ToNot(HaveOccurred())
		Expect(config.Host).To(Equal("https://test-server"))
		Expect(config.TLSClientConfig.CAData).ToNot(BeEmpty())
		Expect(config.BearerToken).To(Equal("dGVzdC10b2tlbg=="))
	})

	It("should create a kubeconfig with basic auth", func() {
		c := clusters.New("test").WithConfigPath("./testdata/config/kubeconfig-basicauth.yaml")
		Expect(c.InitializeRESTConfig()).ToNot(HaveOccurred())

		kubeconfigRaw, err := c.MarshalKubeconfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeconfigRaw).ToNot(BeEmpty())

		config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigRaw)
		Expect(err).ToNot(HaveOccurred())
		Expect(config.Host).To(Equal("https://test-server"))
		Expect(config.TLSClientConfig.CAData).ToNot(BeEmpty())
		Expect(config.Username).To(Equal("foo"))
		Expect(config.Password).To(Equal("bar"))
	})

	It("should create a kubeconfig with client tls", func() {
		c := clusters.New("test").WithConfigPath("./testdata/config/kubeconfig-tls.yaml")
		Expect(c.InitializeRESTConfig()).ToNot(HaveOccurred())

		kubeconfigRaw, err := c.MarshalKubeconfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeconfigRaw).ToNot(BeEmpty())

		config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigRaw)
		Expect(err).ToNot(HaveOccurred())
		Expect(config.Host).To(Equal("https://test-server"))
		Expect(config.TLSClientConfig.CAData).ToNot(BeEmpty())
		Expect(config.TLSClientConfig.CertData).ToNot(BeEmpty())
		Expect(config.TLSClientConfig.KeyData).ToNot(BeEmpty())
	})
})
