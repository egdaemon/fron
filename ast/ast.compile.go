package ast

import (
	"github.com/egdaemon/fron/engine"
)

// Compile translates the tree into a schedule. Any extra options are applied
// after the options derived from the tree.
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

	// a blank timezone is UTC.
	opts = opts.InTzName(t.Tz)
	opts = append(opts, options...)

	switch {
	case t.ZeroOffset:
		return engine.New(t.Hz, engine.ZeroOffset, opts...)
	case t.HasSeed:
		return engine.New(t.Hz, engine.Seed(t.Seed), opts...)
	default:
		return engine.New(t.Hz, engine.Seed(t.Input), opts...)
	}
}
