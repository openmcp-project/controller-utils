package externalsecrets

import (
	"context"
	"fmt"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
	openmcpresources "github.com/openmcp-project/controller-utils/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretCopyConfig holds the configuration for copying a secret.
type SecretCopyConfig struct {
	// SourceClient is the client to read the source secret from.
	SourceClient client.Client
	// SourceNamespace is the namespace of the source secret.
	SourceNamespace string
	// TargetNamespace is the namespace of the target secret.
	TargetNamespace string
	// TargetName is the name of the target secret.
	TargetName string
}

const secretNamePrefix = "sp-eso-"

// ManagePullSecret syncs every image pull secret the to cluster
func ManagePullSecret(targetCluster ManagedCluster, pullSecret corev1.LocalObjectReference, config SecretCopyConfig) {
	secret := NewManagedObject(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.TargetName,
			Namespace: config.TargetNamespace,
		},
	}, ManagedObjectContext{
		ReconcileFunc: func(ctx context.Context, o client.Object) error {
			oSecret, ok := o.(*corev1.Secret)
			if !ok {
				return fmt.Errorf("expected *corev1.Secret, got %T", o)
			}
			sourceSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pullSecret.Name,
					Namespace: config.SourceNamespace,
				},
			}
			// retrieve source secret from platform cluster
			if err := config.SourceClient.Get(ctx, client.ObjectKeyFromObject(sourceSecret), sourceSecret); err != nil {
				return err
			}
			mutator := openmcpresources.NewSecretMutator(config.TargetName, config.TargetNamespace, sourceSecret.Data, corev1.SecretTypeDockerConfigJson)
			return mutator.Mutate(oSecret)
		},
		StatusFunc: SimpleStatus,
	})
	targetCluster.AddObject(secret)
}

// PrefixSecretName adds the "sp-eso-" prefix to the given secret name
// to prevent name collisions in namespaces where multiple service providers operate.
// If the resulting name exceeds 63 characters (K8s limit), it will be truncated
// and a hash suffix appended for uniqueness via ShortenToXCharacters.
func PrefixSecretName(secretName string) (string, error) {
	return ctrlutils.ShortenToXCharacters(fmt.Sprintf("%s%s", secretNamePrefix, secretName), ctrlutils.K8sMaxNameLength)
}

// NewSecretCleaner removes redundant pull secrets in the given target namespace
// by removing any secret labeled as managed by sp-external-secrets that is not in secretsToKeep.
func NewSecretCleaner(cluster ManagedCluster, namespace string, secretsToKeep []corev1.LocalObjectReference) OrphanCleaner {
	return NewOrphanCleaner(cluster, namespace, cleanerType[*corev1.SecretList]{
		EmptyList: func() *corev1.SecretList {
			return &corev1.SecretList{}
		},
		ObjectsToKeep: secretsToKeep,
	})
}
