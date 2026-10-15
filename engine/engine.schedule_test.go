package engine_test

import (
	"testing"
	"time"

	"github.com/egdaemon/fron/engine"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should reject a non positive frequency", func(t *testing.T) {
		_, err := engine.New(0, 0)
		require.Error(t, err)
		_, err = engine.New(-time.Minute, 0)
		require.Error(t, err)
	})

	t.Run("should reject a negative offset", func(t *testing.T) {
		_, err := engine.New(time.Minute, -time.Second)
		require.Error(t, err)
	})

	t.Run("should reduce the offset into one period", func(t *testing.T) {
		s, err := engine.New(time.Hour, 90*time.Minute)
		require.NoError(t, err)
		require.Equal(t, 30*time.Minute, s.(interface{ Spec() engine.Spec }).Spec().Offset)
	})

	t.Run("should let the last option of a kind win", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().Weekdays().Weekends().DaysOfMonth(5).During(0, time.Hour).During(2*time.Hour, 3*time.Hour).InTz(time.UTC).InTz(time.FixedZone("x", 3600))...))(t)
		sp := s.(interface{ Spec() engine.Spec }).Spec()
		require.Empty(t, sp.Days)
		require.Equal(t, []int{5}, sp.DaysOfMonth)
		require.Equal(t, 2*time.Hour, sp.Start)
		require.Equal(t, "x", sp.Location.String())

		s = testx.MustT(engine.New(time.Hour, 0, engine.Options().DaysOfMonth(5).Weekends()...))(t)
		sp = s.(interface{ Spec() engine.Spec }).Spec()
		require.Equal(t, []time.Weekday{time.Sunday, time.Saturday}, sp.Days)
		require.Empty(t, sp.DaysOfMonth)
	})

	t.Run("should reject invalid options", func(t *testing.T) {
		for name, opts := range map[string][]engine.Option{
			"no days":           engine.Options().Days(),
			"bad weekday":       engine.Options().Days(time.Weekday(9)),
			"bad range":         engine.Options().DaysBetween(time.Weekday(9), time.Monday),
			"no days of month":  engine.Options().DaysOfMonth(),
			"day of month 0":    engine.Options().DaysOfMonth(0),
			"day of month 32":   engine.Options().DaysOfMonth(32),
			"empty window":      engine.Options().During(time.Hour, time.Hour),
			"window past a day": engine.Options().During(0, 25*time.Hour),
			"negative start":    engine.Options().During(-time.Hour, time.Hour),
			"nil location":      engine.Options().InTz(nil),
			"unknown tz":        engine.Options().InTzName("Mars/Olympus"),
		} {
			t.Run(name, func(t *testing.T) {
				_, err := engine.New(time.Hour, 0, opts...)
				require.Error(t, err)
			})
		}
	})
}

func TestSeed(t *testing.T) {
	t.Run("should be deterministic", func(t *testing.T) {
		require.Equal(t, engine.Seed("foo"), engine.Seed("foo"))
	})

	t.Run("should treat strings and bytes alike", func(t *testing.T) {
		require.Equal(t, engine.Seed("foo"), engine.Seed([]byte("foo")))
	})

	t.Run("should not be negative", func(t *testing.T) {
		for _, s := range []string{"a", "b", "foo", "billing-job", ""} {
			require.GreaterOrEqual(t, engine.Seed(s), time.Duration(0), s)
		}
	})

	t.Run("should differ across seeds", func(t *testing.T) {
		require.NotEqual(t, engine.Seed("foo"), engine.Seed("bar"))
	})

	t.Run("zero offset should be zero", func(t *testing.T) {
		require.Equal(t, time.Duration(0), engine.ZeroOffset)
	})
}

