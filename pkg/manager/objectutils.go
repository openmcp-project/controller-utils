package externalsecrets

import (
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectID returns a human-readable identifier string for a Kubernetes object.
func ObjectID(obj client.Object) string {
	typeName := reflect.TypeOf(obj).Elem().Name()
	return fmt.Sprintf(`Kind=%s, Name=%s, Namespace=%s`, typeName, obj.GetName(), obj.GetNamespace())
}
