package smartrequeue

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestFor(t *testing.T) {
	tests := []struct {
		name        string
		firstObj    client.Object
		secondObj   client.Object
		expectSame  bool
		description string
	}{
		{
			name: "same object returns same entry",
			firstObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			secondObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			expectSame:  true,
			description: "Expected to get the same entry back",
		},
		{
			name: "different namespace returns different entry",
			firstObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			secondObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test2",
				},
			},
			expectSame:  false,
			description: "Expected to get a different entry back",
		},
		{
			name: "different name returns different entry",
			firstObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			secondObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test2",
					Namespace: "test",
				},
			},
			expectSame:  false,
			description: "Expected to get a different entry back",
		},
		{
			name: "different kind returns different entry",
			firstObj: &dummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			secondObj: &anotherDummyObject{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			expectSame:  false,
			description: "Expected to get a different entry back",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore(time.Second, time.Minute, 2)
			entry1 := store.For(tt.firstObj)

			assert.NotNil(t, entry1, "Expected entry to be created")
			result, err := entry1.Backoff()
			require.NoError(t, err)
			assert.Equal(t, 1*time.Second, getRequeueAfter(result, err))

			entry2 := store.For(tt.secondObj)

			if tt.expectSame {
				assert.Same(t, entry1, entry2, tt.description)
			} else {
				assert.NotSame(t, entry1, entry2, tt.description)
			}
		})
	}
}

// TestClear ensures the Clear method removes all entries
func TestClear(t *testing.T) {
	store := NewStore(time.Second, time.Minute, 2)

	// Add some entries
	obj1 := &dummyObject{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "test1",
			Namespace: "test",
		},
	}

	obj2 := &dummyObject{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "test2",
			Namespace: "test",
		},
	}

	// Get entries to populate the store
	entry1 := store.For(obj1)
	entry2 := store.For(obj2)

	// Verify entries exist
	assert.NotNil(t, entry1)
	assert.NotNil(t, entry2)

	// Clear the store
	store.Clear()

	// Get entries again - they should be new instances
	entry1After := store.For(obj1)
	entry2After := store.For(obj2)

	// Verify they're different instances
	assert.NotSame(t, entry1, entry1After)
	assert.NotSame(t, entry2, entry2After)
}

// TestConcurrentAccess tests that the store handles concurrent access properly
func TestConcurrentAccess(t *testing.T) {
	store := NewStore(time.Second, time.Minute, 2)

	// Create a series of objects
	const numObjects = 100
	objects := make([]client.Object, numObjects)

	for i := 0; i < numObjects; i++ {
		objects[i] = &dummyObject{
			ObjectMeta: ctrl.ObjectMeta{
				Name:      fmt.Sprintf("test-%d", i),
				Namespace: "test",
			},
		}
	}

	// Access concurrently
	var wg sync.WaitGroup
	wg.Add(numObjects)

	for i := 0; i < numObjects; i++ {
		go func(idx int) {
			defer wg.Done()
			obj := objects[idx]
			entry := store.For(obj)
			_, _ = entry.Backoff() // Just exercise the API
		}(i)
	}

	wg.Wait()
}