func TestNext(t *testing.T) {
	t.Run("should align to the frequency grid", func(t *testing.T) {
		s := testx.MustT(engine.New(15*time.Minute, 0))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 8, 14, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC)))(t))
	})

	t.Run("should shift the grid by the offset", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 20*time.Minute))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 20, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 15, 9, 20, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 8, 20, 0, 0, time.UTC)))(t))
	})

	t.Run("should skip days that do not match", func(t *testing.T) {
		// 2026-10-16 is a friday.
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().Weekdays()...))(t)
		require.Equal(t, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 16, 23, 30, 0, 0, time.UTC)))(t))
	})

	t.Run("should honour days of the month", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 0, engine.Options().DaysOfMonth(1, 15)...))(t)
		require.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 2, 7, 0, 0, 0, time.UTC)))(t))
	})

	t.Run("should restrict to a time window", func(t *testing.T) {
		s := testx.MustT(engine.New(15*time.Minute, 0, engine.Options().During(9*time.Hour, 17*time.Hour)...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 15, 16, 45, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 16, 30, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 16, 45, 0, 0, time.UTC)))(t))
	})

	t.Run("should support windows that wrap midnight", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().During(22*time.Hour, 2*time.Hour)...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 22, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 23, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 10, 16, 22, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC)))(t))
	})

	t.Run("should evaluate conditions in the configured timezone", func(t *testing.T) {
		loc := time.FixedZone("plus2", 2*60*60)
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().During(9*time.Hour, 10*time.Hour).InTz(loc)...))(t)
		// 09:00 at +02:00 is 07:00 utc.
		require.Equal(t, time.Date(2026, 10, 15, 7, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)))(t))
	})

	t.Run("should handle daylight saving transitions", func(t *testing.T) {
		loc, err := time.LoadLocation("America/New_York")
		require.NoError(t, err)

		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().During(9*time.Hour, 10*time.Hour).InTz(loc)...))(t)
		// dst ended 2026-11-01, 09:00 local is 14:00 utc after and 13:00 utc before.
		require.Equal(t, time.Date(2026, 10, 31, 13, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)))(t))
		require.Equal(t, time.Date(2026, 11, 1, 14, 0, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)))(t))
	})

	t.Run("should skip months without the day", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 0, engine.Options().DaysOfMonth(31)...))(t)
		next := testx.MustT(s.Next(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)))(t)
		require.Equal(t, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), next)
	})
}

func TestInTzName(t *testing.T) {
	t.Run("should treat an empty name as utc", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().InTzName("")...))(t)
		require.Equal(t, time.UTC, s.(interface{ Spec() engine.Spec }).Spec().Location)
	})

	t.Run("should default to utc when no timezone is configured", func(t *testing.T) {
		loc := time.FixedZone("plus2", 2*60*60)
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().During(9*time.Hour, 10*time.Hour)...))(t)
		require.Equal(t, time.UTC, s.(interface{ Spec() engine.Spec }).Spec().Location)
		// 09:00 utc, regardless of the caller's location.
		require.True(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC).Equal(testx.MustT(s.Next(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).In(loc)))(t)))
	})
}

func TestInclusive(t *testing.T) {
	tick := time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC)

	t.Run("should be strictly after by default", func(t *testing.T) {
		s := testx.MustT(engine.New(15*time.Minute, 0))(t)
		require.Equal(t, tick.Add(15*time.Minute), testx.MustT(s.Next(tick))(t))
	})

	t.Run("should return the reference time when it is a tick", func(t *testing.T) {
		s := testx.MustT(engine.New(15*time.Minute, 0, engine.Options().Inclusive()...))(t)
		require.Equal(t, tick, testx.MustT(s.Next(tick))(t))
		require.Equal(t, tick.Add(15*time.Minute), testx.MustT(s.Next(tick.Add(time.Nanosecond)))(t))
	})

	t.Run("should still apply conditions", func(t *testing.T) {
		// 2026-10-17 is a saturday, the next weekday tick is monday.
		saturday := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
		s := testx.MustT(engine.New(24*time.Hour, 0, engine.Options().Weekdays().Inclusive()...))(t)
		require.Equal(t, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC), testx.MustT(s.Next(saturday))(t))
	})

	t.Run("should be described by the spec", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().Inclusive()...))(t)
		require.Equal(t, -time.Nanosecond, s.(interface{ Spec() engine.Spec }).Spec().LastDelta)
	})
}

