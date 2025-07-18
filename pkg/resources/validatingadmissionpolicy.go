package resources

import (
	"fmt"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ValidatingAdmissionPolicyMutator struct {
	Name string
	Spec admissionv1.ValidatingAdmissionPolicySpec
	meta MetadataMutator
}

var _ Mutator[*admissionv1.ValidatingAdmissionPolicy] = &ValidatingAdmissionPolicyMutator{}

func NewValidatingAdmissionPolicyMutator(name string, sourceSpec admissionv1.ValidatingAdmissionPolicySpec) Mutator[*admissionv1.ValidatingAdmissionPolicy] {
	return &ValidatingAdmissionPolicyMutator{
		Name: name,
		Spec: sourceSpec,
		meta: NewMetadataMutator(),
	}
}

func (m *ValidatingAdmissionPolicyMutator) String() string {
	return fmt.Sprintf("validatingadmissionpolicy %s", m.Name)
}

func (m *ValidatingAdmissionPolicyMutator) Empty() *admissionv1.ValidatingAdmissionPolicy {
	return &admissionv1.ValidatingAdmissionPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "ValidatingAdmissionPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name,
		},
	}
}

func (m *ValidatingAdmissionPolicyMutator) Mutate(r *admissionv1.ValidatingAdmissionPolicy) error {
	r.Spec = *m.Spec.DeepCopy()
	return m.meta.Mutate(r)
}

func (m *ValidatingAdmissionPolicyMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
