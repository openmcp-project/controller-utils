package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
	"github.com/openmcp-project/controller-utils/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewStatusUpdaterBuilder initializes a new StatusUpdaterBuilder.
func NewStatusUpdaterBuilder[Obj client.Object, PhType ~string, ConType comparable]() *StatusUpdaterBuilder[Obj, PhType, ConType] {
	return &StatusUpdaterBuilder[Obj, PhType, ConType]{
		internal: newStatusUpdater[Obj, PhType, ConType](),
	}
}

// StatusUpdaterBuilder is a builder for creating a status updater.
// Do not use this directly, use NewStatusUpdaterBuilder() instead.
type StatusUpdaterBuilder[Obj client.Object, PhType ~string, ConType comparable] struct {
	internal *statusUpdater[Obj, PhType, ConType]
}

// WithFieldOverride overwrites the name of the field.
// Use STATUS_FIELD to override the name of the field that holds the status itself.
// All other fields are expected to be within the status struct.
// Set to an empty string to disable the field. Doing this for STATUS_FIELD disables the complete status update.
// The default names are:
// - STATUS_FIELD: "Status"
// - STATUS_FIELD_OBSERVED_GENERATION: "ObservedGeneration"
// - STATUS_FIELD_LAST_RECONCILE_TIME: "LastReconcileTime"
// - STATUS_FIELD_CONDITIONS: "Conditions"
// - STATUS_FIELD_REASON: "Reason"
// - STATUS_FIELD_MESSAGE: "Message"
// - STATUS_FIELD_PHASE: "Phase"
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithFieldOverride(field StatusField, name string) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	if name == "" {
		delete(b.internal.fieldNames, field)
	} else {
		b.internal.fieldNames[field] = name
	}
	return b
}

// WithFieldOverrides is a wrapper around WithFieldOverride that allows to apply multiple overrides at once.
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithFieldOverrides(overrides map[StatusField]string) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	for field, name := range overrides {
		b.WithFieldOverride(field, name)
	}
	return b
}

// WithNestedStruct is a helper for easily updating the field names if some or all of the fields are wrapped in a nested struct within the status.
// Basically, the field names for all specified fields are prefixed with '<name>.', unless the field is empty (which disables the field).
// If appliesTo is empty, all fields are assumed to be nested (except for the status itself).
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithNestedStruct(name string, appliesTo ...StatusField) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	if len(appliesTo) == 0 {
		appliesTo = AllStatusFields()
	}
	for _, field := range appliesTo {
		oldName := b.internal.fieldNames[field]
		if oldName == "" {
			continue
		}
		b.WithFieldOverride(field, fmt.Sprintf("%s.%s", name, oldName))
	}
	return b
}

// WithoutField removes the field from the status update.
// This is an alias for WithFieldOverride(field, "").
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithoutField(field StatusField) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	return b.WithFieldOverride(field, "")
}

// WithConditionUpdater must be called if the conditions should be updated, because this requires some additional information.
// Note that the conditions will only be updated if this method has been called (with a non-nil condition constructor) AND the conditions field has not been disabled.
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithConditionUpdater(construct func() conditions.Condition[ConType], removeUntouchedConditions bool) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	b.internal.conConstruct = construct
	b.internal.removeUntouchedConditions = removeUntouchedConditions
	return b
}

// WithPhaseUpdateFunc sets the function that determines the phase of the object.
// It is strongly recommended to either disable the phase field or override this function, because the default will simply set the Phase to the zero value of PhType.
// The function is called with a deep copy of the object, after all other status updates have been applied (except for the custom update).
// If the returned error is nil, the object's phase is then set to the returned value.
// Setting this to nil causes the default phase update function to be used. To disable the phase update altogether, use WithoutField(STATUS_FIELD_PHASE).
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithPhaseUpdateFunc(f func(obj Obj, rr ReconcileResult[Obj, ConType]) (PhType, error)) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	b.internal.phaseUpdateFunc = f
	return b
}

// WithCustomUpdateFunc allows to pass in a function that performs a custom update on the object.
// This function is called after all other status updates have been applied.
// It gets the original object passed in and can modify it directly.
// Note that only changes to the status field are sent to the cluster.
// Set this to nil to disable the custom update (it is nil by default).
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) WithCustomUpdateFunc(f func(obj Obj, rr ReconcileResult[Obj, ConType]) error) *StatusUpdaterBuilder[Obj, PhType, ConType] {
	b.internal.customUpdateFunc = f
	return b
}

// Build returns the status updater.
func (b *StatusUpdaterBuilder[Obj, PhType, ConType]) Build() *statusUpdater[Obj, PhType, ConType] {
	return b.internal
}

type StatusField string

const (
	// This is kind of a meta field, it holds the name of the field that stores the status itself.
	STATUS_FIELD                     StatusField = "Status"
	STATUS_FIELD_OBSERVED_GENERATION StatusField = "ObservedGeneration"
	STATUS_FIELD_LAST_RECONCILE_TIME StatusField = "LastReconcileTime"
	STATUS_FIELD_CONDITIONS          StatusField = "Conditions"
	STATUS_FIELD_REASON              StatusField = "Reason"
	STATUS_FIELD_MESSAGE             StatusField = "Message"
	STATUS_FIELD_PHASE               StatusField = "Phase"
)

