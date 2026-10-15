package engine

import (
	"math/bits"
	"time"
)

// DefaultHorizon is how far past the reference time Next searches by default.
const DefaultHorizon = 128 * 365 * 24 * time.Hour

// Next returns the first tick after last that satisfies the configured
// conditions, or ErrNoTick if none exists within the horizon (DefaultHorizon
// unless set with Horizon).
func (t *scheduler) Next(last time.Time) (time.Time, error) {
	loc := t.loc

	limit := last.Add(t.horizon)
	next := t.align(last.Add(t.lastdelta))
	for !next.After(limit) {
		local := next.In(loc)
		tomorrow := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, loc)

		switch {
		case !t.day(local):
			next = t.skip(next, tomorrow)
		case t.window && !t.within(wall(local)):
			if wall(local) < t.start {
				next = t.skip(next, at(local, t.start))
			} else {
				next = t.skip(next, tomorrow)
			}
		default:
			return next.In(last.Location()), nil
		}
	}

	return time.Time{}, ErrNoTick
}

// skip returns the first grid point at or after target. it always advances past
// from, which daylight saving transitions could otherwise prevent.
func (t *scheduler) skip(from, target time.Time) time.Time {
	if !target.After(from) {
		return t.align(from)
	}

	return t.align(target.Add(-time.Nanosecond))
}

// align returns the first grid point strictly after from. ticks sit at the
// unix epoch plus a multiple of hz plus the offset. the distance from the epoch
// is held as a 128 bit count of nanoseconds so it never overflows.
func (t *scheduler) align(from time.Time) time.Time {
	hz := uint64(t.hz)
	rem := (epochmod(from, hz) + hz - uint64(t.offset)%hz) % hz

	return from.Add(time.Duration(hz - rem))
}

// epochmod is the nanoseconds since the unix epoch modulo hz, in [0, hz).
func epochmod(at time.Time, hz uint64) uint64 {
	secs, nsec := at.Unix(), uint64(at.Nanosecond())
	if secs >= 0 {
		hi, lo := bits.Mul64(uint64(secs), 1e9)
		lo, carry := bits.Add64(lo, nsec, 0)
		hi += carry
		_, rem := bits.Div64(hi%hz, lo, hz)
		return rem
	}

	// before the epoch the distance is -(|secs|*1e9 - nsec).
	hi, lo := bits.Mul64(uint64(-secs), 1e9)
	lo, borrow := bits.Sub64(lo, nsec, 0)
	hi -= borrow
	_, rem := bits.Div64(hi%hz, lo, hz)
	return (hz - rem) % hz
}

func (t *scheduler) day(local time.Time) bool {
	if t.weekdays != 0 && t.weekdays&(1<<local.Weekday()) == 0 {
		return false
	}

	if t.monthdays != 0 && t.monthdays&(1<<local.Day()) == 0 {
		return false
	}

	return true
}

func (t *scheduler) within(elapsed time.Duration) bool {
	if t.start < t.end {
		return elapsed >= t.start && elapsed < t.end
	}

	return elapsed >= t.start || elapsed < t.end
}

// wall is the time of day on the wall clock, which differs from the elapsed
// time since midnight on days with a daylight saving transition.
func wall(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second +
		time.Duration(t.Nanosecond())
}

// at is the wall clock time of day d on the same date as t.
func at(t time.Time, d time.Duration) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, int(d/time.Second), int(d%time.Second), t.Location())
}
