package collections

type Queue[T any] interface {
	Collection[T]

	// Push adds the given elements to the queue.
	// Returns an error, if the queue's size restriction prevents the elements from being added.
	Push(elements ...T) error
	// Peek returns the first element without removing it.
	// Returns T's zero value, if the queue is empty.
	Peek() T
	// Element returns the first element without removing it.
	// Returns an error, if the queue is empty.
	Element() (T, error)
	// Poll returns the first element and removes it from the queue.
	// Returns T's zero value, if the queue is empty.
	Poll() T
	// Fetch returns the first element and removes it from the queue.
	// Returns an error, if the queue is empty.
	Fetch() (T, error)
}
