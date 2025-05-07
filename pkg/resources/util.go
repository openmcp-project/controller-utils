package resources

import v1 "k8s.io/api/rbac/v1"

// NewClusterRoleRef creates a RoleRef for a ClusterRole.
func NewClusterRoleRef(name string) v1.RoleRef {
	return v1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     name,
	}
}

// NewRoleRef creates a RoleRef for a Role.
func NewRoleRef(name string) v1.RoleRef {
	return v1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     name,
	}
}
