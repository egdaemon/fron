package ast

import (
	"time"

	"github.com/egdaemon/fron/engine"
)

// Compile translates the tree into a schedule. Any extra options are applied
// after the options derived from the tree. seed is the fallback used when the
// schedule has no seed of its own: a stable offset for the same seed, a random
// offset for an empty one.
func Compile(t T, seed string, options ...engine.Option) (engine.Schedule, error) {
	opts := engine.Options()

	switch t.day {
	case dayWeekdays:
		opts = opts.Weekdays()
	case dayWeekends:
		opts = opts.Weekends()
	case dayEveryday:
		opts = opts.Everyday()
	case dayList:
		opts = opts.Days(t.Days...)
	case dayRange:
		opts = opts.DaysBetween(t.From, t.To)
	case dayOfMonth:
		opts = opts.DaysOfMonth(t.DaysOfMonth...)
	}

	if t.HasWindow {
		opts = opts.During(t.Start, t.End)
	}

	// a blank timezone is UTC.
	opts = opts.InTzName(t.Tz)
	opts = append(opts, options...)

	var offset time.Duration
	switch {
	case t.ZeroOffset:
		// a full period is zero modulo hz, but unlike zero the fallback
		// below leaves it alone, unless a caller option ignores the seed.
		offset = t.Hz
	case t.HasSeed:
		offset = engine.Seed(t.Seed)
	}

	// the fallback seed fills the offset when the schedule did not carry one,
	// or an option cleared it. it is applied last so caller options win.
	return engine.New(t.Hz, offset, opts.AutoSeed(seed)...)
}
