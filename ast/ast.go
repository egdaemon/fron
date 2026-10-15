// Package ast parses the fron textual DSL into a tree, compiles that tree into
// a engine.Schedule, and renders it as a cron expression.
package ast

import "time"

// ZeroOffsetSeed is the reserved seed selecting cron-equivalent mode.
const ZeroOffsetSeed = "fron.zero.offset"

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
	// Input is the trimmed source text, used to auto seed when no seed is given.
	Input string

	Seed       string
	HasSeed    bool
	ZeroOffset bool

	Hz time.Duration

	day         dayclause
	Days        []time.Weekday
	From, To    time.Weekday
	DaysOfMonth []int

	HasWindow  bool
	Start, End time.Duration

	Tz string
}
