// Package engine is a frequency-first scheduler: ticks are driven by an elapsed
// frequency plus a deterministic offset rather than fixed calendar points.
package engine

import (
	"errors"
	"fmt"
	"time"
)

// ErrNoTick is returned by Next when no tick exists within the horizon.
var ErrNoTick = errors.New("engine: no tick within the horizon")

// Schedule computes when a task should next run.
type Schedule interface {
	Next(last time.Time) (time.Time, error)
}

// Option configures a scheduler.
type Option func(*scheduler) error

// Chain is a chainable list of options.
type Chain []Option

// Options starts an empty option chain.
func Options() Chain {
	return Chain(nil)
}

type scheduler struct {
	hz     time.Duration
	offset time.Duration

	// pin is the wall clock time of day (within a period) that ticks are pinned
	// to, -1 means no pin. a pin overrides the offset and anchors the tick grid
	// at local midnight instead of the unix epoch.
	pin time.Duration

	// horizon bounds how far past the reference time Next searches.
	horizon time.Duration
	// lastdelta shifts the reference time before searching, zero is strictly after
	// and Inclusive sets it to one nanosecond back.
	lastdelta time.Duration

	// weekdays is a bitmask indexed by time.Weekday, zero means unrestricted.
	weekdays uint8
	// dayorder is the order the weekdays were configured in, wrapped ranges
	// keep their order (fri-mon is fri, sat, sun, mon).
	dayorder []time.Weekday
	// monthdays is a bitmask where bit n is day of month n, zero means unrestricted.
	monthdays uint32

	window     bool
	start, end time.Duration

	loc *time.Location
}

// New builds a schedule that ticks every hz, shifted by offset.
func New(hz, offset time.Duration, options ...Option) (Schedule, error) {
	if hz <= 0 {
		return nil, fmt.Errorf("engine: frequency must be positive: %s", hz)
	}

	if offset < 0 {
		return nil, fmt.Errorf("engine: offset must not be negative: %s", offset)
	}

	s := &scheduler{hz: hz, offset: offset, pin: -1, loc: time.UTC, horizon: DefaultHorizon}
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	// a pin is applied last so neither a seed nor a fallback can move it.
	if s.pin >= 0 {
		s.offset = s.pin
	}

	s.offset %= hz
	return s, nil
}

// Spec is a read only description of a schedule's configuration.
type Spec struct {
	// Hz is the tick frequency.
	Hz time.Duration
	// Offset is the tick offset within one period, the pin when there is one.
	Offset time.Duration
	// Pin is the wall clock time ticks are pinned to, -1 when there is none.
	Pin time.Duration
	// Days are the permitted weekdays in configured order, empty when unrestricted.
	Days []time.Weekday
	// DaysOfMonth are the permitted days of the month in ascending order, empty when unrestricted.
	DaysOfMonth []int
	// Window is true when Start and End restrict the time of day.
	Window bool
	// Start and End are offsets from midnight. End before Start wraps past midnight.
	Start, End time.Duration
	// Horizon is how far past the reference time Next searches.
	Horizon time.Duration
	// LastDelta shifts the reference time before Next searches, zero is strictly after.
	LastDelta time.Duration
	// Location is the timezone the conditions are evaluated in, UTC unless configured.
	Location *time.Location
}

// Spec describes the scheduler's configuration.
func (t *scheduler) Spec() Spec {
	s := Spec{
		Hz:        t.hz,
		Offset:    t.offset,
		Pin:       t.pin,
		LastDelta: t.lastdelta,
		Horizon:   t.horizon,
		Window:    t.window,
		Start:     t.start,
		End:       t.end,
		Location:  t.loc,
	}

	s.Days = append(s.Days, t.dayorder...)

	for d := 1; d <= 31; d++ {
		if t.monthdays&(1<<d) != 0 {
			s.DaysOfMonth = append(s.DaysOfMonth, d)
		}
	}

	return s
}
