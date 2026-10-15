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
// A schedule with no seed of its own gets a random slot; use Seeded to make it
// stable.
func Compile(t T, options ...Option) (Schedule, error) {
	return ast.Compile(t, options...)
}

// CRON renders the tree as a 5 field cron expression.
func CRON(t T, options ...Option) (string, error) {
	return ast.CRON(t, options...)
}

// NextTick parses s and returns the first tick strictly after t. A schedule
// with no seed of its own gets a random slot; use Seeded to make it stable.
// Pass fron.Options().Inclusive()... to also accept t itself. It returns
// ErrNoTick when no tick exists within the horizon.
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

// Seeder gives schedules that carry no seed of their own a stable slot derived
// from its seed. The zero value has an empty seed, which is a random slot, the
// same as calling the package level functions.
type Seeder struct {
	seed string
}

// Seeded returns a Seeder for the seed, e.g. the id of the row a schedule
// belongs to: the same seed always lands in the same slot.
func Seeded[T ~string | ~[]byte](seed T) Seeder {
	return Seeder{seed: string(seed)}
}

// with returns a copy of options that ends with the seed, so the caller's
// slice is never written to.
func (t Seeder) with(options []Option) []Option {
	return append(append(make([]Option, 0, len(options)+1), options...), Options().AutoSeed(t.seed)...)
}

// New builds a schedule that ticks every hz, at the slot of the seed.
func (t Seeder) New(hz time.Duration, options ...Option) (Schedule, error) {
	return New(hz, 0, t.with(options)...)
}

// Compile translates the tree into a schedule at the slot of the seed unless
// the tree carries a seed of its own.
func (t Seeder) Compile(tree T, options ...Option) (Schedule, error) {
	return Compile(tree, t.with(options)...)
}

// NextTick parses s and returns the first tick strictly after at, in the slot
// of the seed unless s carries a seed of its own.
func (t Seeder) NextTick(at time.Time, s string, options ...Option) (time.Time, error) {
	return NextTick(at, s, t.with(options)...)
}
