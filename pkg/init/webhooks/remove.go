package webhooks

import (
	"context"
	"log"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func removeValidatingWebhook(ctx context.Context, opts *installOptions, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, opts.scheme)
	if err != nil {
		return err
	}

	cfg := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateValidateName(gvk),
		},
	}

	err = opts.remoteClient.Delete(ctx, cfg)
	log.Println("Removing validating webhook config", cfg.Name)
	return client.IgnoreNotFound(err)
}

func removeMutatingWebhook(ctx context.Context, opts *installOptions, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, opts.scheme)
	if err != nil {
		return err
	}

	cfg := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateMutateName(gvk),
		},
	}

	err = opts.remoteClient.Delete(ctx, cfg)
	log.Println("Removing mutating webhook config", cfg.Name)
	return client.IgnoreNotFound(err)
}

func removeWebhookService(ctx context.Context, opts *installOptions) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.webhookService.Name,
			Namespace: opts.webhookService.Namespace,
		},
	}

	err := opts.localClient.Delete(ctx, svc)
	log.Println("Removing webhook service", types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}.String())
	return err
}
