package pairs

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
)

// New creates a new Pair with the given key and value.
func New[K comparable, V any](key K, value V) Pair[K, V] {
	return Pair[K, V]{Key: key, Value: value}
}

// Pair is a generic key-value pair.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type Comparable[T any] interface {
	// CompareTo returns -1 if receiver < b, 0 if receiver == b, and 1 if receiver > b.
	CompareTo(other T) int
}

// Compare returns -1 if p.Key < other.Key, 0 if p.Key == other.Key, and 1 if p.Key > other.Key.
// It will panic if K does not implement the Comparable interface and cannot be parsed to an int64, float64, or string.
func (p Pair[K, V]) CompareTo(other Pair[K, V]) int {
	// check if K implements Comparable
	if a, ok := any(p.Key).(Comparable[K]); ok {
		return a.CompareTo(other.Key)
	}
	// check if *K implements Comparable
	if a, ok := any(&p.Key).(Comparable[*K]); ok {
		return a.CompareTo(&other.Key)
	}
	pKey := reflect.ValueOf(p.Key)
	otherKey := reflect.ValueOf(other.Key)
	// check if K is a number
	if pKey.CanInt() && otherKey.CanInt() {
		return cmp.Compare(pKey.Int(), otherKey.Int())
	}
	if pKey.CanFloat() && otherKey.CanFloat() {
		return cmp.Compare(pKey.Float(), otherKey.Float())
	}
	// check if K is a string
	stringType := reflect.TypeOf("")
	if pKey.CanConvert(stringType) && otherKey.CanConvert(stringType) {
		a := pKey.Convert(stringType).Interface().(string)
		b := otherKey.Convert(stringType).Interface().(string)
		return cmp.Compare(a, b)
	}
	panic(fmt.Sprintf("cannot compare type %T, must either implement Comparable or be a number or string", p.Key))
}

// Sort sorts a list of Pairs alphabetically by their keys.
// The sort happens in-place.
// Note that this uses the Compare method, which may panic if the keys are not comparable.
func Sort[K comparable, V any](pairs []Pair[K, V]) {
	slices.SortFunc(pairs, func(a, b Pair[K, V]) int {
		return a.CompareTo(b)
	})
}

// SortStable sorts a list of Pairs alphabetically by their keys.
// The sort happens in-place and is stable.
// Note that this uses the Compare method, which may panic if the keys are not comparable.
func SortStable[K comparable, V any](pairs []Pair[K, V]) {
	slices.SortStableFunc(pairs, func(a, b Pair[K, V]) int {
		return a.CompareTo(b)
	})
}

// MapToPairs converts a map[K]V to a []Pair[K, V].
// Note that the order of of the list is arbitrary.
func MapToPairs[K comparable, V any](pairs map[K]V) []Pair[K, V] {
	res := make([]Pair[K, V], 0, len(pairs))
	for k, v := range pairs {
		res = append(res, New(k, v))
	}
	return res
}

// PairsToMap converts a []Pair[K, V] to a map[K]V.
// In case of duplicate keys, the last value will be used.
func PairsToMap[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	res := make(map[K]V, len(pairs))
	for _, p := range pairs {
		res[p.Key] = p.Value
	}
	return res
}

func (p Pair[K, V]) String() string {
	return fmt.Sprintf("%v: %v", p.Key, p.Value)
}