func TestHorizon(t *testing.T) {
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	t.Run("should default to 128 years", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0))(t)
		require.Equal(t, engine.DefaultHorizon, s.(interface{ Spec() engine.Spec }).Spec().Horizon)
		require.Equal(t, 128*365*24*time.Hour, engine.DefaultHorizon)
	})

	t.Run("should find ticks inside the horizon", func(t *testing.T) {
		// the first monday tick of a 28 day grid never exists, the first thursday is within a month.
		s := testx.MustT(engine.New(28*24*time.Hour, 0, engine.Options().Days(time.Thursday).Horizon(60*24*time.Hour)...))(t)
		_, err := s.Next(from)
		require.NoError(t, err)
	})

	t.Run("should give up beyond the horizon", func(t *testing.T) {
		// 28 day multiples from the epoch always land on a thursday.
		s := testx.MustT(engine.New(28*24*time.Hour, 0, engine.Options().Days(time.Monday)...))(t)
		_, err := s.Next(from)
		require.ErrorIs(t, err, engine.ErrNoTick)

		s = testx.MustT(engine.New(time.Hour, 0, engine.Options().Weekdays().Horizon(time.Minute)...))(t)
		_, err = s.Next(time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC))
		require.ErrorIs(t, err, engine.ErrNoTick)
	})

	t.Run("should reject a non positive horizon", func(t *testing.T) {
		_, err := engine.New(time.Hour, 0, engine.Options().Horizon(0)...)
		require.Error(t, err)
	})

	t.Run("should search past the year 2262", func(t *testing.T) {
		s := testx.MustT(engine.New(15*time.Minute, 0, engine.Options().Horizon(250*365*24*time.Hour)...))(t)
		require.Equal(t, time.Date(2300, 1, 1, 0, 15, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC)))(t))
	})

	t.Run("should align times before the unix epoch", func(t *testing.T) {
		s := testx.MustT(engine.New(15*time.Minute, 0))(t)
		require.Equal(t, time.Date(1969, 12, 31, 23, 45, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(1969, 12, 31, 23, 31, 7, 5, time.UTC)))(t))
		require.Equal(t, time.Date(1900, 1, 1, 0, 15, 0, 0, time.UTC), testx.MustT(s.Next(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)))(t))
	})
}

func TestAutoSeed(t *testing.T) {
	offset := func(s engine.Schedule) time.Duration {
		return s.(interface{ Spec() engine.Spec }).Spec().Offset
	}

	t.Run("should fill a zero offset from the id", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().AutoSeed("row-1")...))(t)
		require.Equal(t, engine.Seed("row-1")%time.Hour, offset(s))
		require.NotZero(t, offset(s))
	})

	t.Run("should leave a non zero offset alone", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 7*time.Minute, engine.Options().AutoSeed("row-1")...))(t)
		require.Equal(t, 7*time.Minute, offset(s))
	})

	t.Run("should let the first auto seed that fills the offset win", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().AutoSeed("row-1").AutoSeed("row-2")...))(t)
		require.Equal(t, engine.Seed("row-1")%time.Hour, offset(s))
	})

	t.Run("should differ across ids and be deterministic", func(t *testing.T) {
		a := testx.MustT(engine.New(time.Hour, 0, engine.Options().AutoSeed("a")...))(t)
		b := testx.MustT(engine.New(time.Hour, 0, engine.Options().AutoSeed("b")...))(t)
		again := testx.MustT(engine.New(time.Hour, 0, engine.Options().AutoSeed("a")...))(t)
		require.NotEqual(t, offset(a), offset(b))
		require.Equal(t, offset(a), offset(again))
	})

	t.Run("should keep the offset within one period", func(t *testing.T) {
		for _, id := range []string{"a", "b", "c", "row-1", "row-2"} {
			s := testx.MustT(engine.New(15*time.Minute, 0, engine.Options().AutoSeed(id)...))(t)
			require.Less(t, offset(s), 15*time.Minute, id)
			require.GreaterOrEqual(t, offset(s), time.Duration(0), id)
		}
	})

	t.Run("should use a random offset for an empty id", func(t *testing.T) {
		seen := make(map[time.Duration]struct{})
		for range 50 {
			s := testx.MustT(engine.New(24*time.Hour, 0, engine.Options().AutoSeed("")...))(t)
			require.Less(t, offset(s), 24*time.Hour)
			require.GreaterOrEqual(t, offset(s), time.Duration(0))
			seen[offset(s)] = struct{}{}
		}
		require.Greater(t, len(seen), 1)
	})

	t.Run("should keep a zero offset for the reserved zero offset seed", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 0, engine.Options().AutoSeed(engine.ZeroOffsetSeed)...))(t)
		require.Equal(t, time.Duration(0), offset(s))

		s = testx.MustT(engine.New(time.Hour, 0, engine.Options().IgnoreSeed().AutoSeed(engine.ZeroOffsetSeed)...))(t)
		require.Equal(t, time.Duration(0), offset(s))
	})
}

