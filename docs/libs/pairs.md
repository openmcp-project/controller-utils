# Key-Value Pairs

The `pkg/pairs` library contains mainly the `Pairs` type, which is a generic type representing a single key-value pair.
The `MapToPairs` and `PairsToMap` helper functions can convert between `map[K]V` and `[]Pair[K, V]`.
This is useful for example if key-value pairs are meant to be passed into a function as variadic arguments:
```go
func myFunc(labels ...Pair[string]string) {
  labelMap := PairsToMap(labels)
  <...>
}

func main() {
  myFunc(pairs.New("foo", "bar"), pairs.New("bar", "baz"))
}
```

The `Sort` and `SortStable` functions as well as the `Compare` method of `Pair` can be used to compare and sort pairs by their keys. Note that these functions will panic if the key cannot be converted into an `int64`, `float64`, `string`, or does implement the package's `Comparable` interface.
If the interface is implemented, its `Compare` implementation takes precedence over the conversion into one of the mentioned base types.
