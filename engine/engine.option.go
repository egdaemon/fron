package engine

import (
	"errors"
	"fmt"
	"time"
)

func (t Chain) add(o Option) Chain {
	return append(t, o)
}

// Weekdays restricts execution to monday through friday.
func (t Chain) Weekdays() Chain {
	return t.Days(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday)
}

// Weekends restricts execution to saturday and sunday.
func (t Chain) Weekends() Chain {
	return t.Days(time.Sunday, time.Saturday)
}

// Everyday permits every day of the week.
func (t Chain) Everyday() Chain {
	return t.DaysBetween(time.Sunday, time.Saturday)
}

// Days restricts execution to the given days of the week.
func (t Chain) Days(days ...time.Weekday) Chain {
	return t.add(func(s *scheduler) error {
		if len(days) == 0 {
			return errors.New("engine: at least one day is required")
		}

		s.resetDays()

		for _, d := range days {
			if d < time.Sunday || d > time.Saturday {
				return fmt.Errorf("engine: invalid weekday: %d", d)
			}
			s.allow(d)
		}

		return nil
	})
}

// DaysBetween restricts execution to the days from through to inclusive,
// wrapping past saturday when to is before from.
func (t Chain) DaysBetween(from, to time.Weekday) Chain {
	return t.add(func(s *scheduler) error {
		if from < time.Sunday || from > time.Saturday || to < time.Sunday || to > time.Saturday {
			return fmt.Errorf("engine: invalid weekday range: %d-%d", from, to)
		}

		s.resetDays()

		for d := from; ; d = (d + 1) % 7 {
			s.allow(d)
			if d == to {
				break
			}
		}

		return nil
	})
}

// DaysOfMonth restricts execution to the given days of the month.
func (t Chain) DaysOfMonth(days ...int) Chain {
	return t.add(func(s *scheduler) error {
		if len(days) == 0 {
			return errors.New("engine: at least one day of month is required")
		}

		s.resetDays()

		for _, d := range days {
			if d < 1 || d > 31 {
				return fmt.Errorf("engine: day of month out of range: %d", d)
			}
			s.monthdays |= 1 << d
		}

		return nil
	})
}

// During restricts execution to the window between start and end, expressed as
// offsets from midnight. An end before start wraps past midnight.
func (t Chain) During(start, end time.Duration) Chain {
	return t.add(func(s *scheduler) error {
		if start < 0 || start >= 24*time.Hour || end < 0 || end > 24*time.Hour {
			return fmt.Errorf("engine: time window out of range: %s-%s", start, end)
		}

		if start == end {
			return fmt.Errorf("engine: empty time window: %s-%s", start, end)
		}

		s.window, s.start, s.end = true, start, end
		return nil
	})
}

// InTz evaluates the conditions in the given location.
func (t Chain) InTz(loc *time.Location) Chain {
	return t.add(func(s *scheduler) error {
		if loc == nil {
			return errors.New("engine: nil timezone")
		}

		s.loc = loc
		return nil
	})
}

// InTzName evaluates the conditions in the named location, an empty name is UTC.
func (t Chain) InTzName(name string) Chain {
	return t.add(func(s *scheduler) error {
		loc, err := time.LoadLocation(name)
		if err != nil {
			return fmt.Errorf("engine: unknown timezone %q: %w", name, err)
		}

		return Chain(nil).InTz(loc)[0](s)
	})
}

// resetDays discards any earlier day restriction, the last day option wins.
func (t *scheduler) resetDays() {
	t.weekdays, t.dayorder, t.monthdays = 0, nil, 0
}

func (t *scheduler) allow(d time.Weekday) {
	if t.weekdays&(1<<d) != 0 {
		return
	}

	t.weekdays |= 1 << d
	t.dayorder = append(t.dayorder, d)
}

// Inclusive makes Next return the reference time itself when it is a tick,
// by default Next is strictly after the reference time.
func (t Chain) Inclusive() Chain {
	return t.add(func(s *scheduler) error {
		s.lastdelta = -time.Nanosecond
		return nil
	})
}

// Horizon bounds how far past the reference time Next searches for a tick,
// the default is DefaultHorizon. Next returns the zero time when no tick
// exists within it.
func (t Chain) Horizon(d time.Duration) Chain {
	return t.add(func(s *scheduler) error {
		if d <= 0 {
			return fmt.Errorf("engine: horizon must be positive: %s", d)
		}

		s.horizon = d
		return nil
	})
}

// IgnoreSeed discards the offset given to New, so a seed written in a DSL
// string no longer matters. The offset becomes zero; a pin is not affected.
func (t Chain) IgnoreSeed() Chain {
	return t.add(func(s *scheduler) error {
		s.offset = 0
		return nil
	})
}

// AutoSeed fills the offset from id when it is zero, either because none was
// supplied or IgnoreSeed cleared it. A non-empty id hashes like Seed and an
// empty id is a random offset. Options apply in order, so IgnoreSeed must come
// before AutoSeed.
func (t Chain) AutoSeed(id string) Chain {
	return t.add(func(s *scheduler) error {
		if s.offset == 0 {
			s.offset = autoOffset(id, s.hz)
		}

		return nil
	})
}

// At pins ticks to the wall clock time d, in the schedule's timezone, within
// each period of the frequency: on every local day ticks fall at d, d+hz,
// d+2hz, ... while still in that day, so for an hourly schedule 00:07:30 ticks
// at HH:07:30 and for a daily schedule 07:30 ticks at 07:30. For a frequency of
// whole days it is d on every nth local day. A pin overrides any seed or
// offset, and d must be shorter than both the frequency and a day.
func (t Chain) At(d time.Duration) Chain {
	return t.add(func(s *scheduler) error {
		if d < 0 || d >= 24*time.Hour || d >= s.hz {
			return fmt.Errorf("engine: pin out of range: %s", d)
		}

		s.pin = d
		return nil
	})
}

// IgnorePin discards a pin, ticks then follow the seed or offset.
func (t Chain) IgnorePin() Chain {
	return t.add(func(s *scheduler) error {
		s.pin = -1
		return nil
	})
}
