package smartrequeue

import (
	"reflect"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Store is used to manage requeue entries for different objects.
// It holds a map of entries indexed by a key that uniquely identifies the object.
type Store struct {
	minInterval time.Duration
	maxInterval time.Duration
	multiplier  float32
	objects     map[key]*Entry
	mu          sync.RWMutex // Using RWMutex for better read concurrency
}

// NewStore creates a new Store with the specified minimum and maximum intervals
// and a multiplier for the exponential backoff logic.
func NewStore(minInterval, maxInterval time.Duration, multiplier float32) *Store {
	if minInterval <= 0 {
		minInterval = 1 * time.Second // Safe default
	}

	if maxInterval < minInterval {
		maxInterval = minInterval * 60 // Safe default: 1 minute or 60x min
	}

	if multiplier <= 1.0 {
		multiplier = 2.0 // Safe default: double each time
	}

	return &Store{
		minInterval: minInterval,
		maxInterval: maxInterval,
		multiplier:  multiplier,
		objects:     make(map[key]*Entry),
	}
}

// For gets or creates an Entry for the specified object.
func (s *Store) For(obj client.Object) *Entry {
	key := keyFromObject(obj)

	// Try read lock first for better concurrency
	s.mu.RLock()
	entry, exists := s.objects[key]
	s.mu.RUnlock()

	if exists {
		return entry
	}

	// Need to create a new entry
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check again in case another goroutine created it while we were waiting
	entry, exists = s.objects[key]
	if !exists {
		entry = &Entry{
			store:        s,
			nextDuration: s.minInterval,
		}
		s.objects[key] = entry
	}

	return entry
}

// Clear removes all entries from the store (mainly useful for testing).
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.objects = make(map[key]*Entry)
}

// deleteEntry removes an entry from the store.
func (s *Store) deleteEntry(toDelete *Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, entry := range s.objects {
		if entry == toDelete {
			delete(s.objects, k)
			break
		}
	}
}

// keyFromObject generates a unique key for a client.Object.
func keyFromObject(obj client.Object) key {
	kind := ""
	if obj != nil {
		kind = obj.GetObjectKind().GroupVersionKind().Kind
		if kind == "" {
			// Fallback if Kind is not set in GroupVersionKind
			kind = reflect.TypeOf(obj).Elem().Name()
		}
	}

	return key{
		Kind:      kind,
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}

// key uniquely identifies a Kubernetes object.
type key struct {
	Kind      string
	Name      string
	Namespace string
}
