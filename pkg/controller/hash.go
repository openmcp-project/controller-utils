package controller

import (
	"crypto/sha3"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrInvalidNames   = errors.New("list of names must not be empty and contain at least one non-empty string")
	ErrMaxLenTooSmall = errors.New("maxLen must be greater than 10")
)

// version8UUID creates a new UUID (version 8) from a byte slice. Returns an error if the slice does not have a length of 16. The bytes are copied from the slice.
// The bits 48-51 and 64-65 are modified to make the output recognizable as a version 8 UUID, so only 122 out of 128 bits from the input data will be kept.
func version8UUID(data []byte) (uuid.UUID, error) {
	if len(data) != 16 {
		return uuid.Nil, fmt.Errorf("invalid data (got %d bytes)", len(data))
	}

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Set 4-bit version field (ver = 0b1000)
	dataCopy[6] &= 0b00001111
	dataCopy[6] |= 0b10000000

	// Set 2-bit variant field (var = 0b10)
	dataCopy[8] &= 0b00111111
	dataCopy[8] |= 0b10000000

	return uuid.FromBytes(dataCopy)
}

// K8sNameUUID takes any number of string arguments and computes a hash out of it, which is then formatted as a version 8 UUID.
// The arguments are joined with '/' before being hashed.
// Returns an error if the list of ids is empty or contains only empty strings.
// Deprecated: Use NameHashSHAKE128Base32 instead.
func K8sNameUUID(names ...string) (string, error) {
	if err := validateIDs(names); err != nil {
		return "", err
	}

	name := strings.Join(names, "/")
	hash := sha3.SumSHAKE128([]byte(name), 16)
	u, err := version8UUID(hash)

	return u.String(), err
}

func validateIDs(names []string) error {
	for _, name := range names {
		// at least one non-empty string found
		if name != "" {
			return nil
		}
	}
	return ErrInvalidNames
}

// K8sNameUUIDUnsafe works like K8sNameUUID, but panics instead of returning an error.
// This should only be used in places where the input is guaranteed to be valid.
// Deprecated: Use NameHashSHAKE128Base32 instead.
func K8sNameUUIDUnsafe(names ...string) string {
	uuid, err := K8sNameUUID(names...)
	if err != nil {
		panic(err)
	}
	return uuid
}

// K8sObjectUUID takes a client object and computes a hash out of the namespace and name, which is then formatted as a version 8 UUID.
// An empty namespace will be replaced by "default".
// Deprecated: Use ObjectHashSHAKE128Base32 instead.
func K8sObjectUUID(obj client.Object) (string, error) {
	name, namespace := obj.GetName(), obj.GetNamespace()
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}
	return K8sNameUUID(namespace, name)
}

// K8sObjectUUIDUnsafe works like K8sObjectUUID, but panics instead of returning an error.
// This should only be used in places where the input is guaranteed to be valid.
// Deprecated: Use ObjectHashSHAKE128Base32 instead.
func K8sObjectUUIDUnsafe(obj client.Object) string {
	uuid, err := K8sObjectUUID(obj)
	if err != nil {
		panic(err)
	}
	return uuid
}

// NameHashSHAKE128Base32 takes any number of string arguments and computes a hash out of it. The output string will be 8 characters long.
// The arguments are joined with '/' before being hashed.
func NameHashSHAKE128Base32(names ...string) string {
	name := strings.Join(names, "/")

	// Desired output length = 8 chars
	// 8 chars * 5 bits (base32) / 8 bits per byte = 5 bytes
	hash := sha3.SumSHAKE128([]byte(name), 5)

	return base32.NewEncoding(Base32EncodeStdLowerCase).WithPadding(base32.NoPadding).EncodeToString(hash)
}

// ObjectHashSHAKE128Base32 takes a client object and computes a hash out of the namespace and name. The output string will be 8 characters long.
// An empty namespace will be replaced by "default".
func ObjectHashSHAKE128Base32(obj client.Object) string {
	name, namespace := obj.GetName(), obj.GetNamespace()
	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}
	return NameHashSHAKE128Base32(namespace, name)
}

// ShortenToXCharacters shortens the input string if it exceeds maxLen.
// It uses NameHashSHAKE128Base32 to generate a hash which will replace the last few characters.
func ShortenToXCharacters(input string, maxLen int) (string, error) {
	if len(input) <= maxLen {
		return input, nil
	}

	hash := NameHashSHAKE128Base32(input)
	suffix := fmt.Sprintf("--%s", hash)
	trimLength := maxLen - len(suffix)

	if trimLength <= 0 {
		return "", ErrMaxLenTooSmall
	}

	return input[:trimLength] + suffix, nil
}
