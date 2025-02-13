package collections

import (
	cerr "github.tools.sap/CoLa/controller-utils/pkg/collections/errors"
	"github.tools.sap/CoLa/controller-utils/pkg/collections/iterators"
)

var _ List[any] = &LinkedList[any]{}
var _ Queue[any] = &LinkedList[any]{}

type LinkedList[T any] struct {
	abstractList[T]
	// dummy is start and end of the list
	// dummy.next points to the first element of the list
	// dummy.prev points to the last element of the list
	// if the list is empty, dummy points to itself
	dummy *llElement[T]
	size  int
}

var _ iterators.LinkedIteratorElement[any] = &llElement[any]{}

type llElement[T any] struct {
	prev, next *llElement[T]
	value      T
}

func (e *llElement[T]) Next() iterators.LinkedIteratorElement[T] {
	return e.next
}

func (e *llElement[T]) Value() T {
	return e.value
}

func NewLinkedList[T any](elements ...T) *LinkedList[T] {
	res := &LinkedList[T]{
		dummy: &llElement[T]{},
	}
	res.dummy.next = res.dummy
	res.dummy.prev = res.dummy

	res.abstractCollection.funcIterator = res.Iterator
	res.funcAdd = res.Add
	res.funcClear = res.Clear
	res.funcRemove = res.Remove
	res.funcRemoveIf = res.RemoveIf
	res.funcSize = res.Size

	res.Add(elements...)

	return res
}

func NewLinkedListFromCollection[T any](c Collection[T]) *LinkedList[T] {
	res := NewLinkedList[T]()
	res.AddAll(c)
	return res
}

///////////////////////////////
// COLLECTION IMPLEMENTATION //
///////////////////////////////

// Add ensures that this collection contains the specified elements.
// Returns true if the collection changed as a result of the operation.
// (returns false if the collection does not support duplicates and already contained all given elements).
func (l *LinkedList[T]) Add(elements ...T) bool {
	if len(elements) == 0 {
		return false
	}
	for i := range elements {
		elem := &llElement[T]{
			value: elements[i],
		}
		l.dummy.prev.next = elem
		elem.prev = l.dummy.prev
		elem.next = l.dummy
		l.dummy.prev = elem
		l.size++
	}
	return true
}

// AddAll adds all elements from the given collection to this one.
// Returns true if the collection changed as a result of the operation.
func (l *LinkedList[T]) AddAll(c Collection[T]) bool {
	it := c.Iterator()
	changed := false
	for it.HasNext() {
		changed = l.Add(it.Next()) || changed
	}
	return changed
}

// Clear removes all of the elements from this collection.
func (l *LinkedList[T]) Clear() {
	l.dummy.next = l.dummy
	l.dummy.prev = l.dummy
	l.size = 0
}

// Returns an iterator over the elements in this collection.
func (l *LinkedList[T]) Iterator() iterators.Iterator[T] {
	dummy := l.dummy.prev.Next()
	return iterators.NewLinkedIterator[T](l.dummy.next, func(e iterators.LinkedIteratorElement[T]) bool { return e != dummy })
}

// Removes all given elements from the collection, if they are present.
// Each specified element is removed only once, even if it is contained multiple times in the collection.
// Returns true if the collection changed as a result of the operation.
func (l *LinkedList[T]) Remove(elements ...T) bool {
	return l.remove(false, elements...)
}

// RemoveAllOf removes all instances of the given elements from the collection.
// Returns true if the collection changed as a result of the operation.
func (l *LinkedList[T]) RemoveAllOf(elements ...T) bool {
	return l.remove(true, elements...)
}

// RemoveIf removes all elements from the collection that satisfy the given predicate.
// Returns true if the collection changed as a result of the operation.
func (l *LinkedList[T]) RemoveIf(filter Predicate[T]) bool {
	oldSize := l.size
	elem := l.dummy.next
	for elem != l.dummy {
		if filter(elem.value) {
			elem.remove()
			l.size--
		}
		elem = elem.next
	}
	return l.size != oldSize
}