func TestIgnoreSeed(t *testing.T) {
	offset := func(s engine.Schedule) time.Duration {
		return s.(interface{ Spec() engine.Spec }).Spec().Offset
	}

	t.Run("should discard the given offset", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 7*time.Minute, engine.Options().IgnoreSeed()...))(t)
		require.Equal(t, time.Duration(0), offset(s))
	})

	t.Run("should let the auto seed supply the offset afterwards", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 7*time.Minute, engine.Options().IgnoreSeed().AutoSeed("row-1")...))(t)
		require.Equal(t, engine.Seed("row-1")%time.Hour, offset(s))
	})

	t.Run("should apply in order so the ignore must come first", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 7*time.Minute, engine.Options().AutoSeed("row-1").IgnoreSeed()...))(t)
		require.Equal(t, time.Duration(0), offset(s))
	})
}

func TestWindowSnap(t *testing.T) {
	day := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, time.UTC) }

	t.Run("should place a zero offset daily schedule at the start of the window", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 0, engine.Options().During(9*time.Hour, 17*time.Hour)...))(t)
		require.Equal(t, day(15, 9, 0), testx.MustT(s.Next(day(15, 0, 0)))(t))
		require.Equal(t, day(16, 9, 0), testx.MustT(s.Next(day(15, 9, 0)))(t))
		require.Equal(t, day(17, 9, 0), testx.MustT(s.Next(day(16, 9, 0)))(t))
	})

	t.Run("should smear the offset within the window when it falls outside", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 3*time.Hour+25*time.Minute, engine.Options().During(9*time.Hour, 17*time.Hour)...))(t)
		require.Equal(t, day(15, 12, 25), testx.MustT(s.Next(day(15, 0, 0)))(t))
		require.Equal(t, day(16, 12, 25), testx.MustT(s.Next(day(15, 12, 25)))(t))
	})

	t.Run("should wrap a large offset into the window", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 20*time.Hour+10*time.Minute, engine.Options().During(9*time.Hour, 10*time.Hour)...))(t)
		require.Equal(t, day(15, 9, 10), testx.MustT(s.Next(day(15, 0, 0)))(t))
	})

	t.Run("should keep the offset when it falls inside the window", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 9*time.Hour+30*time.Minute, engine.Options().During(9*time.Hour, 10*time.Hour)...))(t)
		require.Equal(t, day(15, 9, 30), testx.MustT(s.Next(day(15, 0, 0)))(t))
		require.Equal(t, day(16, 9, 30), testx.MustT(s.Next(day(15, 9, 30)))(t))
	})

	t.Run("should place a tick in the window when the frequency steps over it", func(t *testing.T) {
		s := testx.MustT(engine.New(6*time.Hour, 0, engine.Options().During(9*time.Hour, 10*time.Hour)...))(t)
		require.Equal(t, day(15, 9, 0), testx.MustT(s.Next(day(15, 0, 0)))(t))
		require.Equal(t, day(16, 9, 0), testx.MustT(s.Next(day(15, 9, 0)))(t))
	})

	t.Run("should not add a tick when the grid already lands in the window", func(t *testing.T) {
		s := testx.MustT(engine.New(time.Hour, 20*time.Minute, engine.Options().During(9*time.Hour, 10*time.Hour)...))(t)
		require.Equal(t, day(15, 9, 20), testx.MustT(s.Next(day(15, 0, 0)))(t))
		require.Equal(t, day(16, 9, 20), testx.MustT(s.Next(day(15, 9, 20)))(t))
	})

	t.Run("should snap windows that wrap midnight", func(t *testing.T) {
		s := testx.MustT(engine.New(24*time.Hour, 13*time.Hour+30*time.Minute, engine.Options().During(22*time.Hour, 2*time.Hour)...))(t)
		require.Equal(t, day(15, 23, 30), testx.MustT(s.Next(day(15, 0, 0)))(t))
		require.Equal(t, day(16, 23, 30), testx.MustT(s.Next(day(15, 23, 30)))(t))
	})

	t.Run("should snap on the grid days of a multi day frequency", func(t *testing.T) {
		s := testx.MustT(engine.New(72*time.Hour, 0, engine.Options().During(6*time.Hour, 12*time.Hour)...))(t)
		first := testx.MustT(s.Next(day(15, 0, 0)))(t)
		second := testx.MustT(s.Next(first))(t)
		require.Equal(t, 6, first.Hour())
		require.Equal(t, 72*time.Hour, second.Sub(first))
	})

	t.Run("should still apply day conditions", func(t *testing.T) {
		// 2026-10-17 is a saturday, so the first weekday window start is monday.
		s := testx.MustT(engine.New(24*time.Hour, 0, engine.Options().Weekdays().During(9*time.Hour, 17*time.Hour)...))(t)
		require.Equal(t, day(19, 9, 0), testx.MustT(s.Next(day(16, 17, 0)))(t))
	})
}
