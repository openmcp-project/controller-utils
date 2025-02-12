package collections

import (
	"encoding/json"

	cerr "github.tools.sap/CoLa/controller-utils/pkg/collections/errors"
)

var _ List[any] = &abstractList[any]{}

type abstractList[T any] struct {
	abstractCollection[T]
}

func (al *abstractList[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(al.ToSlice())
}

func (al *abstractList[T]) UnmarshalJSON(data []byte) error {
	raw := &[]T{}
	err := json.Unmarshal(data, raw)
	if err != nil {
		return err
	}
	al.Clear()
	al.Add((*raw)...)
	return nil
}

func (al *abstractList[T]) AddIndex(element T, idx int) error {
	panic(ErrNotImplemented)
}

func (al *abstractList[T]) RemoveIndex(idx int) error {
	panic(ErrNotImplemented)
}

func (al *abstractList[T]) Get(idx int) (T, error) {
	var res T
	if idx < 0 || idx >= al.Size() {
		return res, cerr.NewIndexOutOfBoundsError(idx)
	}
	it := al.Iterator()
	for range idx {
		it.Next()
	}
	return it.Next(), nil
}
