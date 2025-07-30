package maps

import (
	"reflect"

	"k8s.io/utils/ptr"

	"github.com/openmcp-project/controller-utils/pkg/collections/filters"
	"github.com/openmcp-project/controller-utils/pkg/pairs"
)

// Filter filters a map by applying a filter function to each key-value pair.
// Only the entries for which the filter function returns true are kept in the copy.
// The original map is not modified.
// Passes the key as first and the value as second argument into the filter function.
// Note that the values are not deep-copied.
// This is a convience alias for filters.FilterMap.
func Filter[K comparable, V any](m map[K]V, fil filters.Filter) map[K]V {
	return filters.FilterMap(m, fil)
}

// Merge merges multiple maps into a single one.
// The original maps are not modified.
// Note that the values are not deep-copied.
// If multiple maps contain the same key, the value from the last map in the list is used.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	res := map[K]V{}

	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}

	return res
}

// Intersect takes one source and any number of comparison maps and returns a new map containing only
// the entries of source for which keys exist in all comparison maps.
// The original maps are not modified.
// Note that the values are not deep-copied.
func Intersect[K comparable, V any](source map[K]V, maps ...map[K]V) map[K]V {
	res := map[K]V{}

	for k, v := range source {
		exists := true
		for _, m := range maps {
			if _, ok := m[k]; !ok {
				exists = false
				break
			}
		}
		if exists {
			res[k] = v
		}
	}

	return res
}

// ContainsKeysWithValues checks if 'super' is a superset of 'sub', meaning that all keys of 'sub' are also present in 'super' with the same values.
// Uses reflect.DeepEqual to compare the values, use ContainsMapFunc if you want to use a custom equality function.
func ContainsKeysWithValues[K comparable, V any](super, sub map[K]V) bool {
	return ContainsKeysWithValuesFunc(super, sub, func(a, b V) bool {
		return reflect.DeepEqual(a, b)
	})
}

// ContainsKeysWithValuesFunc checks if 'super' is a superset of 'sub', meaning that all keys of 'sub' are also present in 'super' with the same values.
// The values are compared using the provided equality function.
// If the equality function returns false for any key-value pair, it returns false.
func ContainsKeysWithValuesFunc[K comparable, V any](super, sub map[K]V, equal func(a, b V) bool) bool {
	for k, bv := range sub {
		if av, ok := super[k]; !ok || !equal(av, bv) {
			return false
		}
	}
	return true
}

// ContainsKeys returns true if all given keys are present in the map.
func ContainsKeys[K comparable, V any](super map[K]V, keys ...K) bool {
	for _, k := range keys {
		if _, ok := super[k]; !ok {
			return false
		}
	}
	return true
}

// GetAny returns an arbitrary key-value pair from the map as a pointer to a pairs.Pair.
// If the map is empty, it returns nil.
func GetAny[K comparable, V any](m map[K]V) *pairs.Pair[K, V] {
	for k, v := range m {
		return ptr.To(pairs.New(k, v))
	}
	return nil
}
