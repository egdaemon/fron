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
	var (
		loc   = t.loc
		from  = last.Add(t.lastdelta)
		limit = last.Add(t.horizon)
		next  = t.align(from)
	)

	// a grid point at or before from can snap to a window start after it, so
	// start one grid point back when there is a window.
	if t.window {
		next = next.Add(-t.hz)
	}

	for !next.After(limit) {
		local := next.In(loc)
		tomorrow := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, loc)

		switch {
		case !t.day(local):
			next = t.skip(next, tomorrow)
		case t.window && !t.within(wall(local)):
			// outside the window: move to the start of the next occurrence.
			if wall(local) < t.start {
				next = t.snap(at(local, t.start))
			} else {
				next = t.snap(at(tomorrow, t.start))
			}
		case !next.After(from):
			next = t.align(next)
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
	if t.pin >= 0 {
		return t.alignLocal(from)
	}

	hz := uint64(t.hz)
	rem := (epochmod(from, hz) + hz - uint64(t.offset)%hz) % hz

	return from.Add(time.Duration(hz - rem))
}

// alignLocal returns the first pinned tick strictly after from. pinned grids
// are anchored at local midnight: on each local day ticks fall at pin,
// pin+hz, ... while still in the day, and a frequency of whole days ticks at
// pin on every nth local day (counted from the unix epoch day).
func (t *scheduler) alignLocal(from time.Time) time.Time {
	const day = 24 * time.Hour

	local := from.In(t.loc)
	first := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, t.loc)

	n := int64(1)
	if t.hz >= day && t.hz%day == 0 {
		n = int64(t.hz / day)
	}

	for i := 0; ; i++ {
		d := first.AddDate(0, 0, i)
		if n > 1 && civilDay(d)%n != 0 {
			continue
		}

		var k int64
		if i == 0 {
			// skip the ticks of today that are already past.
			if elapsed := wall(local); elapsed >= t.pin {
				k = int64((elapsed-t.pin)/t.hz) + 1
			}
		}

		for ; ; k++ {
			off := t.pin + time.Duration(k)*t.hz
			if off >= day || (t.hz >= day && k > 0) {
				break
			}

			if c := at(d, off); c.After(from) {
				return c
			}
		}
	}
}

// civilDay is the number of days from the unix epoch to the civil date of t.
func civilDay(t time.Time) int64 {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix() / 86400
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

// at is the wall clock time of day d on the same date as t. a time skipped by
// a daylight saving transition moves forward by the length of the gap.
func at(t time.Time, d time.Duration) time.Time {
	c := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, int(d/time.Second), int(d%time.Second), t.Location())
	if c.Day() == t.Day() {
		if got := wall(c); got < d {
			return c.Add(d - got)
		}
	}

	return c
}

// snap returns the first tick of the window occurrence starting at ws: the
// first grid point inside it, or when the frequency steps over the whole
// window (e.g. a daily schedule with a morning window) the offset smeared
// within the window, so ticks still spread out instead of piling up on ws.
func (t *scheduler) snap(ws time.Time) time.Time {
	if g := t.align(ws.Add(-time.Nanosecond)); g.Sub(ws) < t.span() {
		return g
	}

	return ws.Add(t.offset % t.span())
}

// span is the length of the window.
func (t *scheduler) span() time.Duration {
	if t.start < t.end {
		return t.end - t.start
	}

	return 24*time.Hour - t.start + t.end
}
