package errors

import "fmt"

type CollectionEmptyError struct{}

func (CollectionEmptyError) Error() string {
	return "collection is empty"
}

func NewCollectionEmptyError() CollectionEmptyError {
	return CollectionEmptyError{}
}

type CollectionFullError struct {
	Capacity int
}

func (e CollectionFullError) Error() string {
	return fmt.Sprintf("collection is full (capacity: %d)", e.Capacity)
}

func NewCollectionFullError(cap int) CollectionFullError {
	return CollectionFullError{
		Capacity: cap,
	}
}

type IndexOutOfBoundsError struct {
	Index int
}

func (err IndexOutOfBoundsError) Error() string {
	return fmt.Sprintf("index out of bounds: %d", err.Index)
}

func NewIndexOutOfBoundsError(i int) IndexOutOfBoundsError {
	return IndexOutOfBoundsError{
		Index: i,
	}
}
