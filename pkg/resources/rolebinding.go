package resources

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleBindingMutator struct {
	Name      string
	Namespace string
	Subjects  []v1.Subject
	RoleRef   v1.RoleRef
	meta      Mutator[client.Object]
}

var _ Mutator[*v1.RoleBinding] = &RoleBindingMutator{}

func NewRoleBindingMutator(name, namespace string, subjects []v1.Subject, roleRef v1.RoleRef, labels map[string]string, annotations map[string]string) Mutator[*v1.RoleBinding] {
	return &RoleBindingMutator{
		Name:      name,
		Namespace: namespace,
		Subjects:  subjects,
		RoleRef:   roleRef,
		meta:      NewMetadataMutator(labels, annotations),
	}
}

func (m *RoleBindingMutator) String() string {
	return fmt.Sprintf("rolebinding %s/%s", m.Namespace, m.Name)
}

func (m *RoleBindingMutator) Empty() *v1.RoleBinding {
	return &v1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
	}
}

func (m *RoleBindingMutator) Mutate(rb *v1.RoleBinding) error {
	rb.Subjects = m.Subjects
	rb.RoleRef = m.RoleRef
	return m.meta.Mutate(rb)
}
