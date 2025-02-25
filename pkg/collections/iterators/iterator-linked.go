package iterators

var _ Iterator[any] = &LinkedIterator[any]{}

type LinkedIterator[T any] struct {
	current LinkedIteratorElement[T]
	isValid func(LinkedIteratorElement[T]) bool
}

type LinkedIteratorElement[T any] interface {
	// Next returns the next element.
	Next() LinkedIteratorElement[T]
	// Value returns the value of the element.
	Value() T
}

// NewLinkedIterator returns a new LinkedIterator.
func NewLinkedIterator[T any](start LinkedIteratorElement[T], isValidFunc func(LinkedIteratorElement[T]) bool) *LinkedIterator[T] {
	return &LinkedIterator[T]{
		current: start,
		isValid: isValidFunc,
	}
}

func (i *LinkedIterator[T]) Next() T {
	res := i.current.Value()
	i.current = i.current.Next()
	return res
}

func (i *LinkedIterator[T]) HasNext() bool {
	return i.isValid(i.current)
}