// Size returns the number of elements in this collection.
func (l *LinkedList[T]) Size() int {
	return l.size
}

// ToSlice returns a slice containing the elements in this collection.
func (l *LinkedList[T]) ToSlice() []T {
	res := make([]T, l.size)
	elem := l.dummy.next
	i := 0
	for elem != l.dummy {
		res[i] = elem.value
		i++
		elem = elem.next
	}
	return res
}

/////////////////////////
// LIST IMPLEMENTATION //
/////////////////////////

func (l *LinkedList[T]) AddIndex(element T, idx int) error {
	if idx == l.size {
		l.Add(element)
		return nil
	}
	elem := l.elementAt(idx)
	if elem == nil {
		return cerr.NewIndexOutOfBoundsError(idx)
	}
	newElem := &llElement[T]{
		value: element,
		prev:  elem.prev,
		next:  elem,
	}
	elem.prev.next = newElem
	elem.prev = newElem
	l.size++
	return nil
}

func (l *LinkedList[T]) RemoveIndex(idx int) error {
	elem := l.elementAt(idx)
	if elem == nil {
		return cerr.NewIndexOutOfBoundsError(idx)
	}
	elem.remove()
	l.size--
	return nil
}

func (l *LinkedList[T]) Get(idx int) (T, error) {
	var res T
	elem := l.elementAt(idx)
	if elem == nil {
		return res, cerr.NewIndexOutOfBoundsError(idx)
	}
	return elem.value, nil
}

//////////////////////////
// QUEUE IMPLEMENTATION //
//////////////////////////

func (l *LinkedList[T]) Push(elements ...T) error {
	l.Add(elements...)
	return nil
}

func (l *LinkedList[T]) Peek() T {
	if l.IsEmpty() {
		var zero T
		return zero
	}
	return l.dummy.next.value
}

func (l *LinkedList[T]) Element() (T, error) {
	if l.IsEmpty() {
		var zero T
		return zero, cerr.NewCollectionEmptyError()
	}
	return l.dummy.next.value, nil
}

func (l *LinkedList[T]) Poll() T {
	if l.IsEmpty() {
		var zero T
		return zero
	}
	elem := l.dummy.next
	elem.remove()
	l.size--
	return elem.value
}

func (l *LinkedList[T]) Fetch() (T, error) {
	if l.IsEmpty() {
		var zero T
		return zero, cerr.NewCollectionEmptyError()
	}
	elem := l.dummy.next
	elem.remove()
	l.size--
	return elem.value, nil
}

// New returns a new Collection of the same type.
func (l *LinkedList[T]) New() Collection[T] {
	return NewLinkedList[T]()
}

/////////////////////////
// AUXILIARY FUNCTIONS //
/////////////////////////

// remove removes the element itself from the chain by linking its previous element to its next element.
// The element's pointers are not modified.
func (e *llElement[T]) remove() {
	e.prev.next = e.next
	e.next.prev = e.prev
}

// remove removes all specified elements.
// If all is true, all instances of these elements are removed, otherwise only one.
func (l *LinkedList[T]) remove(all bool, elements ...T) bool {
	oldSize := l.size
	for _, tbr := range elements {
		elem := l.dummy.next
		for elem != l.dummy {
			if Equals(elem.value, tbr) {
				elem.remove()
				l.size--
				if !all {
					break
				}
			}
			elem = elem.next
		}
	}
	return l.size != oldSize
}

func (l *LinkedList[T]) elementAt(idx int) *llElement[T] {
	if idx < 0 || idx >= l.size {
		return nil
	}
	elem := l.dummy.next
	i := 0
	for elem != l.dummy {
		if i == idx {
			return elem
		}
		i++
		elem = elem.next
	}
	return nil
}
