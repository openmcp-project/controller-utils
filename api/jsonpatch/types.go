// +kubebuilder:object:generate=true
package jsonpatch

import (
	"encoding/json"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// JSONPatch represents a JSON patch operation.
// Technically, a single JSON patch (as defined by RFC 6902) is a list of patch operations.
// Opposed to that, this type represents a single operation. Use the JSONPatches type for a list of operations instead.
// +kubebuilder:validation:Schemaless
// +kubebuilder:validation:Type=object
type JSONPatch struct {
	// Operation is the operation to perform.
	// Valid values are: add, remove, replace, move, copy, test
	// +kubebuilder:validation:Enum=add;remove;replace;move;copy;test
	// +kubebuilder:validation:Required
	Operation Operation `json:"op"`

	// Path is the path to the target location in the JSON document.
	// +kubebuilder:validation:Required
	Path string `json:"path"`

	// Value is the value to set at the target location.
	// Required for add, replace, and test operations.
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Value *apiextensionsv1.JSON `json:"value,omitempty"`

	// From is the source location for move and copy operations.
	// +optional
	From *string `json:"from,omitempty"`
}

// JSONPatches is a list of JSON patch operations.
// This is technically a 'JSON patch' as defined in RFC 6902.
type JSONPatches []JSONPatch

type Operation string

const (
	// ADD is the constant for the JSONPatch 'add' operation.
	ADD Operation = "add"
	// REMOVE is the constant for the JSONPatch 'remove' operation.
	REMOVE Operation = "remove"
	// REPLACE is the constant for the JSONPatch 'replace' operation.
	REPLACE Operation = "replace"
	// MOVE is the constant for the JSONPatch 'move' operation.
	MOVE Operation = "move"
	// COPY is the constant for the JSONPatch 'copy' operation.
	COPY Operation = "copy"
	// TEST is the constant for the JSONPatch 'test' operation.
	TEST Operation = "test"
)

// NewJSONPatch creates a new JSONPatch with the given values.
// If value is non-nil, it is marshaled to JSON. Returns an error if the value cannot be marshaled.
func NewJSONPatch(op Operation, path string, value any, from *string) (JSONPatch, error) {
	res := JSONPatch{
		Operation: op,
		Path:      path,
		From:      from,
	}
	var err error
	if value != nil {
		var valueJSON []byte
		valueJSON, err = json.Marshal(value)
		res.Value = &apiextensionsv1.JSON{
			Raw: valueJSON,
		}
	}
	return res, err
}

// NewJSONPatchOrPanic works like NewJSONPatch, but instead of returning an error, it panics if the patch cannot be created.
func NewJSONPatchOrPanic(op Operation, path string, value any, from *string) JSONPatch {
	patch, err := NewJSONPatch(op, path, value, from)
	if err != nil {
		panic(err)
	}
	return patch
}

// NewJSONPatches combines multiple JSONPatch instances into a single JSONPatches instance.
// This is a convenience function to create a JSONPatches instance from multiple JSONPatch instances.
func NewJSONPatches(patches ...JSONPatch) JSONPatches {
	result := make(JSONPatches, 0, len(patches))
	for _, patch := range patches {
		result = append(result, patch)
	}
	return result
}
