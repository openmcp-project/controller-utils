package conditions

import (
	"reflect"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"

	"github.com/openmcp-project/controller-utils/pkg/collections"
)

const EventReasonConditionChanged = "ConditionChanged"

type EventVerbosity string

const (
	// EventPerChange causes one event to be recorded for each condition that has changed.
	// This is the most verbose setting. The old and new status of each changed condition will be visible in the event message.
	EventPerChange EventVerbosity = "perChange"
	// EventPerNewStatus causes one event to be recorded for each new status that any condition has reached.
	// This means that at max four events will be recorded:
	// - the following conditions changed to True (including newly added conditions)
	// - the following conditions changed to False (including newly added conditions)
	// - the following conditions changed to Unknown (including newly added conditions)
	// - the following conditions were removed
	// The old status of the conditions will not be part of the event message.
	EventPerNewStatus EventVerbosity = "perNewStatus"
	// EventIfChanged causes a single event to be recorded if any condition's status has changed.
	// All changed conditions will be listed, but not their old or new status.
	EventIfChanged EventVerbosity = "ifChanged"
)

// conditionUpdater is a helper struct for updating a list of Conditions.
// Use the ConditionUpdater constructor for initializing.
type conditionUpdater struct {
	Now             metav1.Time
	conditions      map[string]metav1.Condition
	original        map[string]metav1.Condition
	eventRecoder    record.EventRecorder
	eventVerbosity  EventVerbosity
	updates         map[string]metav1.ConditionStatus
	removeUntouched bool
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
		Now:             metav1.Now(),
		conditions:      make(map[string]metav1.Condition, len(conditions)),
		updates:         make(map[string]metav1.ConditionStatus),
		removeUntouched: removeUntouched,
		original:        make(map[string]metav1.Condition, len(conditions)),
	}
	for _, con := range conditions {
		res.conditions[con.Type] = con
		res.original[con.Type] = con
	}
	return res
}

// WithEventRecorder enables event recording for condition changes.
// Note that this method must be called before any UpdateCondition calls, otherwise the events for the conditions will not be recorded.
// The verbosity argument controls how many events are recorded and what information they contain.
// If the event recorder is nil, no events will be recorded.
func (c *conditionUpdater) WithEventRecorder(recorder record.EventRecorder, verbosity EventVerbosity) *conditionUpdater {
	c.eventRecoder = recorder
	c.eventVerbosity = verbosity
	return c
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
	c.updates[conType] = status
	c.conditions[conType] = con
	return c
}

// UpdateConditionFromTemplate is a convenience wrapper around UpdateCondition which allows it to be called with a preconstructed ComponentCondition.
func (c *conditionUpdater) UpdateConditionFromTemplate(con metav1.Condition) *conditionUpdater {
	return c.UpdateCondition(con.Type, con.Status, con.ObservedGeneration, con.Reason, con.Message)
}

// HasCondition returns true if a condition with the given type exists in the updated condition list.
func (c *conditionUpdater) HasCondition(conType string) bool {
	_, ok := c.conditions[conType]
	_, updated := c.updates[conType]
	return ok && (!c.removeUntouched || updated)
}

// RemoveCondition removes the condition with the given type from the updated condition list.
func (c *conditionUpdater) RemoveCondition(conType string) *conditionUpdater {
	if !c.HasCondition(conType) {
		return c
	}
	delete(c.conditions, conType)
	delete(c.updates, conType)
	return c
}

// Conditions returns the updated condition list.
// If the condition updater was initialized with removeUntouched=true, this list will only contain the conditions which have been updated
// in between the condition updater creation and this method call. Otherwise, it will potentially also contain old conditions.
// The conditions are returned sorted by their type.
// The second return value indicates whether the condition list has actually changed.
func (c *conditionUpdater) Conditions() ([]metav1.Condition, bool) {
	res := collections.ProjectSlice(c.updatedConditions(), func(con metav1.Condition) metav1.Condition {
		if c.original[con.Type].Status == con.Status {
			// if the status has not changed, reset the LastTransitionTime to the original value
			con.LastTransitionTime = c.original[con.Type].LastTransitionTime
		}
		return con
	})
	slices.SortStableFunc(res, func(a, b metav1.Condition) int {
		return strings.Compare(a.Type, b.Type)
	})
	return res, c.changed(res)
}

