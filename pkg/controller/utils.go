package controller

import (
	"crypto/sha1"
	"encoding/base32"
	"reflect"
	"strings"
)

const (
	maxLength                int = 63
	Base32EncodeStdLowerCase     = "abcdefghijklmnopqrstuvwxyz234567"
)

// K8sNameHash takes any number of string arguments and computes a hash out of it, which is then base32-encoded to be a valid k8s resource name.
// The arguments are joined with '/' before being hashed.
func K8sNameHash(ids ...string) string {
	name := strings.Join(ids, "/")
	h := sha1.New()
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
