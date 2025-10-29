package webhooks

import (
	"context"
	"fmt"
	"log"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/openmcp-project/controller-utils/pkg/collections/maps"
)

func applyValidatingWebhook(ctx context.Context, opts *installOptions, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, opts.scheme)
	if err != nil {
		return err
	}
	webhookPath := generateValidatePath(gvk)

	cfg := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateValidateName(gvk),
		},
	}

	resource := strings.ToLower(gvk.Kind + "s")

	result, err := controllerutil.CreateOrUpdate(ctx, opts.remoteClient, cfg, func() error {
		cfg.Labels = maps.Merge(cfg.Labels, opts.managedLabels)
		webhook := admissionregistrationv1.ValidatingWebhook{
			Name:                    strings.ToLower(fmt.Sprintf("v%s.%s", gvk.Kind, gvk.Group)),
			FailurePolicy:           ptr.To(admissionregistrationv1.Fail),
			SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
			AdmissionReviewVersions: []string{"v1"},
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: opts.caData,
			},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
						admissionregistrationv1.Delete,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{gvk.Group},
						APIVersions: []string{gvk.Version},
						Resources:   []string{resource},
					},
				},
			},
		}

		if opts.customBaseUrl != nil {
			webhook.ClientConfig.URL = ptr.To(*opts.customBaseUrl + webhookPath)
		} else {
			webhook.ClientConfig.Service = &admissionregistrationv1.ServiceReference{
				Name:      opts.webhookService.Name,
				Namespace: opts.webhookService.Namespace,
				Path:      ptr.To(webhookPath),
				Port:      &opts.webhookServicePort,
			}
		}

		cfg.Webhooks = []admissionregistrationv1.ValidatingWebhook{webhook}
		return nil
	})
	log.Println("Validating webhook config", cfg.Name, result)
	return err
}

func applyMutatingWebhook(ctx context.Context, opts *installOptions, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, opts.scheme)
	if err != nil {
		return err
	}
	webhookPath := generateMutatePath(gvk)

	cfg := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateMutateName(gvk),
		},
	}

	resource := strings.ToLower(gvk.Kind + "s")

	result, err := controllerutil.CreateOrUpdate(ctx, opts.remoteClient, cfg, func() error {
		cfg.Labels = maps.Merge(cfg.Labels, opts.managedLabels)
		webhook := admissionregistrationv1.MutatingWebhook{
			Name:                    strings.ToLower(fmt.Sprintf("m%s.%s", gvk.Kind, gvk.Group)),
			FailurePolicy:           ptr.To(admissionregistrationv1.Fail),
			SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
			AdmissionReviewVersions: []string{"v1"},
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: opts.caData,
			},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{gvk.Group},
						APIVersions: []string{gvk.Version},
						Resources:   []string{resource},
					},
				},
			},
		}

		if opts.customBaseUrl == nil {
			webhook.ClientConfig.Service = &admissionregistrationv1.ServiceReference{
				Name:      opts.webhookService.Name,
				Namespace: opts.webhookService.Namespace,
				Path:      ptr.To(webhookPath),
				Port:      &opts.webhookServicePort,
			}
		} else {
			webhook.ClientConfig.URL = ptr.To(*opts.customBaseUrl + webhookPath)
		}

		cfg.Webhooks = []admissionregistrationv1.MutatingWebhook{webhook}
		return nil
	})
	log.Println("Mutating webhook config", cfg.Name, result)
	return err
}

func applyWebhookService(ctx context.Context, opts *installOptions) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.webhookService.Name,
			Namespace: opts.webhookService.Namespace,
		},
	}

	result, err := controllerutil.CreateOrUpdate(ctx, opts.localClient, svc, func() error {
		svc.Labels = maps.Merge(svc.Labels, opts.managedLabels)
		svc.Spec.Selector = opts.managedService.SelectorLabels
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "https",
				Protocol:   corev1.ProtocolTCP,
				Port:       opts.webhookServicePort,
				TargetPort: opts.managedService.TargetPort,
			},
		}
		return nil
	})
	log.Println("Webhook service", types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}.String(), result)
	return err
}
