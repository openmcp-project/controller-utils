package conditions

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCondition is an alias for apimeta.FindStatusCondition.
// It returns a pointer to the condition of the specified type from the given conditions slice, or nil if not found.
func GetCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	return apimeta.FindStatusCondition(conditions, conditionType)
}

// FromBoolPointer returns the metav1.ConditionStatus that matches the given bool pointer.
// nil = ConditionUnknown
// true = ConditionTrue
// false = ConditionFalse
func FromBoolPointer(status *bool) metav1.ConditionStatus {
	if status == nil {
		return metav1.ConditionUnknown
	}
	if *status {
		return metav1.ConditionTrue
	}
	return metav1.ConditionFalse
}

// FromBool returns the metav1.ConditionStatus that matches the given bool value.
// true = ConditionTrue
// false = ConditionFalse
func FromBool(status bool) metav1.ConditionStatus {
	return FromBoolPointer(&status)
}

// ToBoolPointer is the inverse of FromBoolPointer.
// It returns a pointer to a bool that matches the given ConditionStatus.
// If the status is ConditionTrue, it returns a pointer to true.
// If the status is ConditionFalse, it returns a pointer to false.
// If the status is ConditionUnknown or any unknown value, it returns nil.
func ToBoolPointer(status metav1.ConditionStatus) *bool {
	var res *bool
	switch status {
	case metav1.ConditionTrue:
		tmp := true
		res = &tmp
	case metav1.ConditionFalse:
		tmp := false
		res = &tmp
	}
	return res
}

// AllConditionsHaveStatus returns true if all conditions have the specified status.
func AllConditionsHaveStatus(status metav1.ConditionStatus, conditions ...metav1.Condition) bool {
	for _, con := range conditions {
		if con.Status != status {
			return false
		}
	}
	return true
}
