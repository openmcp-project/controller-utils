package crds

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"log"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	errSecretHasNoTLSCert = errors.New("secret is missing " + corev1.TLSCertKey)
)

const (
	webhookPath = "/convert"
)

func Install(ctx context.Context, c client.Client, crdFiles embed.FS, options ...installOption) error {
	opts := &installOptions{
		localClient:        c,
		remoteClient:       c,
		webhookService:     getWebhookServiceFromEnv(),
		webhookSecret:      getWebhookSecretFromEnv(),
		webhookServicePort: 443,
	}
	for _, io := range options {
		io.ApplyToInstallOptions(opts)
	}

	if !opts.noResolveCA {
		secret := &corev1.Secret{}
		if err := opts.localClient.Get(ctx, opts.webhookSecret, secret); err != nil {
			return err
		}

		if _, ok := secret.Data[corev1.TLSCertKey]; !ok {
			return errSecretHasNoTLSCert
		}

		opts.caData = secret.Data[corev1.TLSCertKey]
	}

	// make sure that the client knows about the CustomResourceDefinition type
	utilruntime.Must(apiextensionsv1.AddToScheme(c.Scheme()))

	log.Println("Reading CRD files")
	contents, err := readAllFiles(crdFiles, ".")
	if err != nil {
		return err
	}

	for _, v := range contents {
		crd, err := decodeYaml(v)
		if err != nil {
			return err
		}

		if err := applyCRD(ctx, opts, crd); err != nil {
			return err
		}
	}
	return nil
}

func decodeYaml(yamlBytes []byte) (*apiextensionsv1.CustomResourceDefinition, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlBytes), 100)
	crd := &apiextensionsv1.CustomResourceDefinition{}
	return crd, decoder.Decode(crd)
}

func applyCRD(ctx context.Context, opts *installOptions, crd *apiextensionsv1.CustomResourceDefinition) error {
	copy := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crd.Name,
		},
	}

	result, err := controllerutil.CreateOrUpdate(ctx, opts.remoteClient, copy, func() error {
		copy.Spec = crd.Spec
		applyConversionConfig(copy, opts)
		return nil
	})
	log.Println("CRD", crd.Name, result)
	return err
}

func applyConversionConfig(crd *apiextensionsv1.CustomResourceDefinition, opts *installOptions) {
	conv := crd.Spec.Conversion
	if conv != nil && conv.Strategy == apiextensionsv1.WebhookConverter {
		if conv.Webhook == nil {
			conv.Webhook = &apiextensionsv1.WebhookConversion{ConversionReviewVersions: []string{"v1"}}
		}

		conv.Webhook.ClientConfig = &apiextensionsv1.WebhookClientConfig{
			CABundle: opts.caData,
		}

		if opts.customBaseUrl != nil {
			conv.Webhook.ClientConfig.URL = ptr.To(*opts.customBaseUrl + webhookPath)
		} else {
			conv.Webhook.ClientConfig.Service = &apiextensionsv1.ServiceReference{
				Name:      opts.webhookService.Name,
				Namespace: opts.webhookService.Namespace,
				Path:      ptr.To(webhookPath),
				Port:      &opts.webhookServicePort,
			}
		}
	}
}
