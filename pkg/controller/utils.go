package controller

import (
	"crypto/sha256"
	"encoding/base32"
	"reflect"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Base32EncodeStdLowerCase = "abcdefghijklmnopqrstuvwxyz234567"
)

// K8sNameHash takes any number of string arguments and computes a hash out of it, which is then base32-encoded to be a valid DNS1123Subdomain k8s resource name
// The arguments are joined with '/' before being hashed.
func K8sNameHash(ids ...string) string {
	name := strings.Join(ids, "/")
	// since we are not worried about length-extension attacks (in fact we are not even using hashing for
	// any security purposes), use sha2 for better performance compared to sha3
	h := sha256.New()
	_, _ = h.Write([]byte(name))
	// we need base32 encoding as some base64 (even url safe base64) characters are not supported by k8s
	// see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	return base32.NewEncoding(Base32EncodeStdLowerCase).WithPadding(base32.NoPadding).EncodeToString(h.Sum(nil))
}

// IsNil checks if a given pointer is nil.
// Opposed to 'i == nil', this works for typed and untyped nil values.
func IsNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// ObjectKey returns a client.ObjectKey for the given name and optionally namespace.
// The first argument is the name of the object.
// An optional second argument contains the namespace. All further arguments are ignored.
func ObjectKey(name string, maybeNamespace ...string) client.ObjectKey {
	namespace := ""
	if len(maybeNamespace) > 0 {
		namespace = maybeNamespace[0]
	}
	return client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
}

// RemoveFinalizersWithPrefix removes finalizers with a given prefix from the object and returns their suffixes.
// If the third argument is true, all finalizers with the given prefix are removed, otherwise only the first one.
// The bool return value indicates whether a finalizer was removed.
// If it is true, the slice return value holds the suffixes of all removed finalizers (will be of length 1 if removeAll is false).
// If it is false, no finalizer with the given prefix was found. The slice return value will be empty in this case.
// The logic is based on the controller-runtime's RemoveFinalizer function.
func RemoveFinalizersWithPrefix(obj client.Object, prefix string, removeAll bool) ([]string, bool) {
	fins := obj.GetFinalizers()
	length := len(fins)
	suffixes := make([]string, 0, length)
	found := false

	index := 0
	for i := range length {
		if (removeAll || !found) && strings.HasPrefix(fins[i], prefix) {
			suffixes = append(suffixes, strings.TrimPrefix(fins[i], prefix))
			found = true
			continue
		}
		fins[index] = fins[i]
		index++
	}
	obj.SetFinalizers(fins[:index])
	return suffixes, length != index
}
