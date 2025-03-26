package conditions

import (
	"slices"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

// conditionUpdater is a helper struct for updating a list of Conditions.
// Use the ConditionUpdater constructor for initializing.
type conditionUpdater[T comparable] struct {
	Now         time.Time
	conditions  map[string]Condition[T]
	updated     sets.Set[string]
	constructor func() Condition[T]
	changed     bool
}

// ConditionUpdater creates a builder-like helper struct for updating a list of Conditions.
// The 'constructor' argument is a function that returns a new (empty) instance of the condition implementation type.
// The 'conditions' argument contains the old condition list.
// If removeUntouched is true, the condition list returned with Conditions() will have all conditions removed that have not been updated.
// If false, all conditions will be kept.
// Note that calling this function stores the current time as timestamp that is used as timestamp if a condition's status changed.
// To overwrite this timestamp, modify the 'Now' field of the returned struct manually.
//
// The given condition list is not modified.
//
// Usage example:
// status.conditions = ConditionUpdater(status.conditions, true).UpdateCondition(...).UpdateCondition(...).Conditions()
func ConditionUpdater[T comparable](constructor func() Condition[T], conditions []Condition[T], removeUntouched bool) *conditionUpdater[T] {
	res := &conditionUpdater[T]{
		Now:         time.Now(),
		conditions:  make(map[string]Condition[T], len(conditions)),
		constructor: constructor,
		changed:     false,
	}
	for _, con := range conditions {
		res.conditions[con.GetType()] = con
	}
	if removeUntouched {
		res.updated = sets.New[string]()
	}
	return res
}

// UpdateCondition updates or creates the condition with the specified type.
// All fields of the condition are updated with the values given in the arguments, but the condition's LastTransitionTime is only updated (with the timestamp contained in the receiver struct) if the status changed.
// Returns the receiver for easy chaining.
func (c *conditionUpdater[T]) UpdateCondition(conType string, status T, reason, message string) *conditionUpdater[T] {
	con := c.constructor()
	con.SetType(conType)
	con.SetStatus(status)
	con.SetReason(reason)
	con.SetMessage(message)
	con.SetLastTransitionTime(c.Now)
	old, ok := c.conditions[conType]
	if ok && old.GetStatus() == con.GetStatus() {
		// update LastTransitionTime only if status changed
		con.SetLastTransitionTime(old.GetLastTransitionTime())
	}
	if !c.changed {
		if ok {
			c.changed = old.GetStatus() != con.GetStatus() || old.GetReason() != con.GetReason() || old.GetMessage() != con.GetMessage()
		} else {
			c.changed = true
		}
	}
	c.conditions[conType] = con
	if c.updated != nil {
		c.updated.Insert(conType)
	}
	return c
}

// UpdateConditionFromTemplate is a convenience wrapper around UpdateCondition which allows it to be called with a preconstructed ComponentCondition.
func (c *conditionUpdater[T]) UpdateConditionFromTemplate(con Condition[T]) *conditionUpdater[T] {
	return c.UpdateCondition(con.GetType(), con.GetStatus(), con.GetReason(), con.GetMessage())
}

// HasCondition returns true if a condition with the given type exists in the updated condition list.
func (c *conditionUpdater[T]) HasCondition(conType string) bool {
	_, ok := c.conditions[conType]
	return ok && (c.updated == nil || c.updated.Has(conType))
}

// RemoveCondition removes the condition with the given type from the updated condition list.
func (c *conditionUpdater[T]) RemoveCondition(conType string) *conditionUpdater[T] {
	if !c.HasCondition(conType) {
		return c
	}
	delete(c.conditions, conType)
	if c.updated != nil {
		c.updated.Delete(conType)
	}
	c.changed = true
	return c
}

// Conditions returns the updated condition list.
// If the condition updater was initialized with removeUntouched=true, this list will only contain the conditions which have been updated
// in between the condition updater creation and this method call. Otherwise, it will potentially also contain old conditions.
// The conditions are returned sorted by their type.
func (c *conditionUpdater[T]) Conditions() ([]Condition[T], bool) {
	res := make([]Condition[T], 0, len(c.conditions))
	for _, con := range c.conditions {
		if c.updated == nil {
			res = append(res, con)
			continue
		}
		if c.updated.Has(con.GetType()) {
			res = append(res, con)
		} else {
			c.changed = true
		}
	}
	slices.SortStableFunc(res, func(a, b Condition[T]) int {
		return strings.Compare(a.GetType(), b.GetType())
	})
	return res, c.changed
}
