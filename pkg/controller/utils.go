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
// Deprecated: Use NameHashSHAKE128Base32 instead.
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
