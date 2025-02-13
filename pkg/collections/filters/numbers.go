package filters

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// NumericallyGreaterThan returns a Filter that returns true for any value greater than n.
// The Filter panics if the value is not a number.
func NumericallyGreaterThan[T Number](n T) Filter {
	return func(args ...any) bool {
		x := args[0].(T)
		return x > n
	}
}

// NumericallyLessThan returns a Filter that returns true for any value less than n.
// The Filter panics if the value is not a number.
func NumericallyLessThan[T Number](n T) Filter {
	return func(args ...any) bool {
		x := args[0].(T)
		return n > x
	}
}

// NumericallyEqualTo returns a Filter that returns true for any value equal to n.
// The Filter panics if the value is not a number.
func NumericallyEqualTo[T Number](n T) Filter {
	return func(args ...any) bool {
		x := args[0].(T)
		return n == x
	}
}

// NumericallyGreaterThanOrEqualTo returns a Filter that returns true for any value greater or equal to n.
// The Filter panics if the value is not a number.
func NumericallyGreaterThanOrEqualTo[T Number](n T) Filter {
	return Or(NumericallyGreaterThan(n), NumericallyEqualTo(n))
}

// NumericallyLessThanOrEqualTo returns a Filter that returns true for any value less or equal to n.
// The Filter panics if the value is not a number.
func NumericallyLessThanOrEqualTo[T Number](n T) Filter {
	return Or(NumericallyLessThan(n), NumericallyEqualTo(n))
}
