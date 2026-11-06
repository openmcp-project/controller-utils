package periodic

import (
	"context"
	"time"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

// periodicHandle is a single-use forwarding wrapper around *periodic passed into the run logic function.
// All calls are forwarded to the inner periodic (which is already protected by periodicSynced's write lock).
// After the run logic function returns, the inner pointer is set to nil.
// Any attempt to use the handle after that point will panic.
type periodicHandle[Config any] struct {
	p *periodic[Config]
}

var _ Periodic[any] = &periodicHandle[any]{}

func (h *periodicHandle[Config]) mustBeValid() {
	if h.p == nil {
		panic("periodic: handle used after run logic function returned — do not store the periodic handle")
	}
}

func (h *periodicHandle[Config]) Name() string {
	h.mustBeValid()
	return h.p.Name()
}

func (h *periodicHandle[Config]) Enable() {
	h.mustBeValid()
	h.p.Enable()
}

func (h *periodicHandle[Config]) Disable() {
	h.mustBeValid()
	h.p.Disable()
}

func (h *periodicHandle[Config]) IsDisabled() bool {
	h.mustBeValid()
	return h.p.IsDisabled()
}

func (h *periodicHandle[Config]) GetInterval() time.Duration {
	h.mustBeValid()
	return h.p.GetInterval()
}

func (h *periodicHandle[Config]) SetInterval(interval time.Duration) {
	h.mustBeValid()
	h.p.SetInterval(interval)
}

func (h *periodicHandle[Config]) Trigger(force bool) {
	h.mustBeValid()
	h.p.Trigger(force)
}

func (h *periodicHandle[Config]) GetTrigger() chan<- bool {
	h.mustBeValid()
	return h.p.GetTrigger()
}

func (h *periodicHandle[Config]) LastRunTime() time.Time {
	h.mustBeValid()
	return h.p.LastRunTime()
}

func (h *periodicHandle[Config]) NextRunTime() time.Time {
	h.mustBeValid()
	return h.p.NextRunTime()
}

func (h *periodicHandle[Config]) Log() logging.Logger {
	h.mustBeValid()
	return h.p.Log()
}

// Start panics — starting the periodic from within the run logic function is not supported.
func (h *periodicHandle[Config]) Start(_ context.Context) error {
	panic("periodic: Start must not be called from within the run logic function")
}
