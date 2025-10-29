package webhooks

import (
	"context"
	"errors"
	"log"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	errSecretHasNoTLSCert = errors.New("secret is missing " + corev1.TLSCertKey)
)

// GenerateCertificate
func GenerateCertificate(ctx context.Context, c client.Client, options ...CertOption) error {
	opts := &certOptions{
		webhookService: getWebhookServiceFromEnv(),
		webhookSecret:  getWebhookSecretFromEnv(),
	}
	for _, co := range options {
		co.ApplyToCertOptions(opts)
	}

	log.Printf("Webhook Service: %+v", opts.webhookService)
	log.Printf("Webhook Secret: %+v", opts.webhookSecret)

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      opts.webhookSecret.Name,
			Namespace: opts.webhookSecret.Namespace,
		},
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(secret), secret); client.IgnoreNotFound(err) != nil {
		return err
	}

	if _, ok := secret.Data[corev1.TLSCertKey]; ok {
		// cert exists
		log.Println("Webhook secret exists. Doing nothing.")
		return nil
	}

	log.Println("Generating webhook certificate...")
	cert, err := generateCert(opts.webhookService, opts.additionalDNSNames)
	if err != nil {
		return err
	}

	log.Println("Storing webhook certificate in secret...")
	result, err := controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}

		secret.Data[corev1.TLSPrivateKeyKey] = cert.privateKey
		secret.Data[corev1.TLSCertKey] = cert.publicKey
		return nil
	})
	log.Println("Webhook secret", client.ObjectKeyFromObject(secret), result)
	return err
}

func Install(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	apiTypes []client.Object,
	options ...InstallOption,
) error {
	opts := &installOptions{
		localClient:        c,
		remoteClient:       c,
		scheme:             scheme,
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

	if opts.managedService != nil {
		if err := applyWebhookService(ctx, opts); err != nil {
			return err
		}
	}

	for _, o := range apiTypes {
		_, isCustomValidator := o.(webhook.CustomValidator)
		if isCustomValidator {
			if err := applyValidatingWebhook(ctx, opts, o); err != nil {
				return err
			}
		}
		_, isCustomDefaulter := o.(webhook.CustomDefaulter)
		if isCustomDefaulter {
			if err := applyMutatingWebhook(ctx, opts, o); err != nil {
				return err
			}
		}
	}

	log.Println("Webhooks initialized")
	return nil
}

func Uninstall(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	apiTypes []client.Object,
	options ...InstallOption,
) error {
	opts := &installOptions{
		localClient:        c,
		remoteClient:       c,
		scheme:             scheme,
		webhookService:     getWebhookServiceFromEnv(),
		webhookSecret:      getWebhookSecretFromEnv(),
		webhookServicePort: 443,
	}
	for _, io := range options {
		io.ApplyToInstallOptions(opts)
	}

	if opts.managedService != nil {
		if err := removeWebhookService(ctx, opts); err != nil {
			return err
		}
	}

	for _, o := range apiTypes {
		_, isCustomValidator := o.(webhook.CustomValidator)
		if isCustomValidator {
			if err := removeValidatingWebhook(ctx, opts, o); err != nil {
				return err
			}
		}
		_, isCustomDefaulter := o.(webhook.CustomDefaulter)
		if isCustomDefaulter {
			if err := removeMutatingWebhook(ctx, opts, o); err != nil {
				return err
			}
		}
	}

	log.Println("Webhooks removed")
	return nil
}
