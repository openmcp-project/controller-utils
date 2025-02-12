package collections

import (
	"fmt"

	"github.tools.sap/CoLa/controller-utils/pkg/collections/iterators"
)

var _ Collection[any] = &abstractCollection[any]{}

var ErrNotImplemented = fmt.Errorf("not implemented")

type abstractCollection[T any] struct {
	// workarounds so that the functions that are implemented here use the actual implementations and not this struct's ones
	funcIterator func() iterators.Iterator[T]
	funcAdd      func(elements ...T) bool
	funcClear    func()
	funcContains func(element T) bool
	funcRemove   func(elements ...T) bool
	funcRemoveIf func(filter Predicate[T]) bool
	funcSize     func() int
	funcNew      func() Collection[T]
}

func (ac *abstractCollection[T]) Add(elements ...T) bool {
	if ac.funcAdd != nil {
		return ac.funcAdd(elements...)
	}
	panic(ErrNotImplemented)
}

// AddAll adds all elements from the given collection to this one.
// Returns true if the collection changed as a result of the operation.
func (ac *abstractCollection[T]) AddAll(c Collection[T]) bool {
	it := c.Iterator()
	changed := false
	for it.HasNext() {
		changed = ac.Add(it.Next()) || changed
	}
	return changed
}

// Clear removes all of the elements from this collection.
func (ac *abstractCollection[T]) Clear() {
	if ac.funcClear != nil {
		ac.funcClear()
		return
	}
	panic(ErrNotImplemented)
}

// Returns true if this collection contains the specified element.
func (ac *abstractCollection[T]) Contains(element T) bool {
	if ac.funcContains != nil {
		return ac.funcContains(element)
	}
	it := ac.Iterator()
	for it.HasNext() {
		if Equals(element, it.Next()) {
			return true
		}
	}
	return false
}

// ContainsAll returns true if this collection contains all of the elements in the specified collection.
func (ac *abstractCollection[T]) ContainsAll(c Collection[T]) bool {
	it := c.Iterator()
	for it.HasNext() {
		if !ac.Contains(it.Next()) {
			return false
		}
	}
	return true
}

// Compares the specified object with this collection for equality.
func (ac *abstractCollection[T]) Equals(c Collection[T]) bool {
	if ac.Size() != c.Size() {
		return false
	}
	lit := ac.Iterator()
	cit := c.Iterator()
	for lit.HasNext() {
		if !Equals(lit.Next(), cit.Next()) {
			return false
		}
	}
	return true
}

// IsEmpty returns true if this collection contains no elements.
func (ac *abstractCollection[T]) IsEmpty() bool {
	return ac.Size() == 0
}

// Returns an iterator over the elements in this collection.
func (ac *abstractCollection[T]) Iterator() iterators.Iterator[T] {
	if ac.funcIterator != nil {
		return ac.funcIterator()
	}
	panic(ErrNotImplemented)
}

// Removes all given elements from the collection, if they are present.
// Each specified element is removed only once, even if it is contained multiple times in the collection.
// Returns true if the collection changed as a result of the operation.
func (ac *abstractCollection[T]) Remove(elements ...T) bool {
	if ac.funcRemove != nil {
		return ac.funcRemove(elements...)
	}
	panic(ErrNotImplemented)
}

// RemoveAllOf removes all instances of the given elements from the collection.
// Returns true if the collection changed as a result of the operation.
func (ac *abstractCollection[T]) RemoveAllOf(elements ...T) bool {
	panic(ErrNotImplemented)
}

// Removes all of this collection's elements that are also contained in the specified collection.
// Each element is removed only as often as it is in the specified collection (or less).
// Returns true if the collection changed as a result of the operation.
func (ac *abstractCollection[T]) RemoveAll(c Collection[T]) bool {
	it := c.Iterator()
	changed := false
	for it.HasNext() {
		changed = ac.Remove(it.Next()) || changed
	}
	return changed
}

// RetainAll removes all elements from the collection that are not contained in the specified collection.
// Returns true if the collection changed as a result of the operation.
func (ac *abstractCollection[T]) RetainAll(c Collection[T]) bool {
	return ac.RemoveIf(func(e T) bool { return !c.Contains(e) })
}

// RemoveIf removes all elements from the collection that satisfy the given predicate.
// Returns true if the collection changed as a result of the operation.
func (ac *abstractCollection[T]) RemoveIf(filter Predicate[T]) bool {
	if ac.funcRemoveIf != nil {
		return ac.funcRemoveIf(filter)
	}
	panic(ErrNotImplemented)
}

// Size returns the number of elements in this collection.
func (ac *abstractCollection[T]) Size() int {
	if ac.funcSize != nil {
		return ac.funcSize()
	}
	panic(ErrNotImplemented)
}

// ToSlice returns a slice containing the elements in this collection.
func (ac *abstractCollection[T]) ToSlice() []T {
	res := make([]T, ac.Size())
	it := ac.Iterator()
	for i := range res {
		res[i] = it.Next()
	}
	return res
}

// New returns a new Collection of the same type.
func (ac *abstractCollection[T]) New() Collection[T] {
	if ac.funcNew != nil {
		return ac.funcNew()
	}
	panic(ErrNotImplemented)
}
