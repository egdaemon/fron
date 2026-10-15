// Package fron is a frequency-first scheduler. It delegates to fron/ast, which
// parses the textual DSL, and fron/engine, which computes ticks, so a
// single import covers both.
package fron

import (
	"time"

	"github.com/egdaemon/fron/ast"
	"github.com/egdaemon/fron/engine"
)

type (
	// Schedule computes when a task should next run.
	Schedule = engine.Schedule
	// Option configures a schedule.
	Option = engine.Option
	// Chain is a chainable list of options.
	Chain = engine.Chain
	// Spec is a read only description of a schedule's configuration.
	Spec = engine.Spec
	// T is a node in the parsed schedule tree.
	T = ast.T
)

const (
	// DefaultHorizon is how far past the reference time Next searches by default.
	DefaultHorizon = engine.DefaultHorizon
	// ZeroOffset is the cron-equivalent offset: ticks land on clock boundaries.
	ZeroOffset = engine.ZeroOffset
	// ZeroOffsetSeed is the reserved DSL seed selecting ZeroOffset.
	ZeroOffsetSeed = engine.ZeroOffsetSeed
)

// ErrNoTick is returned by Next when no tick exists within the horizon.
var ErrNoTick = engine.ErrNoTick

// New builds a schedule that ticks every hz, shifted by offset.
func New(hz, offset time.Duration, options ...Option) (Schedule, error) {
	return engine.New(hz, offset, options...)
}

// Options starts an empty option chain.
func Options() Chain {
	return engine.Options()
}

// Seed hashes a seed identifier into a deterministic offset.
func Seed[T ~string | ~[]byte](s T) time.Duration {
	return engine.Seed(s)
}

// Parse reads a DSL string into a tree in a single pass.
func Parse(input string) (T, error) {
	return ast.Parse(input)
}

// Compile translates the tree into a schedule, extra options are applied last.
func Compile(t T, options ...Option) (Schedule, error) {
	return ast.Compile(t, options...)
}

// CRON renders the tree as a 5 field cron expression.
func CRON(t T, options ...Option) (string, error) {
	return ast.CRON(t, options...)
}

// NextTick parses s and returns the first tick strictly after t. Pass
// fron.Options().Inclusive()... to also accept t itself. It returns ErrNoTick
// when no tick exists within the horizon.
func NextTick(t time.Time, s string, options ...Option) (time.Time, error) {
	tree, err := ast.Parse(s)
	if err != nil {
		return time.Time{}, err
	}

	sched, err := ast.Compile(tree, options...)
	if err != nil {
		return time.Time{}, err
	}

	return sched.Next(t)
}
