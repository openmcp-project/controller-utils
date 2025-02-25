package testing

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// GetFakeClient returns a fake Kubernetes client with the given objects initialized.
func GetFakeClient(scheme *runtime.Scheme, initObjects ...client.Object) (client.WithWatch, error) {
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjects...).WithStatusSubresource(initObjects...).Build(), nil
}

// GetFakeClientWithDynamicObjects returns a fake Kubernetes client with the given objects initialized.
// Dynamic objects are objects that are not known at compile time, but are create while running the controller.
// The dynamic objects are used to initialize the status subresource.
func GetFakeClientWithDynamicObjects(scheme *runtime.Scheme, dynamicObjects []client.Object, initObjects ...client.Object) (client.WithWatch, error) {
	statusObjects := make([]client.Object, 0, len(initObjects)+len(dynamicObjects))
	statusObjects = append(statusObjects, dynamicObjects...)
	statusObjects = append(statusObjects, initObjects...)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjects...).WithStatusSubresource(statusObjects...).Build(), nil
}
