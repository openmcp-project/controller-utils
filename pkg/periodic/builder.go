package periodic

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

const (
	DefaultPeriodicInterval = 24 * time.Hour
)

type periodicBuilder[Config any] struct {
	interval         time.Duration
	name             string
	log              *logging.Logger
	loggerAppendName bool
	runLogicFunc     RunLogicFunc[Config]
	configFuncs      []func(ctx context.Context, cfg Config) (Config, error)
}

// NewPeriodicBuilder creates a builder for a new periodic instance.
// Use the builder methods to configure it, then call Build() to create the periodic.
func NewPeriodicBuilder[Config any](runLogicFunc RunLogicFunc[Config]) *periodicBuilder[Config] {
	return (&periodicBuilder[Config]{}).
		WithInterval(DefaultPeriodicInterval).
		WithRunLogicFunc(runLogicFunc)
}

// WithName sets the name of the periodic.
// This is mainly useful for logging and identification purposes when running more than one periodic in parallel.
// Can be left empty if not needed.
func (b *periodicBuilder[Config]) WithName(name string) *periodicBuilder[Config] {
	b.name = name
	return b
}

// WithInterval sets the interval at which the periodic runs.
// This can be changed during runtime using the SetInterval method on the periodic instance.
// Defaults to 24h if not specified.
func (b *periodicBuilder[Config]) WithInterval(interval time.Duration) *periodicBuilder[Config] {
	b.interval = interval
	return b
}

// WithLogger sets the logger for the periodic.
// Note that controller-runtime does not inject a logger for non-controller runnables, so the logger is not taken from the provided context.
// Set to logging.Discard() to disable logging.
// If appendNames is true, "Periodic" and the periodic's name, if not empty, will be appended to the logger.
// Is set to a default logger, if not specified.
func (b *periodicBuilder[Config]) WithLogger(log logging.Logger, appendNames bool) *periodicBuilder[Config] {
	b.log = &log
	b.loggerAppendName = appendNames
	return b
}

// WithRunLogicFunc sets the function that holds the actual periodic logic.
// It gets the following arguments:
//   - ctx: The current context.
//   - p: A handle to the periodic. Can be used to read or modify it, e.g. adjust the interval based on the loaded config.
//     Must not be stored and used after the function returns.
//   - cfg: If config getter functions have been configured, they are called to fetch the configuration at the beginning of the run. Otherwise, this is the zero value of Config.
func (b *periodicBuilder[Config]) WithRunLogicFunc(runLogicFunc RunLogicFunc[Config]) *periodicBuilder[Config] {
	b.runLogicFunc = runLogicFunc
	return b
}

// WithConfigFuncs sets the functions that will be called to fetch the configuration for the periodic.
// These functions are optional and can be omitted if no configuration needs to be fetched.
// Each function receives the current context and the current configuration as arguments.
// The first function receives the zero value of Config as the current configuration.
// The returned configuration overwrites the current configuration for the next function in the chain and ultimately for the run.
func (b *periodicBuilder[Config]) WithConfigFuncs(getConfigFuncs ...func(ctx context.Context, cfg Config) (Config, error)) *periodicBuilder[Config] {
	b.configFuncs = getConfigFuncs
	return b
}

// WithDefaultLogger sets the logger to the default logger.
// Overwrites any previously set logger.
func (b *periodicBuilder[Config]) WithDefaultLogger() *periodicBuilder[Config] {
	b.log = nil
	return b
}

func (b *periodicBuilder[Config]) Build() (Periodic[Config], error) {
	if b.runLogicFunc == nil {
		return nil, fmt.Errorf("runLogicFunc must not be nil")
	}
	if b.log == nil {
		log, err := logging.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("unable to construct default logger: %w", err)
		}
		b.log = &log
		b.loggerAppendName = true
	}
	if b.loggerAppendName && b.name != "" {
		*b.log = b.log.WithName(b.name)
	}
	p := &periodic[Config]{
		interval:     b.interval,
		name:         b.name,
		log:          *b.log,
		trigger:      make(chan bool, 10), // buffered to avoid blocking if multiple triggers arrive in quick succession
		ticker:       time.NewTicker(b.interval),
		configFuncs:  slices.Clone(b.configFuncs),
		runLogicFunc: b.runLogicFunc,
	}
	return &periodicSynced[Config]{p: p}, nil
}
