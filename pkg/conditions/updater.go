package conditions

import (
	"slices"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// conditionUpdater is a helper struct for updating a list of Conditions.
// Use the ConditionUpdater constructor for initializing.
type conditionUpdater struct {
	Now        metav1.Time
	conditions map[string]metav1.Condition
	updated    sets.Set[string]
	changed    bool
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
func ConditionUpdater(conditions []metav1.Condition, removeUntouched bool) *conditionUpdater {
	res := &conditionUpdater{
		Now:        metav1.Now(),
		conditions: make(map[string]metav1.Condition, len(conditions)),
		changed:    false,
	}
	for _, con := range conditions {
		res.conditions[con.Type] = con
	}
	if removeUntouched {
		res.updated = sets.New[string]()
	}
	return res
}

// UpdateCondition updates or creates the condition with the specified type.
// All fields of the condition are updated with the values given in the arguments, but the condition's LastTransitionTime is only updated (with the timestamp contained in the receiver struct) if the status changed.
// Returns the receiver for easy chaining.
func (c *conditionUpdater) UpdateCondition(conType string, status metav1.ConditionStatus, observedGeneration int64, reason, message string) *conditionUpdater {
	con := metav1.Condition{
		Type:               conType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: c.Now,
	}
	old, ok := c.conditions[conType]
	if ok && old.Status == con.Status {
		// update LastTransitionTime only if status changed
		con.LastTransitionTime = old.LastTransitionTime
	}
	if !c.changed {
		if ok {
			c.changed = old.Status != con.Status || old.Reason != con.Reason || old.Message != con.Message
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
func (c *conditionUpdater) UpdateConditionFromTemplate(con metav1.Condition) *conditionUpdater {
	return c.UpdateCondition(con.Type, con.Status, con.ObservedGeneration, con.Reason, con.Message)
}

// HasCondition returns true if a condition with the given type exists in the updated condition list.
func (c *conditionUpdater) HasCondition(conType string) bool {
	_, ok := c.conditions[conType]
	return ok && (c.updated == nil || c.updated.Has(conType))
}

// RemoveCondition removes the condition with the given type from the updated condition list.
func (c *conditionUpdater) RemoveCondition(conType string) *conditionUpdater {
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
// The second return value indicates whether the condition list has actually changed.
func (c *conditionUpdater) Conditions() ([]metav1.Condition, bool) {
	res := make([]metav1.Condition, 0, len(c.conditions))
	for _, con := range c.conditions {
		if c.updated == nil {
			res = append(res, con)
			continue
		}
		if c.updated.Has(con.Type) {
			res = append(res, con)
		} else {
			c.changed = true
		}
	}
	slices.SortStableFunc(res, func(a, b metav1.Condition) int {
		return strings.Compare(a.Type, b.Type)
	})
	return res, c.changed
}
