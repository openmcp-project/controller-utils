// +kubebuilder:object:generate=true
package jsonpatch

import (
	"github.com/fluxcd/pkg/apis/kustomize"
)

// JSONPatch represents a JSON patch operation.
// Technically, a single JSON patch (as defined by RFC 6902) is a list of patch operations.
// Opposed to that, this type represents a single operation. Use the JSONPatches type for a list of operations instead.
type JSONPatch = kustomize.JSON6902

// JSONPatches is a list of JSON patch operations.
// This is technically a 'JSON patch' as defined in RFC 6902.
type JSONPatches []JSONPatch

const (
	// ADD is the constant for the JSONPatch 'add' operation.
	ADD = "add"
	// REMOVE is the constant for the JSONPatch 'remove' operation.
	REMOVE = "remove"
	// REPLACE is the constant for the JSONPatch 'replace' operation.
	REPLACE = "replace"
	// MOVE is the constant for the JSONPatch 'move' operation.
	MOVE = "move"
	// COPY is the constant for the JSONPatch 'copy' operation.
	COPY = "copy"
	// TEST is the constant for the JSONPatch 'test' operation.
	TEST = "test"
)
