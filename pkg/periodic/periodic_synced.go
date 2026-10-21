package periodic

import (
	"context"
	"sync"
	"time"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

// periodicSynced wraps a periodic with a RWMutex.
// All public interface methods acquire the appropriate lock before delegating to the inner periodic.
// runLogic holds the write lock for its entire duration, so external callers block while the logic func runs.
type periodicSynced[Config any] struct {
	lock sync.RWMutex
	p    *periodic[Config]
}

var _ Periodic[any] = &periodicSynced[any]{}

func (s *periodicSynced[Config]) Name() string {
	// name is immutable after construction
	return s.p.Name()
}

func (s *periodicSynced[Config]) Enable() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.p.Enable()
}

func (s *periodicSynced[Config]) Disable() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.p.Disable()
}

func (s *periodicSynced[Config]) IsDisabled() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.p.IsDisabled()
}

func (s *periodicSynced[Config]) GetInterval() time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.p.GetInterval()
}

func (s *periodicSynced[Config]) SetInterval(interval time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.p.SetInterval(interval)
	s.p.Trigger(false)
}

func (s *periodicSynced[Config]) Trigger(force bool) {
	s.p.GetTrigger() <- force
}

func (s *periodicSynced[Config]) GetTrigger() chan<- bool {
	return s.p.GetTrigger()
}

func (s *periodicSynced[Config]) LastRunTime() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.p.LastRunTime()
}

func (s *periodicSynced[Config]) NextRunTime() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.p.NextRunTime()
}

func (s *periodicSynced[Config]) Log() logging.Logger {
	// logger is immutable after construction
	return s.p.Log()
}

func (s *periodicSynced[Config]) setNextRun(maybeInterval ...time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.p.disabled {
		s.p.log.Debug("Not scheduling next run because Periodic is disabled")
		return
	}
	interval := s.p.interval
	if len(maybeInterval) > 0 {
		interval = maybeInterval[0]
	}
	s.p.ticker.Reset(interval)
	s.p.log.Debug("Scheduling next run", "time", time.Now().Add(interval).Format(time.RFC3339))
}

func (s *periodicSynced[Config]) runLogic(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.p.log.Info("Starting run")

	var cfg Config
	for i, cfn := range s.p.configFuncs {
		s.p.log.Debug("Fetching config", "index", i)
		var err error
		cfg, err = cfn(ctx, cfg)
		if err != nil {
			s.p.log.Error(err, "Error fetching config, aborting run ...", "index", i)
			return
		}
		s.p.log.Debug("Successfully fetched config", "index", i)
	}

	handle := &periodicHandle[Config]{p: s.p}
	s.p.runLogicFunc(ctx, handle, cfg)
	handle.p = nil

	s.p.lastRun = time.Now()
	s.p.log.Info("Finished run")
}

// Start implements manager.Runnable.
func (s *periodicSynced[Config]) Start(ctx context.Context) error {
	s.p.log.Info("Starting periodic runnable")

	defer s.p.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.p.log.Info("Stopping periodic runnable")
			return nil
		case <-s.p.ticker.C:
			s.runLogic(ctx)
			s.setNextRun()
		case force := <-s.p.trigger:
			nextRunTime := s.NextRunTime()
			if force || !nextRunTime.After(time.Now()) {
				// if the trigger forces a run or the scheduled time lies in the past (or is now), run immediately
				// due to the zero time being very long ago, this will always trigger a run on the first trigger
				s.runLogic(ctx)
				s.setNextRun()
			} else {
				// compute the duration until the next run and reset the ticker accordingly
				s.setNextRun(time.Until(nextRunTime))
			}
		}
	}
}
