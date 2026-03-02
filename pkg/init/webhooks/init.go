package webhooks

import (
	"context"
	"errors"
	"log"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// APITypes defines an API type along with whether it has a validator and/or defaulter webhook.
type APITypes struct {
	// Obj is the API type object.
	Obj client.Object
	// Validator indicates whether the type has a validating webhook.
	Validator bool
	// Defaulter indicates whether the type has a mutating webhook.
	Defaulter bool
	// Mutation allows to mutate the webhooks before they are created or updated. Use with caution, as it may break the webhooks or interfere with the library's management of the resources.
	Mutation Mutation
}

// ValidatingMutation is a function that can be used to mutate the validating webhook before it is created or updated.
// Use with caution, as it may break the webhook or interfere with the library's management of the webhook.
type ValidatingMutation func(webhook *admissionregistrationv1.ValidatingWebhook) error

// MutatingMutation is a function that can be used to mutate the mutating webhook before it is created or updated.
// Use with caution, as it may break the webhook or interfere with the library's management of the webhook.
type MutatingMutation func(webhook *admissionregistrationv1.MutatingWebhook) error

type Mutation struct {
	ValidatingWebhook ValidatingMutation
	MutatingWebhook   MutatingMutation
}

func Install(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	apiTypes []APITypes,
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

	for _, t := range apiTypes {
		if t.Validator {
			if err := applyValidatingWebhook(ctx, opts, t.Obj, t.Mutation.ValidatingWebhook); err != nil {
				return err
			}
		}
		if t.Defaulter {
			if err := applyMutatingWebhook(ctx, opts, t.Obj, t.Mutation.MutatingWebhook); err != nil {
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
	apiTypes []APITypes,
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

	for _, t := range apiTypes {
		if t.Validator {
			if err := removeValidatingWebhook(ctx, opts, t.Obj); err != nil {
				return err
			}
		}
		if t.Defaulter {
			if err := removeMutatingWebhook(ctx, opts, t.Obj); err != nil {
				return err
			}
		}
	}

	log.Println("Webhooks removed")
	return nil
}
