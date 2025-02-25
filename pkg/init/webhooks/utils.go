package webhooks

import (
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func generateMutateName(gvk schema.GroupVersionKind) string {
	return "mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func generateValidatePath(gvk schema.GroupVersionKind) string {
	return "/validate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func generateValidateName(gvk schema.GroupVersionKind) string {
	return "validate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func getWebhookServiceFromEnv() types.NamespacedName {
	return types.NamespacedName{
		Name:      os.Getenv("WEBHOOK_SERVICE_NAME"),
		Namespace: os.Getenv("WEBHOOK_SERVICE_NAMESPACE"),
	}
}

func getWebhookSecretFromEnv() types.NamespacedName {
	return types.NamespacedName{
		Name:      os.Getenv("WEBHOOK_SECRET_NAME"),
		Namespace: os.Getenv("WEBHOOK_SECRET_NAMESPACE"),
	}
}
