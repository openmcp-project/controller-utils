package errors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/errors"
)

func TestIgnoreInvalidUserInput(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		err          error
		wrappedError error
		wantErr      bool
	}{
		{
			name:    "user input error is ignored",
			err:     fmt.Errorf("value out of range %w", ctrlutils.ErrInvalidUserInput),
			wantErr: false,
		},
		{
			name:    "regular error is returned",
			err:     errors.New("regular error"),
			wantErr: true,
		},
		{
			name:    "nil is nil",
			err:     nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := ctrlutils.IgnoreInvalidUserInput(tt.err)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("IgnoreInvalidUserInput() failed: %v", gotErr)
				}
				assert.Equal(t, tt.err, gotErr)
				if tt.wrappedError != nil {
					assert.ErrorIs(t, tt.err, tt.wrappedError)
				}
				return
			}
			assert.False(t, tt.wantErr)
		})
	}
}
