package resources

import (
	"fmt"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ValidatingAdmissionPolicyBindingMutator struct {
	Name string
	Spec admissionv1.ValidatingAdmissionPolicyBindingSpec
	meta MetadataMutator
}

var _ Mutator[*admissionv1.ValidatingAdmissionPolicyBinding] = &ValidatingAdmissionPolicyBindingMutator{}

func NewValidatingAdmissionPolicyBindingMutator(name string, sourceSpec admissionv1.ValidatingAdmissionPolicyBindingSpec) Mutator[*admissionv1.ValidatingAdmissionPolicyBinding] {
	return &ValidatingAdmissionPolicyBindingMutator{
		Name: name,
		Spec: sourceSpec,
		meta: NewMetadataMutator(),
	}
}

func (m *ValidatingAdmissionPolicyBindingMutator) String() string {
	return fmt.Sprintf("validatingadmissionpolicy %s", m.Name)
}

func (m *ValidatingAdmissionPolicyBindingMutator) Empty() *admissionv1.ValidatingAdmissionPolicyBinding {
	return &admissionv1.ValidatingAdmissionPolicyBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "ValidatingAdmissionPolicyBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name,
		},
	}
}

func (m *ValidatingAdmissionPolicyBindingMutator) Mutate(r *admissionv1.ValidatingAdmissionPolicyBinding) error {
	r.Spec = *m.Spec.DeepCopy()
	return m.meta.Mutate(r)
}

func (m *ValidatingAdmissionPolicyBindingMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
