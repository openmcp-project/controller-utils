package resources

import (
	"fmt"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapMutator struct {
	Name      string
	Namespace string
	Data      map[string]string
	meta      MetadataMutator
}

var _ Mutator[*core.ConfigMap] = &ConfigMapMutator{}

func NewConfigMapMutator(name, namespace string, data map[string]string, labels map[string]string, annotations map[string]string) Mutator[*core.ConfigMap] {
	return &ConfigMapMutator{
		Name:      name,
		Namespace: namespace,
		Data:      data,
		meta:      NewMetadataMutator(labels, annotations),
	}
}

func (m *ConfigMapMutator) String() string {
	return fmt.Sprintf("configmap %s/%s", m.Namespace, m.Name)
}

func (m *ConfigMapMutator) Empty() *core.ConfigMap {
	return &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Data: m.Data,
	}
}

func (m *ConfigMapMutator) Mutate(cm *core.ConfigMap) error {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	for key, value := range m.Data {
		cm.Data[key] = value
	}
	return m.meta.Mutate(cm)
}

func (m *ConfigMapMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
