package controller

// This package contains predicates which can be used for constructing controllers.

import (
	"reflect"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//////////////////
/// CONVERSION ///
//////////////////

// ToTypedPredicate wraps a predicate.Predicate (which is an alias for predicate.TypedPredicate[client.Object]) into a predicate.TypedPredicate[Obj].
func ToTypedPredicate[Obj client.Object](p predicate.Predicate) predicate.TypedPredicate[Obj] {
	return &typedPredicateConverter[Obj]{internal: p}
}

type typedPredicateConverter[Obj client.Object] struct {
	internal predicate.Predicate
}

var _ predicate.TypedPredicate[client.Object] = &typedPredicateConverter[client.Object]{}

func (t *typedPredicateConverter[Obj]) Create(e event.TypedCreateEvent[Obj]) bool {
	return t.internal.Create(event.TypedCreateEvent[client.Object]{
		Object: e.Object,
	})
}

func (t *typedPredicateConverter[Obj]) Delete(e event.TypedDeleteEvent[Obj]) bool {
	return t.internal.Delete(event.TypedDeleteEvent[client.Object]{
		Object:             e.Object,
		DeleteStateUnknown: e.DeleteStateUnknown,
	})
}

func (t *typedPredicateConverter[Obj]) Generic(e event.TypedGenericEvent[Obj]) bool {
	return t.internal.Generic(event.TypedGenericEvent[client.Object]{
		Object: e.Object,
	})
}

func (t *typedPredicateConverter[Obj]) Update(e event.TypedUpdateEvent[Obj]) bool {
	return t.internal.Update(event.TypedUpdateEvent[client.Object]{
		ObjectOld: e.ObjectOld,
		ObjectNew: e.ObjectNew,
	})
}

/////////////////////////////////////
/// DELETION TIMESTAMP PREDICATES ///
/////////////////////////////////////

// DeletionTimestampChangedPredicate reacts to changes of the deletion timestamp.
type DeletionTimestampChangedPredicate struct {
	predicate.Funcs
}

var _ predicate.Predicate = DeletionTimestampChangedPredicate{}

func (DeletionTimestampChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	oldDel := e.ObjectOld.GetDeletionTimestamp()
	newDel := e.ObjectNew.GetDeletionTimestamp()
	return !reflect.DeepEqual(newDel, oldDel)
}

///////////////////////////////////////
/// ANNOTATION AND LABEL PREDICATES ///
///////////////////////////////////////

func hasMetadataEntryPredicate(mType metadataEntryType, key, val string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if obj == nil {
			return false
		}
		actual, ok := getMetadataEntry(mType, obj, key)
		return ok && (val == "" || actual == val)
	})
}

type metadataEntryChangedPredicate struct {
	predicate.Funcs
	mType metadataEntryType // type of metadata entry (annotation or label)
	key   string            // key of the metadata entry
	value string            // value of the metadata entry (empty string if value doesn't matter)
	mod   int               // positive means the predicate returns true if the entry was added (includes being set to correct value), negative if it was removed, 0 if either happened
}

func (p metadataEntryChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldValue, ok := getMetadataEntry(p.mType, e.ObjectOld, p.key)
	oldHasEntry := ok && (p.value == "" || oldValue == p.value)

	newValue, ok := getMetadataEntry(p.mType, e.ObjectNew, p.key)
	newHasEntry := ok && (p.value == "" || newValue == p.value)

	if p.mod > 0 {
		// check if entry was added
		return !oldHasEntry && newHasEntry
	} else if p.mod < 0 {
		// check if entry was removed
		return oldHasEntry && !newHasEntry
	}
	// check if entry was changed (added, removed, or value changed)
	return oldHasEntry != newHasEntry
}

// HasAnnotationPredicate reacts if the resource has the specified annotation.
// If val is empty, the value of the annotation doesn't matter, only its existence.
// Otherwise, true is only returned if the annotation has the specified value.
// Note that GotAnnotationPredicate can be used to check if a resource just got a specific annotation.
func HasAnnotationPredicate(key, val string) predicate.Predicate {
	return hasMetadataEntryPredicate(ANNOTATION, key, val)
}