func (c *conditionUpdater) updatedConditions() []metav1.Condition {
	res := make([]metav1.Condition, 0, len(c.conditions))
	for _, con := range c.conditions {
		if _, updated := c.updates[con.Type]; updated || !c.removeUntouched {
			res = append(res, con)
		}
	}
	return res
}

func (c *conditionUpdater) changed(newCons []metav1.Condition) bool {
	if len(c.original) != len(newCons) {
		return true
	}
	for _, newCon := range newCons {
		oldCon, found := c.original[newCon.Type]
		if !found || !reflect.DeepEqual(newCon, oldCon) {
			return true
		}
	}
	return false
}

// Record records events for the updated conditions on the given object.
// Which events are recorded depends on the eventVerbosity setting.
// In any setting, events are only recorded for conditions that have somehow changed.
// This is a no-op if either the event recorder or the given object is nil.
// Note that events will be duplicated if this method is called multiple times.
// Returns the receiver for easy chaining.
func (c *conditionUpdater) Record(obj runtime.Object) *conditionUpdater {
	if c.eventRecoder == nil || obj == nil {
		return c
	}

	updatedCons := c.updatedConditions()
	if !c.changed(updatedCons) {
		// nothing to do if there are no changes
		return c
	}
	lostCons := collections.ProjectMapToMap(c.original, func(conType string, con metav1.Condition) (string, metav1.ConditionStatus) {
		return conType, con.Status
	})
	for _, con := range updatedCons {
		delete(lostCons, con.Type)
	}

	switch c.eventVerbosity {
	case EventPerChange:
		for _, con := range updatedCons {
			oldCon, found := c.original[con.Type]
			if !found {
				c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "Condition '%s' added with status '%s'", con.Type, con.Status)
				continue
			}
			if con.Status != oldCon.Status {
				c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "Condition '%s' changed from '%s' to '%s'", con.Type, oldCon.Status, con.Status)
				continue
			}
		}
		for conType, oldStatus := range lostCons {
			c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "Condition '%s' with status '%s' removed", conType, oldStatus)
		}

	case EventPerNewStatus:
		trueCons := sets.New[string]()
		falseCons := sets.New[string]()
		unknownCons := sets.New[string]()

		for _, con := range updatedCons {
			// only add conditions that have changed
			if oldCon, found := c.original[con.Type]; found && con.Status == oldCon.Status {
				continue
			}
			switch con.Status {
			case metav1.ConditionTrue:
				trueCons.Insert(con.Type)
			case metav1.ConditionFalse:
				falseCons.Insert(con.Type)
			case metav1.ConditionUnknown:
				unknownCons.Insert(con.Type)
			}
		}

		if trueCons.Len() > 0 {
			c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "The following conditions changed to 'True': %s", strings.Join(sets.List(trueCons), ", "))
		}
		if falseCons.Len() > 0 {
			c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "The following conditions changed to 'False': %s", strings.Join(sets.List(falseCons), ", "))
		}
		if unknownCons.Len() > 0 {
			c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "The following conditions changed to 'Unknown': %s", strings.Join(sets.List(unknownCons), ", "))
		}
		if len(lostCons) > 0 {
			c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "The following conditions were removed: %s", strings.Join(sets.List(sets.KeySet(lostCons)), ", "))
		}

	case EventIfChanged:
		changedCons := sets.New[string]()
		for _, con := range updatedCons {
			if oldCon, found := c.original[con.Type]; !found || con.Status != oldCon.Status {
				changedCons.Insert(con.Type)
			}
		}
		for conType := range lostCons {
			changedCons.Insert(conType)
		}
		if changedCons.Len() > 0 {
			c.eventRecoder.Eventf(obj, corev1.EventTypeNormal, EventReasonConditionChanged, "The following conditions have changed: %s", strings.Join(sets.List(changedCons), ", "))
		}
	}

	ns := &corev1.Namespace{}
	ns.GetObjectKind()

	return c
}
