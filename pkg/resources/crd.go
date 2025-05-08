package resources

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type crdMutator struct {
	crd  *apiextv1.CustomResourceDefinition
	meta Mutator[client.Object]
}

var _ Mutator[*apiextv1.CustomResourceDefinition] = &crdMutator{}

func NewCRDMutator(crd *apiextv1.CustomResourceDefinition, labels map[string]string, annotations map[string]string) Mutator[*apiextv1.CustomResourceDefinition] {
	return &crdMutator{crd: crd, meta: NewMetadataMutator(labels, annotations)}
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
	return m.meta.Mutate(r)
}
