package maps

import "github.tools.sap/CoLa/controller-utils/pkg/collections/filters"

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