// GotAnnotationPredicate reacts if the specified annotation was added to the resource.
// If val is empty, the value of the annotation doesn't matter, just that it was added.
// Otherwise, true is only returned if the annotation was added (or changed) with the specified value.
func GotAnnotationPredicate(key, val string) predicate.Predicate {
	return metadataEntryChangedPredicate{
		mType: ANNOTATION,
		key:   key,
		value: val,
		mod:   1,
	}
}

// LostAnnotationPredicate reacts if the specified annotation was removed from the resource.
// If val is empty, this predicate returns true if the annotation was removed completely, independent of which value it had.
// Otherwise, true is returned if the annotation had the specified value before and now lost it, either by being removed or by being set to a different value.
func LostAnnotationPredicate(key, val string) predicate.Predicate {
	return metadataEntryChangedPredicate{
		mType: ANNOTATION,
		key:   key,
		value: val,
		mod:   -1,
	}
}

// HasLabelPredicate reacts if the resource has the specified label.
// If val is empty, the value of the label doesn't matter, only its existence.
// Otherwise, true is only returned if the label has the specified value.
// Note that GotLabelPredicate can be used to check if a resource just got a specific label.
func HasLabelPredicate(key, val string) predicate.Predicate {
	return hasMetadataEntryPredicate(LABEL, key, val)
}

// GotLabelPredicate reacts if the specified label was added to the resource.
// If val is empty, the value of the label doesn't matter, just that it was added.
// Otherwise, true is only returned if the label was added (or changed) with the specified value.
func GotLabelPredicate(key, val string) predicate.Predicate {
	return metadataEntryChangedPredicate{
		mType: LABEL,
		key:   key,
		value: val,
		mod:   1,
	}
}

// LostLabelPredicate reacts if the specified label was removed from the resource.
// If val is empty, this predicate returns true if the label was removed completely, independent of which value it had.
// Otherwise, true is returned if the label had the specified value before and now lost it, either by being removed or by being set to a different value.
func LostLabelPredicate(key, val string) predicate.Predicate {
	return metadataEntryChangedPredicate{
		mType: LABEL,
		key:   key,
		value: val,
		mod:   -1,
	}
}

// LabelSelectorPredicate returns a predicate based on a label selector.
// Opposed to the similarly named function from the controller-runtime library, this one works on label.Selector
// instead of metav1.LabelSelector.
func LabelSelectorPredicate(sel labels.Selector) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if obj == nil {
			return false
		}
		ls := obj.GetLabels()
		if ls == nil {
			ls = map[string]string{}
		}
		return sel.Matches(labels.Set(ls))
	})
}

/////////////////////////
/// STATUS PREDICATES ///
/////////////////////////

var _ predicate.Predicate = StatusChangedPredicate{}

// StatusChangedPredicate returns true if the object's status changed.
// Getting the status is done via reflection and only works if the corresponding field is named 'Status'.
// If getting the status fails, this predicate always returns true.
type StatusChangedPredicate struct {
	predicate.Funcs
}

func (p StatusChangedPredicate) Update(e event.UpdateEvent) bool {
	oldStatus := GetField(e.ObjectOld, "Status", false)
	newStatus := GetField(e.ObjectNew, "Status", false)
	if oldStatus == nil || newStatus == nil {
		return true
	}
	return !reflect.DeepEqual(oldStatus, newStatus)
}

////////////////////////////////////
/// IDENTITY MATCHING PREDICATES ///
////////////////////////////////////

// ExactNamePredicate returns true if the object's name and namespace exactly match the specified values.
// The namespace can be set to '*' to match any namespace.
func ExactNamePredicate(name, namespace string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if obj == nil {
			return false
		}
		return obj.GetName() == name && (namespace == "*" || obj.GetNamespace() == namespace)
	})
}
