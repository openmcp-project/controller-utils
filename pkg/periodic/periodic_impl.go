package periodic

import (
	"context"
	"time"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

// periodic is the lock-free implementation holding all state.
// It must only be accessed while the owner (periodicSynced) holds its write lock,
// or via a periodicHandle passed into the run logic function.
type periodic[Config any] struct {
	name         string
	interval     time.Duration
	trigger      chan bool
	ticker       *time.Ticker
	lastRun      time.Time
	runLogicFunc RunLogicFunc[Config]
	log          logging.Logger
	configFuncs  []func(ctx context.Context, cfg Config) (Config, error)
	disabled     bool
}

var _ Periodic[any] = &periodic[any]{}

func (p *periodic[Config]) Name() string {
	return p.name
}

func (p *periodic[Config]) Enable() {
	if p.disabled {
		p.disabled = false
		p.log.Info("Enabling periodic", "nextRunAt", p.nextRunTime())
	}
	if p.ticker == nil {
		p.ticker = time.NewTicker(p.interval)
	}
}

func (p *periodic[Config]) Disable() {
	if !p.disabled {
		p.disabled = true
		p.log.Info("Disabling periodic")
	}
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}
}

func (p *periodic[Config]) IsDisabled() bool {
	return p.disabled
}

func (p *periodic[Config]) GetInterval() time.Duration {
	return p.interval
}

func (p *periodic[Config]) SetInterval(interval time.Duration) {
	p.interval = interval
}

func (p *periodic[Config]) Trigger(force bool) {
	p.GetTrigger() <- force
}

func (p *periodic[Config]) GetTrigger() chan<- bool {
	return p.trigger
}

func (p *periodic[Config]) LastRunTime() time.Time {
	return p.lastRun
}

func (p *periodic[Config]) NextRunTime() time.Time {
	return p.nextRunTime()
}

func (p *periodic[Config]) nextRunTime() time.Time {
	return p.lastRun.Add(p.interval)
}

func (p *periodic[Config]) Log() logging.Logger {
	return p.log
}

func (p *periodic[Config]) Start(_ context.Context) error {
	panic("periodic: Start must be called on an instance of periodicSynced, not periodic")
}
