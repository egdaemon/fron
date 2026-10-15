package ast

import (
	"time"

	"github.com/egdaemon/fron/engine"
)

// Compile translates the tree into a schedule. Any extra options are applied
// after the options derived from the tree. A schedule that ends up with no
// offset of its own, because it carries no seed or an option cleared it, gets a
// random one; give the options an AutoSeed to make it stable instead.
func Compile(t T, options ...engine.Option) (engine.Schedule, error) {
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

	if t.Pin >= 0 {
		opts = opts.At(t.Pin)
	}

	// a blank timezone is UTC.
	opts = opts.InTzName(t.Tz)
	opts = append(opts, options...)

	var offset time.Duration
	if t.HasSeed {
		offset = engine.Seed(t.Seed)
	}

	// a random offset fills the gap when the schedule did not carry a seed, or
	// an option cleared it. it is applied last, so an AutoSeed given by the
	// caller has already filled the offset and wins. a pin overrides the offset
	// whatever it is.
	return engine.New(t.Hz, offset, opts.AutoSeed("")...)
}
