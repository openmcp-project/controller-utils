package resources

import (
	"fmt"

	"maps"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretMutator struct {
	Name       string
	Namespace  string
	Data       map[string][]byte
	StringData map[string]string
	Type       core.SecretType
	meta       MetadataMutator
}

var _ Mutator[*core.Secret] = &SecretMutator{}

func NewSecretMutator(name, namespace string, data map[string][]byte, secretType core.SecretType) Mutator[*core.Secret] {
	return &SecretMutator{
		Name:      name,
		Namespace: namespace,
		Data:      data,
		Type:      secretType,
		meta:      NewMetadataMutator(),
	}
}

func NewSecretMutatorWithStringData(name, namespace string, stringData map[string]string, secretType core.SecretType) Mutator[*core.Secret] {
	return &SecretMutator{
		Name:       name,
		Namespace:  namespace,
		StringData: stringData,
		Type:       secretType,
		meta:       NewMetadataMutator(),
	}
}

func (m *SecretMutator) String() string {
	return fmt.Sprintf("secret %s/%s", m.Namespace, m.Name)
}

func (m *SecretMutator) Empty() *core.Secret {
	s := &core.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Type: m.Type,
	}
	if m.Data != nil {
		s.Data = make(map[string][]byte, len(m.Data))
		maps.Copy(s.Data, m.Data)
	}
	if m.StringData != nil {
		s.StringData = make(map[string]string, len(m.StringData))
		maps.Copy(s.StringData, m.StringData)
	}
	return s
}

func (m *SecretMutator) Mutate(s *core.Secret) error {
	if m.Data != nil {
		s.Data = make(map[string][]byte, len(m.Data))
		maps.Copy(s.Data, m.Data)
	}
	if m.StringData != nil {
		s.StringData = make(map[string]string, len(m.StringData))
		maps.Copy(s.StringData, m.StringData)
	}
	return m.meta.Mutate(s)
}

func (m *SecretMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
