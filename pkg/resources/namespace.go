package resources

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespaceMutator struct {
	name string
	meta MetadataMutator
}

var _ Mutator[*v1.Namespace] = &namespaceMutator{}

func NewNamespaceMutator(name string) Mutator[*v1.Namespace] {
	return &namespaceMutator{name: name, meta: NewMetadataMutator()}
}

func (m *namespaceMutator) String() string {
	return fmt.Sprintf("namespace %s", m.name)
}

func (m *namespaceMutator) Empty() *v1.Namespace {
	return &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.name,
		},
	}
}

func (m *namespaceMutator) Mutate(r *v1.Namespace) error {
	return m.meta.Mutate(r)
}

func (m *namespaceMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
