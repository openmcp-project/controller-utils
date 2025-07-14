package collections

// ProjectSlice takes a slice and a projection function and applies this function to each element of the slice.
// It returns a new slice containing the results of the projection.
// The original slice is not modified.
// If the projection function is nil, it returns nil.
func ProjectSlice[X any, Y any](src []X, project func(X) Y) []Y {
	if project == nil {
		return nil
	}
	res := make([]Y, len(src))
	for i, src := range src {
		res[i] = project(src)
	}
	return res
}

// ProjectMapToSlice takes a map and a projection function and applies this function to each key-value pair in the map.
// It returns a new slice containing the results of the projection.
// The original map is not modified.
// If the projection function is nil, it returns nil.
func ProjectMapToSlice[K comparable, V any, R any](src map[K]V, project func(K, V) R) []R {
	if project == nil {
		return nil
	}
	res := make([]R, 0, len(src))
	for k, v := range src {
		res = append(res, project(k, v))
	}
	return res
}

// ProjectMapToMap takes a map and a projection function and applies this function to each key-value pair in the map.
// It returns a new map containing the results of the projection.
// The original map is not modified.
// Note that the resulting map may be smaller if the projection function does not guarantee unique keys.
// If the projection function is nil, it returns nil.
func ProjectMapToMap[K1 comparable, V1 any, K2 comparable, V2 any](src map[K1]V1, project func(K1, V1) (K2, V2)) map[K2]V2 {
	if project == nil {
		return nil
	}
	res := make(map[K2]V2, len(src))
	for k, v := range src {
		newK, newV := project(k, v)
		res[newK] = newV
	}
	return res
}

// AggregateSlice takes a slice, an aggregation function and an initial value.
// It applies the aggregation function to each element of the slice, also passing in the current result.
// For the first element, it uses the initial value as the current result.
// Returns initial if the aggregation function is nil.
func AggregateSlice[X any, Y any](src []X, agg func(X, Y) Y, initial Y) Y {
	if agg == nil {
		return initial
	}
	res := initial
	for _, x := range src {
		res = agg(x, res)
	}
	return res
}

// AggregateMap takes a map, an aggregation function and an initial value.
// It applies the aggregation function to each key-value pair in the map, also passing in the current result.
// For the first key-value pair, it uses the initial value as the current result.
// Returns initial if the aggregation function is nil.
// Note that the iteration order over the map elements is undefined and may vary between executions.
func AggregateMap[K comparable, V any, R any](src map[K]V, agg func(K, V, R) R, initial R) R {
	if agg == nil {
		return initial
	}
	res := initial
	for k, v := range src {
		res = agg(k, v, res)
	}
	return res
}
