package conditions

import "time"

// Condition represents a condition consisting of type, status, reason, message and a last transition timestamp.
type Condition[T comparable] interface {
	// SetStatus sets the status of the condition.
	SetStatus(status T)
	// GetStatus returns the status of the condition.
	GetStatus() T

	// SetType sets the type of the condition.
	SetType(conType string)
	// GetType returns the type of the condition.
	GetType() string

	// SetLastTransitionTime sets the timestamp of the condition.
	SetLastTransitionTime(timestamp time.Time)
	// GetLastTransitionTime returns the timestamp of the condition.
	GetLastTransitionTime() time.Time

	// SetReason sets the reason of the condition.
	SetReason(reason string)
	// GetReason returns the reason of the condition.
	GetReason() string

	// SetMessage sets the message of the condition.
	SetMessage(message string)
	// GetMessage returns the message of the condition.
	GetMessage() string
}

// GetCondition returns a pointer to the condition for the given type, if it exists.
// Otherwise, nil is returned.
func GetCondition[T comparable](ccl []Condition[T], t string) Condition[T] {
	for i := range ccl {
		if ccl[i].GetType() == t {
			return ccl[i]
		}
	}
	return nil
}

// AllConditionsHaveStatus returns true if all conditions have the specified status.
func AllConditionsHaveStatus[T comparable](status T, conditions ...Condition[T]) bool {
	for _, con := range conditions {
		if con.GetStatus() != status {
			return false
		}
	}
	return true
}
