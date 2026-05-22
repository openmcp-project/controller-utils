package errors

import (
	"errors"
)

// ErrInvalidUserInput indicates invalid user input.
var ErrInvalidUserInput = errors.New("invalid user input")

// IgnoreUserInputError returns nil on invalid user input.
func IgnoreInvalidUserInput(err error) error {
	if errors.Is(err, ErrInvalidUserInput) {
		return nil
	}
	return err
}
