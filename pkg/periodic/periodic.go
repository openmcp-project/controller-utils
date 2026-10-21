package periodic

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

// Periodic is the public interface for a periodic runnable.
// It can be registered with a controller-runtime manager and runs a configured logic function at a set interval.
// The mutating methods (Enable, Disable, SetInterval) are safe to call from within the run logic function.
type Periodic[Config any] interface {
	manager.Runnable
	// Name returns the name of the periodic. Only used for logging and identification purposes.
	Name() string
	// Enable re-enables the periodic if it was previously disabled.
	Enable()
	// Disable prevents the periodic from running until re-enabled.
	Disable()
	// IsDisabled returns whether the periodic is currently disabled.
	IsDisabled() bool
	// GetInterval returns the interval at which the logic is run.
	GetInterval() time.Duration
	// SetInterval sets the interval at which the logic is run.
	// Setting the interval automatically triggers a check whether the logic should run now.
	SetInterval(interval time.Duration)
	// Trigger is a convenience wrapper for 'GetTrigger() <- force'.
	// If force is true, a new run is triggered immediately.
	// Otherwise, the periodic checks the scheduled interval and runs if necessary.
	Trigger(force bool)
	// GetTrigger returns the channel that can be used to trigger runs.
	// Send false to check the interval and run if due; send true to force an immediate run.
	GetTrigger() chan<- bool
	// LastRunTime returns the time at which the periodic was last run.
	// Returns the zero time if the periodic has never run.
	LastRunTime() time.Time
	// NextRunTime returns the next scheduled run time.
	// Note that changes to the interval or manual triggers can cause the actual next run time to differ.
	NextRunTime() time.Time
	// Log returns a logger instance scoped to the periodic.
	Log() logging.Logger
}

// RunLogicFunc is the type of the function passed to NewPeriodicBuilder.
// It receives:
//   - ctx: the current context
//   - p: a handle to the periodic itself; mutating methods (SetInterval, Enable, Disable) are safe to call here
//   - cfg: the assembled configuration, or the zero value if no config funcs were set
//
// The handle p must not be stored and used after the function returns — it will panic.
type RunLogicFunc[Config any] func(ctx context.Context, p Periodic[Config], cfg Config)
