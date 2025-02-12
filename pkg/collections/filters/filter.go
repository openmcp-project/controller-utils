package filters

import "github.com/openmcp-project/controller-utils/pkg/collections"

type Filter func(args ...any) bool

// FilterSlice filters a slice by applying a filter function to each entry.
// Only the entries for which the filter function returns true are kept the copy.
// The original slice is not modified.
// Note that the entries are not deep-copied.
func FilterSlice[T any](s []T, filter Filter) []T {
	res := make([]T, 0, len(s))
	for _, entry := range s {
		if filter(entry) {
			res = append(res, entry)
		}
	}
	return res
}

// FilterMap filters a map by applying a filter function to each key-value pair.
// Only the entries for which the filter function returns true are kept in the copy.
// The original map is not modified.
// Passes the key as first and the value as second argument into the filter function.
// Note that the values are not deep-copied.
func FilterMap[K comparable, V any](m map[K]V, filter Filter) map[K]V {
	res := make(map[K]V)
	for k, v := range m {
		if filter(k, v) {
			res[k] = v
		}
	}
	return res
}

// FilterCollection filters a collection by applying a filter function to each entry.
// Only the entries for which the filter function returns true are kept in the copy.
// The original collection is not modified.
// Note that the entries are not deep-copied.
func FilterCollection[T any](c collections.Collection[T], filter Filter) collections.Collection[T] {
	res := c.New()
	it := c.Iterator()
	for it.HasNext() {
		entry := it.Next()
		if filter(entry) {
			res.Add(entry)
		}
	}
	return res
}
