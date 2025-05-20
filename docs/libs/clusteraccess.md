# Generating Kubeconfigs for k8s Clusters

The `pkg/clusteraccess` package contains useful helper functions to create a kubeconfig for a k8s cluster. This includes functions to create ServiceAccounts as well as (Cluster)Roles and (Cluster)RoleBindings, but also generating a ServiceAccount token and building a kubeconfig from this token.