// AllStatusFields returns all status fields that are used by the status updater.
// The meta field STATUS_FIELD is not included.
func AllStatusFields() []StatusField {
	return []StatusField{
		STATUS_FIELD_OBSERVED_GENERATION,
		STATUS_FIELD_LAST_RECONCILE_TIME,
		STATUS_FIELD_CONDITIONS,
		STATUS_FIELD_REASON,
		STATUS_FIELD_MESSAGE,
		STATUS_FIELD_PHASE,
	}
}

type statusUpdater[Obj client.Object, PhType ~string, ConType comparable] struct {
	fieldNames                map[StatusField]string
	phaseUpdateFunc           func(obj Obj, rr ReconcileResult[Obj, ConType]) (PhType, error)
	customUpdateFunc          func(obj Obj, rr ReconcileResult[Obj, ConType]) error
	conConstruct              func() conditions.Condition[ConType]
	removeUntouchedConditions bool
}

func newStatusUpdater[Obj client.Object, PhType ~string, ConType comparable]() *statusUpdater[Obj, PhType, ConType] {
	return &statusUpdater[Obj, PhType, ConType]{
		fieldNames: map[StatusField]string{
			STATUS_FIELD:                     string(STATUS_FIELD),
			STATUS_FIELD_OBSERVED_GENERATION: string(STATUS_FIELD_OBSERVED_GENERATION),
			STATUS_FIELD_LAST_RECONCILE_TIME: string(STATUS_FIELD_LAST_RECONCILE_TIME),
			STATUS_FIELD_CONDITIONS:          string(STATUS_FIELD_CONDITIONS),
			STATUS_FIELD_REASON:              string(STATUS_FIELD_REASON),
			STATUS_FIELD_MESSAGE:             string(STATUS_FIELD_MESSAGE),
			STATUS_FIELD_PHASE:               string(STATUS_FIELD_PHASE),
		},
		phaseUpdateFunc: defaultPhaseUpdateFunc[Obj, PhType, ConType],
	}
}

func defaultPhaseUpdateFunc[Obj client.Object, PhType ~string, ConType comparable](obj Obj, _ ReconcileResult[Obj, ConType]) (PhType, error) {
	// Default phase update function does nothing.
	// This should be overridden by the caller.
	var zero PhType
	return zero, nil
}

// UpdateStatus updates the status of the object in the given ReconcileResult, using the previously set field names and functions.
// The object is expected to be a pointer to a struct with the status field.
// If the 'Object' field in the ReconcileResult is nil, the status update becomes a no-op.
func (s *statusUpdater[Obj, PhType, ConType]) UpdateStatus(ctx context.Context, c client.Client, rr ReconcileResult[Obj, ConType]) (ctrl.Result, error) {
	if IsNil(rr.Object) {
		return rr.Result, nil
	}
	if s.fieldNames[STATUS_FIELD] == "" {
		return rr.Result, nil
	}
	if IsNil(rr.OldObject) || IsSameObject(rr.OldObject, rr.Object) {
		// create old object based on given one
		rr.OldObject = rr.Object.DeepCopyObject().(Obj)
	}
	status := GetField(rr.Object, s.fieldNames[STATUS_FIELD], true)
	if IsNil(status) {
		return rr.Result, fmt.Errorf("unable to get pointer to status field '%s' of object %T", s.fieldNames[STATUS_FIELD], rr.Object)
	}

	errs := errors.NewReasonableErrorList(rr.ReconcileError)
	now := time.Now()
	if s.fieldNames[STATUS_FIELD_LAST_RECONCILE_TIME] != "" {
		SetField(status, s.fieldNames[STATUS_FIELD_LAST_RECONCILE_TIME], metav1.NewTime(now))
	}
	if s.fieldNames[STATUS_FIELD_OBSERVED_GENERATION] != "" {
		SetField(status, s.fieldNames[STATUS_FIELD_OBSERVED_GENERATION], rr.Object.GetGeneration())
	}
	if s.fieldNames[STATUS_FIELD_MESSAGE] != "" {
		message := rr.Message
		if message == "" && rr.ReconcileError != nil {
			message = rr.ReconcileError.Error()
		}
		SetField(status, s.fieldNames[STATUS_FIELD_MESSAGE], message)
	}
	if s.fieldNames[STATUS_FIELD_REASON] != "" {
		reason := rr.Reason
		if reason == "" && rr.ReconcileError != nil {
			reason = rr.ReconcileError.Reason()
		}
		SetField(status, s.fieldNames[STATUS_FIELD_REASON], reason)
	}
	if s.fieldNames[STATUS_FIELD_CONDITIONS] != "" && s.conConstruct != nil {
		oldConsRaw := GetField(status, s.fieldNames[STATUS_FIELD_CONDITIONS], false)
		var oldCons []conditions.Condition[ConType]
		if !IsNil(oldConsRaw) {
			oldCons = oldConsRaw.([]conditions.Condition[ConType])
		}
		cu := conditions.ConditionUpdater(s.conConstruct, oldCons, s.removeUntouchedConditions)
		cu.Now = now
		for _, con := range rr.Conditions {
			cu.UpdateConditionFromTemplate(con)
		}
		newConsRaw, _ := cu.Conditions()
		conType := reflect.TypeOf(s.conConstruct())
		newCons := reflect.MakeSlice(reflect.SliceOf(conType), len(newConsRaw), len(newConsRaw)).Interface()
		for i := range newConsRaw {
			val := reflect.ValueOf(newConsRaw[i]).Convert(conType)
			reflect.ValueOf(newCons).Index(i).Set(val)
		}
		SetField(status, s.fieldNames[STATUS_FIELD_CONDITIONS], newCons)
	}
	if s.fieldNames[STATUS_FIELD_PHASE] != "" {
		phase, err := s.phaseUpdateFunc(rr.Object, rr)
		if err != nil {
			phase, _ = defaultPhaseUpdateFunc[Obj, PhType, ConType](rr.Object, rr)
			errs.Append(fmt.Errorf("error computing phase: %w", err))
		}
		SetField(status, s.fieldNames[STATUS_FIELD_PHASE], phase)
	}
	if s.customUpdateFunc != nil {
		if err := s.customUpdateFunc(rr.Object, rr); err != nil {
			errs.Append(fmt.Errorf("error performing custom status update: %w", err))
		}
	}

	// update status in cluster
	if err := c.Status().Patch(ctx, rr.Object, client.MergeFrom(rr.OldObject)); err != nil {
		errs.Append(fmt.Errorf("error patching status: %w", err))
	}

	return rr.Result, errs.Aggregate()
}

