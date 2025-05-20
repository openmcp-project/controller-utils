package resources

import (
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterRoleBindingMutator struct {
	ClusterRoleBindingName string
	RoleRef                v1.RoleRef
	Subjects               []v1.Subject
	meta                   MetadataMutator
}

var _ Mutator[*v1.ClusterRoleBinding] = &ClusterRoleBindingMutator{}

func NewClusterRoleBindingMutator(clusterRoleBindingName string, subjects []v1.Subject, roleRef v1.RoleRef) Mutator[*v1.ClusterRoleBinding] {
	return &ClusterRoleBindingMutator{
		ClusterRoleBindingName: clusterRoleBindingName,
		RoleRef:                roleRef,
		Subjects:               subjects,
		meta:                   NewMetadataMutator(),
	}
}

func (m *ClusterRoleBindingMutator) String() string {
	return fmt.Sprintf("clusterrolebinding %s", m.ClusterRoleBindingName)
}

func (m *ClusterRoleBindingMutator) Empty() *v1.ClusterRoleBinding {
	return &v1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.ClusterRoleBindingName,
		},
	}
}

func (m *ClusterRoleBindingMutator) Mutate(r *v1.ClusterRoleBinding) error {
	r.RoleRef = m.RoleRef
	r.Subjects = m.Subjects
	return m.meta.Mutate(r)
}

func (m *ClusterRoleBindingMutator) MetadataMutator() MetadataMutator {
	return m.meta
}
