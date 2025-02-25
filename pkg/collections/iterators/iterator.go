package iterators

import "fmt"

var (
	ErrNoNextElement error = fmt.Errorf("Next() called on empty iterator")
)

type Iterable[T any] interface {
	// Iterator returns an Iterator over the contained elements.
	// Note that the behavior is undefined if the underlying collection is modified while being iterated over.
	Iterator() Iterator[T]
}

type Iterator[T any] interface {

	// Next returns the next element of the iterated collection.
	// Panics if there is no next element.
	Next() T

	// HasNext returns true if there is a next element.
	HasNext() bool
}