// GetField returns the value of the field with the given name from the given object.
// Nested fields can be accessed by separating them with '.' (e.g. "Foo.Bar").
// If pointer is true, it returns a pointer to the field value instead.
// WARNING: This function will panic if pointer is true but obj is not a pointer itself!
// Panics if the object is nil or the field is not found.
func GetField(obj any, field string, pointer bool) any {
	if obj == nil {
		panic("object is nil")
	}
	current, next, _ := strings.Cut(field, ".")
	val, ok := obj.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(obj)
	}
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	for i := range val.NumField() {
		if val.Type().Field(i).Name == current {
			res := val.Field(i)
			if next != "" {
				return GetField(res, next, pointer)
			}
			if pointer {
				res = res.Addr()
			}
			return res.Interface()
		}
	}
	panic(fmt.Sprintf("field '%s' not found in object %T", current, obj))
}

// SetField sets the field in the given object to the given value.
// Nested fields can be accessed by separating them with '.' (e.g. "Foo.Bar").
// Panics if the object is nil or the field is not found.
// WARNING: This function will panic if the specified field is not settable (e.g. because obj is not a pointer).
func SetField(obj any, field string, value any) {
	if obj == nil {
		panic("object is nil")
	}
	current, next, _ := strings.Cut(field, ".")
	val, ok := obj.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(obj)
	}
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	for i := range val.NumField() {
		if val.Type().Field(i).Name == current {
			res := val.Field(i)
			if next != "" {
				SetField(res, next, value)
				return
			}
			res.Set(reflect.ValueOf(value))
			return
		}
	}
	panic(fmt.Sprintf("field '%s' not found in object %T", current, obj))
}

// IsSameObject takes two interfaces of the same type and returns true if they both are pointers to the same underlying object.
func IsSameObject[T any](a, b T) bool {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)
	if aVal.Kind() != reflect.Ptr || bVal.Kind() != reflect.Ptr {
		return false
	}
	if aVal.IsNil() && bVal.IsNil() {
		return true
	}
	if aVal.IsNil() || bVal.IsNil() {
		return false
	}
	if aVal.Type() != bVal.Type() {
		return false
	}
	return aVal.Interface() == bVal.Interface()
}

// The result of a reconciliation.
// Obj is the k8s resource that is reconciled.
// ConType is the type of the "Status" field of the condition, usually a string alias. Simply use string if your object's status does not have conditions.
type ReconcileResult[Obj client.Object, ConType comparable] struct {
	// The old object, before it was modified.
	// Basically, if OldObject.Status differs from Object.Status, the status will be patched during status updating.
	// May be nil, in this case only the status metadata (observedGeneration etc.) is updated.
	// Changes to anything except the status are ignored.
	OldObject Obj
	// The current objectclient.Object	// If nil, the status update becomes a no-op.
	Object Obj
	// The result of the reconciliation.
	// Is simply passed through.
	Result ctrl.Result
	// The error that occurred during reconciliation, if any.
	ReconcileError errors.ReasonableError
	// The reason to display in the object's status.
	// If empty, it is taken from the error, if any.
	Reason string
	// The message to display in the object's status.
	// Potential error messages from the reconciliation error are appended.
	Message string
	// Conditions contains a list of conditions that should be updated on the object.
	// Also note that names of conditions are globally unique, so take care to avoid conflicts with other objects.
	// The lastTransition timestamp of the condition will be overwritten with the current time while updating.
	Conditions []conditions.Condition[ConType]
}
