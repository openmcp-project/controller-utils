package testing

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	yaml2 "sigs.k8s.io/yaml"
)

// ShouldReconcile calls the given reconciler with the given request and expects no error.
func ShouldReconcile(ctx context.Context, reconciler reconcile.Reconciler, req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	res, err := reconciler.Reconcile(ctx, req)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred(), optionalDescription...)
	return res
}

// ShouldNotReconcile calls the given reconciler with the given request and expects an error.
func ShouldNotReconcile(ctx context.Context, reconciler reconcile.Reconciler, req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	res, err := reconciler.Reconcile(ctx, req)
	gomega.ExpectWithOffset(1, err).To(gomega.HaveOccurred(), optionalDescription...)
	return res
}

// ExpectRequeue expects the given result to indicate a requeue.
func ExpectRequeue(res reconcile.Result) {
	requeue := res.Requeue || res.RequeueAfter > 0
	gomega.ExpectWithOffset(1, requeue).To(gomega.BeTrue())
}

// ExpectNoRequeue expects the given result to indicate no requeue.
func ExpectNoRequeue(res reconcile.Result) {
	requeue := res.Requeue || res.RequeueAfter > 0
	gomega.ExpectWithOffset(1, requeue).To(gomega.BeFalse())
}

// RequestFromObject creates a reconcile.Request from the given object.
func RequestFromObject(obj client.Object) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		},
	}
}

// RequestFromStrings creates a reconcile.Request using the specified name and namespace.
// The first argument is the name of the object.
// An optional second argument contains the namespace. All further arguments are ignored.
func RequestFromStrings(name string, maybeNamespace ...string) reconcile.Request {
	namespace := ""
	if len(maybeNamespace) > 0 {
		namespace = maybeNamespace[0]
	}
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}
}

// LoadObject reads a file and unmarshals it into the given object.
// obj must be a non-nil pointer.
func LoadObject(obj any, paths ...string) error {
	objRaw, err := os.ReadFile(path.Join(paths...))
	if err != nil {
		return err
	}
	return yaml2.Unmarshal(objRaw, obj) // for some reason doesn't work with the other yaml library -.-
}

// LoadObjects loads all Kubernetes manifests from the given path and returns them as `client.Object`.
func LoadObjects(p string, scheme *runtime.Scheme) ([]client.Object, error) {
	var objectList []client.Object
	objDecoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	err := filepath.Walk(p, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !info.IsDir() && filepath.Ext(path) == ".yaml" {
			reader, err := os.Open(path)
			if err != nil {
				return err
			}

			yamlDecoder := yaml.NewDecoder(reader)
			if yamlDecoder == nil {
				return errors.New("failed to create YAML decoder")
			}

			for {
				var raw map[string]interface{}
				err = yamlDecoder.Decode(&raw)

				if raw == nil {
					break
				}

				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					} else {
						return err
					}
				}

				var data []byte
				data, err = yaml.Marshal(raw)
				if err != nil {
					return err
				}

				into := &unstructured.Unstructured{}
				obj, _, err := objDecoder.Decode(data, nil, into)
				if err != nil {
					return err
				}

				objectList = append(objectList, obj.(client.Object))
			}
		}

		return nil
	})
	return objectList, err
}
