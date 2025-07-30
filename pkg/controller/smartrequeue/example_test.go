package smartrequeue_test

import (
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/controller/smartrequeue"
)

// This example shows how to use the SmartRequeue package in a Kubernetes controller.
func Example_controllerUsage() {
	// Create a store with min and max requeue intervals
	store := smartrequeue.NewStore(5*time.Second, 10*time.Minute, 2.0)

	// In your controller's Reconcile function:
	reconcileFunction := func(_ ctrl.Request) (ctrl.Result, error) {
		// Create a dummy object representing what you'd get from the client
		var obj client.Object // In real code: Get this from the client

		// Get the Entry for this specific object
		entry := store.For(obj)

		// Determine the state of the external resource...
		inProgress := false  // This would be determined by your logic
		errOccurred := false // This would be determined by your logic

		// nolint:gocritic
		if errOccurred {
			// Handle error case
			err := fmt.Errorf("something went wrong")
			return entry.Error(err)
		} else if inProgress {
			// Resource is changing - check back soon
			return entry.Reset()
		} else {
			// Resource is stable - gradually back off
			return entry.Backoff()
		}
	}

	// Call the reconcile function
	_, _ = reconcileFunction(ctrl.Request{})
}
