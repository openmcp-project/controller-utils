package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/////////////////////////
/// BOILERPLATE STUFF ///
/////////////////////////

type metadataEntryType interface {
	// Name returns the name of the metadata type.
	// Should return either 'annotation' or 'label'.
	Name() string
	// GetData returns the metadata map.
	GetData(obj client.Object) map[string]string
	// SetData sets the metadata map.
	SetData(obj client.Object, data map[string]string)
}

type annotationMetadata struct{}

func (*annotationMetadata) Name() string {
	return "annotation"
}

func (*annotationMetadata) GetData(obj client.Object) map[string]string {
	return obj.GetAnnotations()
}

func (*annotationMetadata) SetData(obj client.Object, data map[string]string) {
	obj.SetAnnotations(data)
}

type labelMetadata struct{}

func (*labelMetadata) Name() string {
	return "label"
}

func (*labelMetadata) GetData(obj client.Object) map[string]string {
	return obj.GetLabels()
}

func (*labelMetadata) SetData(obj client.Object, data map[string]string) {
	obj.SetLabels(data)
}

var (
	ANNOTATION = &annotationMetadata{}
	LABEL      = &labelMetadata{}
)

////////////////////////////////////////
/// MODIFYING ANNOTATIONS AND LABELS ///
////////////////////////////////////////

type MetadataEntryAlreadyExistsError struct {
	MType        metadataEntryType
	Key          string
	DesiredValue string
	ActualValue  string
}

func NewMetadataEntryAlreadyExistsError(mType metadataEntryType, key, desired, actual string) *MetadataEntryAlreadyExistsError {
	return &MetadataEntryAlreadyExistsError{
		MType:        mType,
		Key:          key,
		DesiredValue: desired,
		ActualValue:  actual,
	}
}

func (e *MetadataEntryAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s '%s' already exists on the object and value '%s' could not be updated to '%s'", e.MType.Name(), e.Key, e.ActualValue, e.DesiredValue)
}

func IsMetadataEntryAlreadyExistsError(err error) bool {
	_, ok := err.(*MetadataEntryAlreadyExistsError)
	return ok
}

// EnsureAnnotation ensures that the given annotation has the desired state on the object.
// If the annotation already exists with the expected value or doesn't exist when deletion is desired, this is a no-op.
// If the annotation already exists with a different value, a MetadataEntryAlreadyExistsError is returned, unless mode OVERWRITE is set.
// To remove an annotation, set mode to DELETE. The given annValue does not matter in this case.
// If patch is set to true, the object will be patched in the cluster immediately, otherwise only the in-memory object is modified. client may be nil when patch is false.
func EnsureAnnotation(ctx context.Context, c client.Client, obj client.Object, annKey, annValue string, patch bool, mode ...ModifyMetadataEntryMode) error {
	return ensureMetadataEntry(ANNOTATION, ctx, c, obj, annKey, annValue, patch, mode...)
}

// EnsureLabel ensures that the given label has the desired state on the object.
// If the label already exists with the expected value or doesn't exist when deletion is desired, this is a no-op.
// If the label already exists with a different value, a MetadataEntryAlreadyExistsError is returned, unless mode OVERWRITE is set.
// To remove an label, set mode to DELETE. The given labelValue does not matter in this case.
// If patch is set to true, the object will be patched in the cluster immediately, otherwise only the in-memory object is modified. client may be nil when patch is false.
func EnsureLabel(ctx context.Context, c client.Client, obj client.Object, labelKey, labelValue string, patch bool, mode ...ModifyMetadataEntryMode) error {
	return ensureMetadataEntry(LABEL, ctx, c, obj, labelKey, labelValue, patch, mode...)
}

// ensureMetadataEntry is the common base method for EnsureAnnotation and EnsureLabel.
// This is mainly exposed for testing purposes, usually it is statically known whether an annotation or label is to be modified and the respective method should be used.
func ensureMetadataEntry(mType metadataEntryType, ctx context.Context, c client.Client, obj client.Object, key, value string, patch bool, mode ...ModifyMetadataEntryMode) error {
	modeDelete := false
	modeOverwrite := false
	for _, m := range mode {
		switch m {
		case DELETE:
			modeDelete = true
		case OVERWRITE:
			modeOverwrite = true
		}
	}
	quote := "\""
	data := mType.GetData(obj)
	if data == nil {
		data = map[string]string{}
	}
	val, ok := data[key]
	if (ok && val == value && !modeDelete) || (!ok && modeDelete) {
		// annotation/label already exists on the object, nothing to do
		return nil
	}
	if modeDelete {
		// delete annotation/label
		value = "null"
		quote = ""
		delete(data, key)
	} else {
		if ok && !modeOverwrite {
			return NewMetadataEntryAlreadyExistsError(mType, key, value, val)
		}
		// add annotation/label to obj
		data[key] = value
	}
	mType.SetData(obj, data)
	if patch {
		// patch annotation/label to in-cluster object
		if err := c.Patch(ctx, obj, client.RawPatch(types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"%ss":{"%s":%s%s%s}}}`, mType.Name(), key, quote, value, quote)))); err != nil {
			return err
		}
	}
	return nil
}

type ModifyMetadataEntryMode string

const (
	OVERWRITE ModifyMetadataEntryMode = "overwrite"
	DELETE    ModifyMetadataEntryMode = "delete"
)

//////////////////////////////////////
/// GETTING ANNOTATIONS AND LABELS ///
//////////////////////////////////////

// getMetadataEntry returns the value of the given annotation/label key on the object and whether it exists.
// This is mainly exposed for testing purposes, usually it is statically known whether an annotation or label is to be fetched and the respective method should be used.
func getMetadataEntry(mType metadataEntryType, obj client.Object, key string) (string, bool) {
	data := mType.GetData(obj)
	if data == nil {
		return "", false
	}
	val, ok := data[key]
	return val, ok
}

// HasAnnotation returns true if the given annotation key exists on the object.
func HasAnnotation(obj client.Object, key string) bool {
	_, ok := getMetadataEntry(ANNOTATION, obj, key)
	return ok
}

// HasLabel returns true if the given label key exists on the object.
func HasLabel(obj client.Object, key string) bool {
	_, ok := getMetadataEntry(LABEL, obj, key)
	return ok
}

// HasAnnotationWithValue returns true if the given annotation key exists on the object and has the given value.
func HasAnnotationWithValue(obj client.Object, key, value string) bool {
	val, ok := getMetadataEntry(ANNOTATION, obj, key)
	return ok && val == value
}

// HasLabelWithValue returns true if the given label key exists on the object and has the given value.
func HasLabelWithValue(obj client.Object, key, value string) bool {
	val, ok := getMetadataEntry(LABEL, obj, key)
	return ok && val == value
}

// GetAnnotation returns the value of the given annotation key on the object and whether it exists.
func GetAnnotation(obj client.Object, key string) (string, bool) {
	return getMetadataEntry(ANNOTATION, obj, key)
}

// GetLabel returns the value of the given label key on the object and whether it exists.
func GetLabel(obj client.Object, key string) (string, bool) {
	return getMetadataEntry(LABEL, obj, key)
}
