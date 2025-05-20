package resources

import (
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleMutator struct {
	Name      string
	Namespace string
	Rules     []v1.PolicyRule
	meta      MetadataMutator
}

var _ Mutator[*v1.Role] = &RoleMutator{}

func NewRoleMutator(name, namespace string, rules []v1.PolicyRule) Mutator[*v1.Role] {
	return &RoleMutator{
		Name:      name,
		Namespace: namespace,
		Rules:     rules,
		meta:      NewMetadataMutator(),
	}
}

func (m *RoleMutator) String() string {
	return fmt.Sprintf("role %s/%s", m.Namespace, m.Name)
}

func (m *RoleMutator) Empty() *v1.Role {
	return &v1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
	}
}

func (m *RoleMutator) Mutate(r *v1.Role) error {
	r.Rules = m.Rules
	return m.meta.Mutate(r)
}

func (m *RoleMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
