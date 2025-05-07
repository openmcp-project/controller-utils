package resources

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccountMutator struct {
	Name      string
	Namespace string
	meta      Mutator[client.Object]
}

var _ Mutator[*core.ServiceAccount] = &ServiceAccountMutator{}

func NewServiceAccountMutator(name, namespace string, labels map[string]string, annotations map[string]string) Mutator[*core.ServiceAccount] {
	return &ServiceAccountMutator{
		Name:      name,
		Namespace: namespace,
		meta:      NewMetadataMutator(labels, annotations),
	}
}

func (m *ServiceAccountMutator) String() string {
	return fmt.Sprintf("service account %s/%s", m.Namespace, m.Name)
}

func (m *ServiceAccountMutator) Empty() *core.ServiceAccount {
	return &core.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
	}
}

func (m *ServiceAccountMutator) Mutate(s *core.ServiceAccount) error {
	return m.meta.Mutate(s)
}
