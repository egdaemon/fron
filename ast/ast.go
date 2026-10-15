// Package ast parses the fron textual DSL into a tree, compiles that tree into
// a engine.Schedule, and renders it as a cron expression.
package ast

import "time"

type dayclause uint8

const (
	dayNone dayclause = iota
	dayWeekdays
	dayWeekends
	dayEveryday
	dayList
	dayRange
	dayOfMonth
)

// T is a node in the parsed schedule tree.
type T struct {
	Seed    string
	HasSeed bool

	Hz time.Duration

	day         dayclause
	Days        []time.Weekday
	From, To    time.Weekday
	DaysOfMonth []int

	HasWindow  bool
	Start, End time.Duration

	// Pin pins ticks to a wall clock time within each period of the frequency,
	// e.g. 1h @ 00:07:30 ticks at HH:07:30. It is -1 when there is no pin.
	Pin time.Duration

	Tz string
}
