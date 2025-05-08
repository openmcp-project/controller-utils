# Kubernetes Resource Management

The `pkg/resource` package contains some useful functions for working with Kubernetes resources. The `Mutator` interface can be used to modify resources in a generic way. It is used by the `Mutate` function, which takes a resource and a mutator and applies the mutator to the resource.
The package also contains convenience types for the most common resource types, e.g. `ConfigMap`, `Secret`, `ClusterRole`, `ClusterRoleBinding`, etc. These types implement the `Mutator` interface and can be used to modify the corresponding resources.

### Examples

Create or update a `ConfigMap`, a `ServiceAccount` and a `Deployment` using the `Mutator` interface:

```go
type myDeploymentMutator struct {
}

var _ resource.Mutator[*appsv1.Deployment] = &myDeploymentMutator{}

func newDeploymentMutator() resources.Mutator[*appsv1.Deployment] {
	return &MyDeploymentMutator{}
}

func (m *MyDeploymentMutator) String() string {
	return "deployment default/test"
}

func (m *MyDeploymentMutator) Empty() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{	
			Name:      "test",
			Namespace: "default",
		},
	}
}

func (m *MyDeploymentMutator) Mutate(deployment *appsv1.Deployment) error {
	// create one container with an image
	deployment.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "test",
			Image: "test-image:latest",
		},
	}
	return nil
}


func ReconcileResources(ctx context.Context, client client.Client) error {
	configMapResource := resource.NewConfigMap("my-configmap", "my-namespace", map[string]string{
		"label1": "value1",
		"label2": "value2",
	}, nil)

	serviceAccountResource := resource.NewServiceAccount("my-serviceaccount", "my-namespace", nil, nil)
	
	myDeploymentMutator := newDeploymentMutator()
	
	var err error
	
	err = resources.CreateOrUpdateResource(ctx, client, configMapResource)
	if err != nil {
		return err
	}
	
	resources.CreateOrUpdateResource(ctx, client, serviceAccountResource)
	if err != nil {
		return err
	}
	
	err = resources.CreateOrUpdateResource(ctx, client, myDeploymentMutator)
	if err != nil {
		return err
	}
	
	return nil
}
```
