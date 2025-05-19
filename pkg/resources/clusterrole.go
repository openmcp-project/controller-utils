package resources

import (
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterRoleMutator struct {
	Name  string
	Rules []v1.PolicyRule
	meta  MetadataMutator
}

var _ Mutator[*v1.ClusterRole] = &ClusterRoleMutator{}

func NewClusterRoleMutator(name string, rules []v1.PolicyRule, labels map[string]string, annotations map[string]string) Mutator[*v1.ClusterRole] {
	return &ClusterRoleMutator{
		Name:  name,
		Rules: rules,
		meta:  NewMetadataMutator(labels, annotations),
	}
}

func (m *ClusterRoleMutator) String() string {
	return fmt.Sprintf("clusterrole %s", m.Name)
}

func (m *ClusterRoleMutator) Empty() *v1.ClusterRole {
	return &v1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name,
		},
	}
}

func (m *ClusterRoleMutator) Mutate(r *v1.ClusterRole) error {
	r.Rules = m.Rules
	return m.meta.Mutate(r)
}

func (m *ClusterRoleMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
