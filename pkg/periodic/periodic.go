package periodic

import (
	"context"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

type Periodic[Config any] struct {
	lock         sync.RWMutex
	name         string
	interval     time.Duration
	trigger      chan bool
	ticker       *time.Ticker
	lastRun      time.Time
	runLogicFunc func(ctx context.Context, p *PeriodicReadOnly[Config], cfg Config)
	log          logging.Logger
	configFuncs  []func(ctx context.Context, cfg Config) (Config, error)
	disabled     bool
}

var _ manager.Runnable = &Periodic[any]{}

// Name returns the name of the periodic.
func (p *Periodic[Config]) Name() string {
	// name is immutable, so no need to acquire a lock
	return p.name
}

// Enable is the sibling function to 'Disable' and re-enables the periodic if it was previously disabled.
func (p *Periodic[Config]) Enable() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.disabled {
		p.disabled = false
		p.log.Info("Enabling periodic", "nextRunAt", p.nextRunTime())
	}
	if p.ticker == nil {
		p.ticker = time.NewTicker(p.interval)
	}
}

// Disable disables the periodic, preventing it from running until it is re-enabled using 'Enable'.
func (p *Periodic[Config]) Disable() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.disabled {
		p.disabled = true
		p.log.Info("Disabling periodic")
	}
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}
}

// IsDisabled returns whether the periodic is currently disabled.
func (p *Periodic[Config]) IsDisabled() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.disabled
}

// GetInterval returns the interval at which the logic is run.
func (p *Periodic[Config]) GetInterval() time.Duration {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.interval
}

// SetInterval sets the interval at which the logic is run.
// Setting the interval automatically triggers a check whether the logic should run now.
func (p *Periodic[Config]) SetInterval(interval time.Duration) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.interval = interval
	p.Trigger(false)
}

// Trigger is a convenience wrapper for 'p.GetTrigger() <- force'.
// If force is true, a new run is triggered immediately.
// Otherwise, the periodic just checks for the scheduled interval and runs the logic if necessary.
// This is called by SetInterval automatically to notify the periodic of the changed interval.
func (p *Periodic[Config]) Trigger(force bool) {
	p.GetTrigger() <- force
}

// GetTrigger returns the channel that can be used to trigger runs.
// The boolean pushed into the channel is interpreted as 'force' flag:
// - false: check if logic should be executed based on the configured interval and potentially run it
// - true: force a run regardless of the configured interval
func (p *Periodic[Config]) GetTrigger() chan<- bool {
	return p.trigger
}

// LastRunTime returns the time at which the periodic was last run.
// This returns the zero time if the periodic has never run.
func (p *Periodic[Config]) LastRunTime() time.Time {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.lastRun
}

// NextRunTime returns the next scheduled run time.
// Note that changes to the interval or manual triggers can cause the actual next run time to differ from this value.
func (p *Periodic[Config]) NextRunTime() time.Time {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.nextRunTime()
}

// nextRunTime is the internal variant of NextRunTime that does not acquire the read lock.
func (p *Periodic[Config]) nextRunTime() time.Time {
	return p.lastRun.Add(p.interval)
}

// Log returns a logger instance scoped to the periodic.
func (p *Periodic[Config]) Log() logging.Logger {
	// logger is immutable, so no need to acquire a lock
	return p.log
}

func (p *Periodic[Config]) setNextRun(maybeInterval ...time.Duration) {
	p.lock.Lock()
	defer p.lock.Unlock()
	interval := p.interval
	if len(maybeInterval) > 0 {
		interval = maybeInterval[0]
	}
	p.ticker.Reset(interval)
	p.log.Debug("Scheduling next run", "time", time.Now().Add(interval).Format(time.RFC3339))
}

func (p *Periodic[Config]) readOnly() *PeriodicReadOnly[Config] {
	return &PeriodicReadOnly[Config]{periodic: p}
}

func (p *Periodic[Config]) runLogic(ctx context.Context) {
	p.log.Info("Starting run")

	var cfg Config
	for i, cfn := range p.configFuncs {
		p.log.Debug("Fetching config", "index", i)
		var err error
		cfg, err = cfn(ctx, cfg)
		if err != nil {
			p.log.Error(err, "Error fetching config, aborting run ...", "index", i)
			return
		}
		p.log.Debug("Successfully fetched config", "index", i)
	}

	p.lock.RLock()
	p.runLogicFunc(ctx, p.readOnly(), cfg)
	p.lock.RUnlock()

	p.lock.Lock()
	p.lastRun = time.Now()
	p.lock.Unlock()

	p.log.Info("Finished run")
}

// Start makes the Periodic implement the Runnable interface, so it can be started by the controller-runtime manager.
// This method starts the loop, listening for ticker events and manual triggers.
// It is blocking and will run until the context is canceled.
func (p *Periodic[Config]) Start(ctx context.Context) error {
	p.log.Info("Starting periodic runnable")

	defer p.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.log.Info("Stopping periodic runnable")
			return nil
		case <-p.ticker.C:
			p.runLogic(ctx)
			p.setNextRun()
		case force := <-p.trigger:
			// due to the zero time being very long ago, this will always trigger a run on the first trigger
			nextRunTime := p.NextRunTime()
			if force || !nextRunTime.After(time.Now()) {
				// if the trigger forces a run or the scheduled time for the next run lies in the past (or is now), run immediately
				p.runLogic(ctx)
				p.setNextRun()
			} else {
				// compute the duration until the next run and reset the ticker accordingly
				p.setNextRun(time.Until(nextRunTime))
			}
		}
	}
}

// PeriodicReadOnly provides a read-only view of the Periodic, exposing only non-mutating methods.
type PeriodicReadOnly[Config any] struct {
	periodic *Periodic[Config]
}

// Name returns the name of the periodic.
// Only used for logging purposes. May be empty.
func (p *PeriodicReadOnly[Config]) Name() string {
	return p.periodic.Name()
}

// GetInterval returns the interval at which the periodic logic is run.
func (p *PeriodicReadOnly[Config]) GetInterval() time.Duration {
	return p.periodic.GetInterval()
}

// LastRunTime returns the time at which the periodic was last run.
// This returns the zero time if the periodic has never run.
func (p *PeriodicReadOnly[Config]) LastRunTime() time.Time {
	return p.periodic.LastRunTime()
}

// NextRunTime returns the next scheduled run time.
// Note that changes to the interval or manual triggers can cause the actual next run time to differ from this value.
func (p *PeriodicReadOnly[Config]) NextRunTime() time.Time {
	return p.periodic.NextRunTime()
}

// Log returns a logger instance scoped to the periodic.
func (p *PeriodicReadOnly[Config]) Log() logging.Logger {
	return p.periodic.Log()
}
