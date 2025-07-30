package smartrequeue

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Helper function to get requeue duration from Result
func getRequeueAfter(res ctrl.Result, _ error) time.Duration {
	return res.RequeueAfter.Round(time.Second)
}

func TestEntry_Stable(t *testing.T) {
	// Setup
	store := NewStore(time.Second, time.Minute, 2)
	entry := newEntry(store)

	// Test the exponential backoff behavior
	t.Run("exponential backoff sequence", func(t *testing.T) {
		expectedDurations := []time.Duration{
			1 * time.Second,
			2 * time.Second,
			4 * time.Second,
			8 * time.Second,
			16 * time.Second,
			32 * time.Second,
			60 * time.Second, // Capped at maxInterval
			60 * time.Second, // Still capped
		}

		for i, expected := range expectedDurations {
			result, err := entry.Backoff()
			require.NoError(t, err)
			assert.Equal(t, expected, getRequeueAfter(result, err), "Iteration %d should have correct duration", i)
		}
	})
}

func TestEntry_Progressing(t *testing.T) {
	// Setup
	minInterval := time.Second
	maxInterval := time.Minute
	store := NewStore(minInterval, maxInterval, 2)
	entry := newEntry(store)

	// Ensure state is not at minimum
	_, _ = entry.Backoff()
	_, _ = entry.Backoff()

	// Test progressing resets duration to minimum
	t.Run("resets to minimum interval", func(t *testing.T) {
		result, err := entry.Reset()
		require.NoError(t, err)
		assert.Equal(t, minInterval, getRequeueAfter(result, err))

		// Second call should also return min interval with small increment
		result, err = entry.Reset()
		require.NoError(t, err)
		assert.Equal(t, minInterval, getRequeueAfter(result, err))
	})

	// After progressing, Stable should restart exponential backoff
	t.Run("stable continues from minimum", func(t *testing.T) {
		result, err := entry.Backoff()
		require.NoError(t, err)
		assert.Equal(t, 2*time.Second, getRequeueAfter(result, err))

		result, err = entry.Backoff()
		require.NoError(t, err)
		assert.Equal(t, 4*time.Second, getRequeueAfter(result, err))
	})
}

func TestEntry_Error(t *testing.T) {
	// Setup
	store := NewStore(time.Second, time.Minute, 2)
	entry := newEntry(store)
	testErr := errors.New("test error")

	// Ensure state is not at minimum
	_, _ = entry.Backoff()
	_, _ = entry.Backoff()

	// Test error handling
	t.Run("returns error and resets backoff", func(t *testing.T) {
		result, err := entry.Error(testErr)
		assert.Equal(t, testErr, err, "Should return the passed error")
		assert.Equal(t, 0*time.Second, getRequeueAfter(result, err), "Should have zero requeue time")
	})

	// After error, stable should continue from minimum
	t.Run("stable continues properly after error", func(t *testing.T) {
		result, err := entry.Backoff()
		require.NoError(t, err)
		assert.Equal(t, time.Second, getRequeueAfter(result, err))

		result, err = entry.Backoff()
		require.NoError(t, err)
		assert.Equal(t, 2*time.Second, getRequeueAfter(result, err))
	})
}

func TestEntry_Never(t *testing.T) {
	// Setup
	store := NewStore(time.Second, time.Minute, 2)
	entry := newEntry(store)

	// Test Never behavior
	t.Run("returns empty result", func(t *testing.T) {
		result, err := entry.Never()
		require.NoError(t, err)
		assert.Equal(t, time.Duration(0), getRequeueAfter(result, err))
	})
}
