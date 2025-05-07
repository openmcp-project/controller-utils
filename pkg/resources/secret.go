package resources

import (
	"fmt"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretMutator struct {
	Name      string
	Namespace string
	Data      map[string][]byte
	Type      core.SecretType
	meta      Mutator[client.Object]
}

var _ Mutator[*core.Secret] = &SecretMutator{}

func NewSecretMutator(name, namespace string, data map[string][]byte, secretType core.SecretType, labels map[string]string, annotations map[string]string) Mutator[*core.Secret] {
	return &SecretMutator{
		Name:      name,
		Namespace: namespace,
		Data:      data,
		Type:      secretType,
		meta:      NewMetadataMutator(labels, annotations),
	}
}

func (m *SecretMutator) String() string {
	return fmt.Sprintf("secret %s/%s", m.Namespace, m.Name)
}

func (m *SecretMutator) Empty() *core.Secret {
	return &core.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Data: m.Data,
		Type: m.Type,
	}
}

func (m *SecretMutator) Mutate(s *core.Secret) error {
	if s.Data == nil {
		s.Data = make(map[string][]byte)
	}
	for key, value := range m.Data {
		s.Data[key] = value
	}
	return m.meta.Mutate(s)
}
