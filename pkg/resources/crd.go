package resources

import (
	"fmt"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type crdMutator struct {
	crd *apiextv1.CustomResourceDefinition
}

var _ Mutator[*apiextv1.CustomResourceDefinition] = &crdMutator{}

func NewCRDMutator(crd *apiextv1.CustomResourceDefinition) Mutator[*apiextv1.CustomResourceDefinition] {
	return &crdMutator{crd: crd}
}

func (m *crdMutator) String() string {
	return fmt.Sprintf("crd %s", m.crd.Name)
}

func (m *crdMutator) Empty() *apiextv1.CustomResourceDefinition {
	return &apiextv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.crd.Name,
		},
	}
}

func (m *crdMutator) Mutate(r *apiextv1.CustomResourceDefinition) error {
	m.crd.Spec.DeepCopyInto(&r.Spec)
	return nil
}
