package externalsecrets

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretName      = "privateregcred"
	sourceNamespace = "source"
	targetNamespace = "target"
)

func TestManagePullSecrets(t *testing.T) {
	fakeCluster := CreateFakeCluster(t, "platform", &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: sourceNamespace,
		},
		Data: map[string][]byte{
			"test": []byte("testdata"),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, &corev1.Secret{
		// existing secret with the same source secret name in the target namespace
		// this secret needs to be untouched by the service provider secret copy functionality
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: targetNamespace,
		},
		Data: map[string][]byte{
			"existing-secret-data": []byte("must-not-be-altered"),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	})
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		targetCluster ManagedCluster
		pullSecret    corev1.LocalObjectReference
		config        SecretCopyConfig
	}{
		{
			name:          "copy secret privateregcred from source to target namespace and adjust its name",
			targetCluster: NewManagedCluster(fakeCluster, &rest.Config{}, sourceNamespace, PlatformCluster),
			pullSecret:    corev1.LocalObjectReference{Name: secretName},
			config: SecretCopyConfig{
				SourceClient:    fakeCluster.Client(),
				SourceNamespace: sourceNamespace,
				TargetNamespace: targetNamespace,
				TargetName:      fmt.Sprintf("%s%s", secretNamePrefix, secretName),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ManagePullSecret(tt.targetCluster, tt.pullSecret, tt.config)
			ExecApply(t, []ManagedCluster{tt.targetCluster}, 1, []string{})
			sourceSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.pullSecret.Name,
					Namespace: tt.config.SourceNamespace,
				},
			}
			targetSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.config.TargetName,
					Namespace: tt.config.TargetNamespace,
				},
			}
			existingSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: tt.config.TargetNamespace,
				},
			}
			require.NoError(t, tt.config.SourceClient.Get(context.TODO(), client.ObjectKeyFromObject(sourceSecret), sourceSecret))
			require.NoError(t, tt.config.SourceClient.Get(context.TODO(), client.ObjectKeyFromObject(targetSecret), targetSecret))
			require.NoError(t, tt.config.SourceClient.Get(context.TODO(), client.ObjectKeyFromObject(existingSecret), existingSecret))
			assert.Equal(t, sourceSecret.Data, targetSecret.Data)
			assert.Equal(t, map[string][]byte{"existing-secret-data": []byte("must-not-be-altered")}, existingSecret.Data)
			assert.Equal(t, corev1.SecretTypeDockerConfigJson, targetSecret.Type, "target secret should have the correct type")
		})
	}
}

func TestPrefixSecretName(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short name", "privateregcred"},
		{"long name truncated", strings.Repeat("a", 60)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrefixSecretName(tt.input)
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(got, secretNamePrefix))
			assert.LessOrEqual(t, len(got), 63)
		})
	}
}
