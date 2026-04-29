package smartrequeue

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Entry is used to manage the requeue logic for a specific object.
// It holds the next duration to requeue and the store it belongs to.
type Entry struct {
	store        *Store
	nextDuration time.Duration
}

func newEntry(s *Store) *Entry {
	return &Entry{
		store:        s,
		nextDuration: s.minInterval,
	}
}

// ReturnError resets the backoff to minInterval and returns the given error,
// delegating backoff handling to controller-runtime.
func (e *Entry) ReturnError(err error) (ctrl.Result, error) {
	e.nextDuration = e.store.minInterval
	return ctrl.Result{}, err
}

// IsStable indicates the resource has reached its desired state. It requeues after
// the current interval and increases the interval for the next call, implementing
// exponential backoff.
func (e *Entry) IsStable() (ctrl.Result, error) {
	// Save current duration for result
	current := e.nextDuration

	// Schedule calculation of next duration
	defer e.setNext()

	return ctrl.Result{RequeueAfter: current}, nil
}

// RequeueWithBackoff requeues after the current interval and increases the interval for the next call.
// It is an alias for IsStable.
func (e *Entry) RequeueWithBackoff() (ctrl.Result, error) {
	return e.IsStable()
}

// IsProgressing indicates the resource is actively changing. It resets the backoff
// to minInterval and requeues after that interval.
func (e *Entry) IsProgressing() (ctrl.Result, error) {
	e.nextDuration = e.store.minInterval
	defer e.setNext()
	return ctrl.Result{RequeueAfter: e.nextDuration}, nil
}

// RequeueWithReset requeues after the minimum interval and resets the backoff for the next call.
// It is an alias for IsProgressing.
func (e *Entry) RequeueWithReset() (ctrl.Result, error) {
	return e.IsProgressing()
}

// StopRequeue removes the entry from the store and returns an empty result,
// stopping further requeues for this object.
func (e *Entry) StopRequeue() (ctrl.Result, error) {
	e.store.deleteEntry(e)
	return ctrl.Result{}, nil
}

// setNext updates the next requeue duration using exponential backoff.
// It multiplies the current duration by the store's multiplier and ensures
// the result doesn't exceed the configured maximum interval.
func (e *Entry) setNext() {
	newDuration := time.Duration(float32(e.nextDuration) * e.store.multiplier)

	if newDuration > e.store.maxInterval {
		newDuration = e.store.maxInterval
	}

	e.nextDuration = newDuration
}
