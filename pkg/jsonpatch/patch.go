package jsonpatch

import (
	"encoding/json"
	"fmt"
	"reflect"

	jplib "github.com/evanphx/json-patch/v5"

	jpapi "github.com/openmcp-project/controller-utils/api/jsonpatch"
)

type Untyped = []byte

type JSONPatch = TypedJSONPatch[Untyped]

type TypedJSONPatch[T any] struct {
	jpapi.JSONPatches
}

type JSONPatchOptions struct {
	*jplib.ApplyOptions

	// Indent is the string used for indentation in the output JSON.
	// Empty string means no indentation.
	Indent string
}

type JSONPatchOption func(*JSONPatchOptions)

// New creates a new JSONPatch with the given patches.
// This JSONPatch's Apply method works on plain JSON bytes.
// To apply the patches to an arbitrary type (which is marshalled to JSON before and unmarshalled back afterwards),
// use NewTyped instead.
func New(patches jpapi.JSONPatches) *JSONPatch {
	return &TypedJSONPatch[Untyped]{
		JSONPatches: patches,
	}
}

// NewTyped creates a new TypedJSONPatch with the given patches.
func NewTyped[T any](patches jpapi.JSONPatches) *TypedJSONPatch[T] {
	return &TypedJSONPatch[T]{
		JSONPatches: patches,
	}
}

// Apply applies the patch to the given document.
// If the generic type is Untyped (which is an alias for []byte),
// it will treat the document as raw JSON bytes.
// Otherwise, doc is marshalled to JSON before applying the patch and then again unmarshalled back to the original type afterwards.
func (p *TypedJSONPatch[T]) Apply(doc T, options ...JSONPatchOption) (T, error) {
	var result T
	var rawDoc []byte
	isUntyped := reflect.TypeFor[T]() == reflect.TypeFor[Untyped]()
	if isUntyped {
		rawDoc = any(doc).(Untyped)
	} else {
		tmp, err := json.Marshal(doc)
		if err != nil {
			return result, fmt.Errorf("failed to marshal document: %w", err)
		}
		rawDoc = tmp
	}

	opts := &JSONPatchOptions{
		ApplyOptions: jplib.NewApplyOptions(),
	}
	for _, opt := range options {
		opt(opts)
	}

	rawPatch, err := json.Marshal(p)
	if err != nil {
		return result, fmt.Errorf("failed to marshal JSONPatch: %w", err)
	}
	patch, err := jplib.DecodePatch(rawPatch)
	if err != nil {
		return result, fmt.Errorf("failed to decode JSONPatch: %w", err)
	}

	if opts.Indent != "" {
		rawDoc, err = patch.ApplyIndentWithOptions(rawDoc, opts.Indent, opts.ApplyOptions)
	} else {
		rawDoc, err = patch.ApplyWithOptions(rawDoc, opts.ApplyOptions)
	}
	if err != nil {
		return result, fmt.Errorf("failed to apply JSONPatch: %w", err)
	}

	if isUntyped {
		return any(rawDoc).(T), nil
	}
	if err := json.Unmarshal(rawDoc, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal result into type %T: %w", result, err)
	}
	return result, nil
}

// SupportNegativeIndices decides whether to support non-standard practice of
// allowing negative indices to mean indices starting at the end of an array.
// Default to true.
func SupportNegativeIndices(val bool) JSONPatchOption {
	return func(opts *JSONPatchOptions) {
		opts.SupportNegativeIndices = val
	}
}

// AccumulatedCopySizeLimit limits the total size increase in bytes caused by
// "copy" operations in a patch.
func AccumulatedCopySizeLimit(val int64) JSONPatchOption {
	return func(opts *JSONPatchOptions) {
		opts.AccumulatedCopySizeLimit = val
	}
}

// AllowMissingPathOnRemove indicates whether to fail "remove" operations when the target path is missing.
// Default to false.
func AllowMissingPathOnRemove(val bool) JSONPatchOption {
	return func(opts *JSONPatchOptions) {
		opts.AllowMissingPathOnRemove = val
	}
}

// EnsurePathExistsOnAdd instructs json-patch to recursively create the missing parts of path on "add" operation.
// Defaults to false.
func EnsurePathExistsOnAdd(val bool) JSONPatchOption {
	return func(opts *JSONPatchOptions) {
		opts.EnsurePathExistsOnAdd = val
	}
}

// EscapeHTML sets the EscapeHTML flag for json marshalling.
// Defaults to true.
func EscapeHTML(val bool) JSONPatchOption {
	return func(opts *JSONPatchOptions) {
		opts.EscapeHTML = val
	}
}

// Indent sets the indentation string for the output JSON.
// If empty, no indentation is applied.
func Indent(val string) JSONPatchOption {
	return func(opts *JSONPatchOptions) {
		opts.Indent = val
	}
}

var _ json.Marshaler = &TypedJSONPatch[Untyped]{}

// MarshalJSON marshals the TypedJSONPatch to JSON.
// Note that this uses the ConvertPath function to ensure that the paths are in the correct format.
func (p *TypedJSONPatch[T]) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte("null"), nil
	}

	// copy the single patches to convert their paths without modifying the original
	patches := make(jpapi.JSONPatches, len(p.JSONPatches))
	for i, jsonPatch := range p.JSONPatches {
		p := jsonPatch.DeepCopy()
		convertedPath, iperr := ConvertPath(p.Path)
		if iperr != nil {
			return nil, fmt.Errorf("failed to convert path at index %d: %w", i, iperr)
		}
		p.Path = convertedPath

		if p.From != nil {
			convertedFrom, iperr := ConvertPath(*p.From)
			if iperr != nil {
				return nil, fmt.Errorf("failed to convert 'from' path at index %d: %w", i, iperr)
			}
			p.From = &convertedFrom
		}

		patches[i] = *p
	}

	return json.Marshal(patches)
}
