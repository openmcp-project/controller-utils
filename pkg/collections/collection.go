package collections

import (
	"fmt"
	"reflect"

	"github.com/openmcp-project/controller-utils/pkg/collections/iterators"
)

// Collection represents a collection of elements.
// Note that T must be either comparable or implement the Comparable interface, otherwise runtime panics could occur.
type Collection[T any] interface {
	iterators.Iterable[T]

	// Add ensures that this collection contains the specified elements.
	// Returns true if the collection changed as a result of the operation.
	// (returns false if the collection does not support duplicates and already contained all given elements).
	Add(elements ...T) bool

	// AddAll adds all elements from the given collection to this one.
	// Returns true if the collection changed as a result of the operation.
	AddAll(c Collection[T]) bool

	// Clear removes all of the elements from this collection.
	Clear()

	// Returns true if this collection contains the specified element.
	Contains(element T) bool

	// ContainsAll returns true if this collection contains all of the elements in the specified collection.
	ContainsAll(c Collection[T]) bool

	// Compares the specified object with this collection for equality.
	Equals(c Collection[T]) bool

	// IsEmpty returns true if this collection contains no elements.
	IsEmpty() bool

	// Removes all given elements from the collection, if they are present.
	// Each specified element is removed only once, even if it is contained multiple times in the collection.
	// Returns true if the collection changed as a result of the operation.
	Remove(elements ...T) bool

	// RemoveAllOf removes all instances of the given elements from the collection.
	// Returns true if the collection changed as a result of the operation.
	RemoveAllOf(elements ...T) bool

	// Removes all of this collection's elements that are also contained in the specified collection.
	// Returns true if the collection changed as a result of the operation.
	RemoveAll(c Collection[T]) bool

	// RetainAll removes all elements from the collection that are not contained in the specified collection.
	// Returns true if the collection changed as a result of the operation.
	RetainAll(c Collection[T]) bool

	// RemoveIf removes all elements from the collection that satisfy the given predicate.
	// Returns true if the collection changed as a result of the operation.
	RemoveIf(filter Predicate[T]) bool

	// Size returns the number of elements in this collection.
	Size() int

	// ToSlice returns a slice containing the elements in this collection.
	ToSlice() []T

	// New returns a new Collection of the same type.
	New() Collection[T]
}

type Predicate[T any] func(T) bool

// Comparable is an interface that can be implemented to use types as generic argument which don't satisfy the 'comparable' constraint.
type Comparable[T any] interface {
	// Equals returns true if the receiver and the argument are equal.
	Equals(T) bool
}

// Equals compares two values of the same type.
// If the type implements the Comparable[T] interface, its Equals method is used for comparison.
// Otherwise, the standard comparison is used.
// Panics if the type is not comparable and does not implement Comparable[T].
func Equals[T any](a, b T) bool {
	aComp, ok := any(a).(Comparable[T])
	if ok {
		return aComp.Equals(b)
	}
	valOfA := reflect.ValueOf(a)
	if valOfA.Comparable() {
		valOfB := reflect.ValueOf(b)
		return valOfA.Equal(valOfB)
	}
	panic(fmt.Sprintf("type '%s' is neither comparable nor does it implement the Comparable interface", valOfA.Type().String()))
}
