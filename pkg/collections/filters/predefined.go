package filters

import "reflect"

// And returns a filter that returns true if all given filters return true.
func And(filters ...Filter) Filter {
	return func(args ...any) bool {
		for _, filter := range filters {
			if !filter(args...) {
				return false
			}
		}
		return true
	}
}

// Or returns a filter that returns true if at least one of the given filters returns true.
func Or(filters ...Filter) Filter {
	return func(args ...any) bool {
		for _, filter := range filters {
			if filter(args...) {
				return true
			}
		}
		return false
	}
}

// Not returns the negation of the given filter.
func Not(filter Filter) Filter {
	return func(args ...any) bool {
		return !filter(args...)
	}
}

// True returns a filter that always returns true.
func True(args ...any) bool {
	return true
}

// False returns a filter that always returns false.
func False(args ...any) bool {
	return false
}

// ApplyToNthArgument applies feeds only the nth argument of the given arguments to the given filter.
// Indexing starts with 0.
func ApplyToNthArgument(n int, filter Filter) Filter {
	return func(args ...any) bool {
		return filter(args[n])
	}
}

// Wrap takes a function that returns a bool and turns it into a filter.
// If staticArgs is not nil or empty, the provided values are passed to the function at the given positions when the filter is called.
// This function panics in several cases:
// - the given value is not a function
// - the given function does not return a bool as first return value (further return values are ignored)
// - the args when calling the filter do not match the function's signature
// - the indices of the staticArgs are out of bounds for the function's signature
func Wrap(f any, staticArgs map[int]any) Filter {
	return func(args ...any) bool {
		vsLen := len(args) + len(staticArgs)
		vs := make([]reflect.Value, vsLen)
		j := 0
		for i := range vsLen {
			if sarg, ok := staticArgs[i]; ok {
				vs[i] = reflect.ValueOf(sarg)
			} else {
				vs[i] = reflect.ValueOf(args[j])
				j++
			}
		}
		return reflect.ValueOf(f).Call(vs)[0].Bool()
	}
}
