package jsonpatch

import "encoding/json"

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
	Value *Any `json:"value,omitempty"`

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
func NewJSONPatch(op Operation, path string, value *Any, from *string) JSONPatch {
	return JSONPatch{
		Operation: op,
		Path:      path,
		Value:     value,
		From:      from,
	}
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

func NewAny(value any) *Any {
	return &Any{Value: value}
}

type Any struct {
	Value any `json:"-"`
}

var _ json.Marshaler = &Any{}
var _ json.Unmarshaler = &Any{}

func (a *Any) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	return json.Marshal(a.Value)
}

func (a *Any) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		a.Value = nil
		return nil
	}
	return json.Unmarshal(data, &a.Value)
}

func (in *JSONPatch) DeepCopy() *JSONPatch {
	if in == nil {
		return nil
	}
	out := &JSONPatch{}
	in.DeepCopyInto(out)
	return out
}

func (in *JSONPatch) DeepCopyInto(out *JSONPatch) {
	if out == nil {
		return
	}
	out.Operation = in.Operation
	out.Path = in.Path
	if in.Value != nil {
		valueCopy := *in.Value
		out.Value = &valueCopy
	} else {
		out.Value = nil
	}
	if in.From != nil {
		fromCopy := *in.From
		out.From = &fromCopy
	} else {
		out.From = nil
	}
}

func (in *JSONPatches) DeepCopy() *JSONPatches {
	if in == nil {
		return nil
	}
	out := &JSONPatches{}
	for _, item := range *in {
		outItem := item.DeepCopy()
		*out = append(*out, *outItem)
	}
	return out
}

func (in *JSONPatches) DeepCopyInto(out *JSONPatches) {
	if out == nil {
		return
	}
	*out = make(JSONPatches, len(*in))
	for i, item := range *in {
		outItem := item.DeepCopy()
		(*out)[i] = *outItem
	}
}

func (in *Any) DeepCopy() *Any {
	if in == nil {
		return nil
	}
	out := &Any{}
	in.DeepCopyInto(out)
	return out
}

func (in *Any) DeepCopyInto(out *Any) {
	if out == nil {
		return
	}
	if in.Value == nil {
		out.Value = nil
		return
	}
	// Use json.Marshal and json.Unmarshal to deep copy the value.
	data, err := json.Marshal(in.Value)
	if err != nil {
		panic("failed to marshal Any value: " + err.Error())
	}
	if err := json.Unmarshal(data, &out.Value); err != nil {
		panic("failed to unmarshal Any value: " + err.Error())
	}
}

// Ptr is a convenience function to create a pointer to the given value.
func Ptr[T any](val T) *T {
	return &val
}
