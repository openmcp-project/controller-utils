# Connecting to Kubernetes Clusters

Both, the `pkg/clientconfig` and the `pkg/clusters` package, provide library functions that help with constructing clients for interacting with kubernetes clusters. The former one is more low-level, while the latter one tries to hide most of the boilerplate coding related to client creation.

## clientconfig

The `pkg/clientconfig` package provides helper functions for creating Kubernetes clients using multiple connection methods. It defines a `Config` struct that encapsulates a Kubernetes API target and supports various authentication methods like kubeconfig file and a Service Account.

### Noteworthy Functions

- `GetRESTConfig` generates a `*rest.Config` for interacting with the Kubernetes API. It supports using a kubeconfig string, a kubeconfig file path, a secret reference that contains a kubeconfig file or a Service Account.
- `GetClient` creates a client.Client for managing Kubernetes resources.

## clusters

The `pkg/clusters` package helps with loading kubeconfigs and creating clients for multiple clusters.
```go
foo := clusters.New("foo") // initializes a new cluster with id 'foo'
foo.RegisterConfigPathFlag(cmd.Flags()) // adds a '--foo-cluster' flag to the flag set for passing in a kubeconfig path
foo.InitializeRESTConfig() // loads the kubeconfig using the 'LoadKubeconfig' function from the 'controller' package
foo.InitializeClient(myScheme) // initializes the 'Client' and 'Cluster' interfaces from the controller-runtime
```
You can then use the different getter methods for working with the cluster.
