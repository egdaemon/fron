package fron_test

import (
	"testing"
	"time"

	"github.com/egdaemon/fron"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestNextTick(t *testing.T) {
	t.Run("should be strictly after the reference time by default", func(t *testing.T) {
		for from, expected := range map[time.Time]time.Time{
			time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC):  time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC),
			time.Date(2026, 10, 15, 8, 14, 0, 0, time.UTC): time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC),
			time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC): time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC),
		} {
			actual := testx.MustT(fron.NextTick(from, pin00("15m"), ""))(t)
			require.Equal(t, expected, actual, from)
		}
	})

	t.Run("should accept the reference time when inclusive", func(t *testing.T) {
		opts := fron.Options().Inclusive()

		actual := testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), pin00("15m"), "", opts...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 8, 16, 0, 0, time.UTC), pin00("15m"), "", opts...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC), actual)
	})

	t.Run("should honour conditions", func(t *testing.T) {
		// 2026-10-16 is a friday.
		actual := testx.MustT(fron.NextTick(time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC), pin00("1h @ on weekdays during 09:00-17:00"), ""))(t)
		require.Equal(t, time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), pin00("1d @ on day 1,15"), ""))(t)
		require.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 23, 0, 0, 0, time.UTC), pin00("1h @ during 22:00-02:00"), ""))(t)
		require.Equal(t, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), actual)
	})

	t.Run("should handle daylight saving transitions", func(t *testing.T) {
		// dst ended 2026-11-01 in new york, 09:00 local moves from 13:00 to 14:00 utc.
		actual := testx.MustT(fron.NextTick(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), pin00("1h @ during 09:00-10:00 in America/New_York"), ""))(t)
		require.Equal(t, time.Date(2026, 11, 1, 14, 0, 0, 0, time.UTC), actual)
	})

	t.Run("should return the result in the reference location", func(t *testing.T) {
		loc := time.FixedZone("plus2", 2*60*60)
		from := time.Date(2026, 10, 15, 8, 0, 0, 0, loc)
		actual := testx.MustT(fron.NextTick(from, pin00("15m"), ""))(t)
		require.Equal(t, loc, actual.Location())
		require.True(t, from.Add(15*time.Minute).Equal(actual))
	})

	t.Run("should be deterministic for a typed seed or a fallback seed", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)
		require.Equal(t, testx.MustT(fron.NextTick(from, "billing-job | 2h", ""))(t), testx.MustT(fron.NextTick(from, "billing-job | 2h", ""))(t))
		require.Equal(t, testx.MustT(fron.NextTick(from, "30m", "row-1"))(t), testx.MustT(fron.NextTick(from, "30m", "row-1"))(t))
	})

	t.Run("should fail for invalid input", func(t *testing.T) {
		for _, in := range []string{"", "1h @ on 2nd wed", "*/15 * * * *", "1h @ in Mars/Olympus"} {
			_, err := fron.NextTick(time.Now(), in, "")
			require.Error(t, err, in)
		}
	})

	t.Run("should fail when no tick exists", func(t *testing.T) {
		// 28 day multiples from the epoch always land on a thursday.
		_, err := fron.NextTick(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), pin00("28d @ on mon"), "")
		require.ErrorIs(t, err, fron.ErrNoTick)
	})
}

func TestNextTickSeeds(t *testing.T) {
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	t.Run("should give every id its own stable slot", func(t *testing.T) {
		a := testx.MustT(fron.NextTick(from, "1h", "", fron.Options().IgnoreSeed().AutoSeed("row-a")...))(t)
		b := testx.MustT(fron.NextTick(from, "1h", "", fron.Options().IgnoreSeed().AutoSeed("row-b")...))(t)
		again := testx.MustT(fron.NextTick(from, "1h", "", fron.Options().IgnoreSeed().AutoSeed("row-a")...))(t)

		require.NotEqual(t, a, b)
		require.Equal(t, a, again)
		require.True(t, a.After(from))
	})

	t.Run("should ignore a seed typed in the string", func(t *testing.T) {
		opts := fron.Options().IgnoreSeed().AutoSeed("row-a")
		plain := testx.MustT(fron.NextTick(from, "1h", "", opts...))(t)
		typed := testx.MustT(fron.NextTick(from, "someone-else | 1h", "", opts...))(t)

		require.Equal(t, plain, typed)
	})

	t.Run("should keep conditions when seeding", func(t *testing.T) {
		// 2026-10-17 is a saturday.
		got := testx.MustT(fron.NextTick(time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC), "1h @ on weekdays", "", fron.Options().IgnoreSeed().AutoSeed("row-a")...))(t)
		require.Equal(t, time.Monday, got.Weekday())
	})
}

func TestNextTickWindowSnap(t *testing.T) {
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	t.Run("should smear a daily schedule within the window by seed", func(t *testing.T) {
		seen := make(map[time.Time]struct{})
		for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
			opts := fron.Options().IgnoreSeed().AutoSeed(id)
			got := testx.MustT(fron.NextTick(from, "1d @ during 09:00-17:00", "", opts...))(t)
			require.False(t, got.Before(time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)), id)
			require.Less(t, got, time.Date(2026, 10, 15, 17, 0, 0, 0, time.UTC), id)
			require.Equal(t, got, testx.MustT(fron.NextTick(from, "1d @ during 09:00-17:00", "", opts...))(t), id)
			seen[got] = struct{}{}
		}
		require.Greater(t, len(seen), 1)
	})

	t.Run("should keep a daily tick that already falls inside the window", func(t *testing.T) {
		got := testx.MustT(fron.NextTick(from, "1d @ 00:00 during 00:00-10:00", ""))(t)
		require.Equal(t, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), got)
	})

	t.Run("should render the same time the cron form fires", func(t *testing.T) {
		tree := testx.MustT(fron.Parse("1d @ 12:00 during 09:00-17:00"))(t)
		require.Equal(t, "0 12 * * *", testx.MustT(fron.CRON(tree))(t))
		require.Equal(t, time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC), testx.MustT(fron.NextTick(from, "1d @ 12:00 during 09:00-17:00", ""))(t))
	})
}

func TestNextTickFallbackSeed(t *testing.T) {
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	t.Run("should give each fallback seed its own stable slot", func(t *testing.T) {
		a := testx.MustT(fron.NextTick(from, "1h", "row-a"))(t)
		b := testx.MustT(fron.NextTick(from, "1h", "row-b"))(t)
		require.NotEqual(t, a, b)
		require.Equal(t, a, testx.MustT(fron.NextTick(from, "1h", "row-a"))(t))
	})

	t.Run("should not depend on the text of the schedule", func(t *testing.T) {
		// two different strings with the same fallback land on the same offset.
		a := testx.MustT(fron.NextTick(from, "1h", "row-a"))(t)
		b := testx.MustT(fron.NextTick(from, "1h @ on everyday", "row-a"))(t)
		require.Equal(t, a, b)
	})

	t.Run("should pick a random slot for an empty fallback seed", func(t *testing.T) {
		seen := make(map[time.Time]struct{})
		for range 40 {
			got := testx.MustT(fron.NextTick(from, "1h", ""))(t)
			require.True(t, got.After(from))
			require.False(t, got.After(from.Add(time.Hour)))
			seen[got] = struct{}{}
		}
		require.Greater(t, len(seen), 1)
	})

	t.Run("should let a typed seed beat the fallback", func(t *testing.T) {
		typed := testx.MustT(fron.NextTick(from, "someone | 1h", "row-a"))(t)
		require.Equal(t, typed, testx.MustT(fron.NextTick(from, "someone | 1h", "row-b"))(t))
	})

	t.Run("should let a pin beat both the typed seed and the fallback", func(t *testing.T) {
		want := time.Date(2026, 10, 15, 0, 7, 30, 0, time.UTC)
		require.Equal(t, want, testx.MustT(fron.NextTick(from, "1h @ 00:07:30", "row-a"))(t))
		require.Equal(t, want, testx.MustT(fron.NextTick(from, "someone | 1h @ 00:07:30", "row-a"))(t))
	})

	t.Run("should let IgnoreSeed hand the schedule to the fallback", func(t *testing.T) {
		ignore := fron.Options().IgnoreSeed()
		plain := testx.MustT(fron.NextTick(from, "1h", "row-a"))(t)
		require.Equal(t, plain, testx.MustT(fron.NextTick(from, "someone | 1h", "row-a", ignore...))(t))
	})
}

func TestNextTickPinnedTime(t *testing.T) {
	from := time.Date(2026, 10, 15, 8, 20, 0, 0, time.UTC)

	t.Run("should tick at the same time past every hour", func(t *testing.T) {
		in := "1h @ 00:07:30"
		first := testx.MustT(fron.NextTick(from, in, ""))(t)
		second := testx.MustT(fron.NextTick(first, in, ""))(t)
		require.Equal(t, time.Date(2026, 10, 15, 9, 7, 30, 0, time.UTC), first)
		require.Equal(t, time.Date(2026, 10, 15, 10, 7, 30, 0, time.UTC), second)
	})

	t.Run("should repeat within the day for a shorter frequency", func(t *testing.T) {
		in := "15m @ 00:02:30"
		first := testx.MustT(fron.NextTick(from, in, ""))(t)
		second := testx.MustT(fron.NextTick(first, in, ""))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 32, 30, 0, time.UTC), first)
		require.Equal(t, time.Date(2026, 10, 15, 8, 47, 30, 0, time.UTC), second)
	})

	t.Run("should pin a daily time to the millisecond", func(t *testing.T) {
		got := testx.MustT(fron.NextTick(from, "1d @ 09:07:30.250", ""))(t)
		require.Equal(t, time.Date(2026, 10, 15, 9, 7, 30, 250_000_000, time.UTC), got)
	})

	t.Run("should pin a daily time on weekdays", func(t *testing.T) {
		// 2026-10-16 is a friday.
		got := testx.MustT(fron.NextTick(time.Date(2026, 10, 16, 7, 30, 0, 0, time.UTC), "1d @ 07:30 on weekdays", ""))(t)
		require.Equal(t, time.Date(2026, 10, 19, 7, 30, 0, 0, time.UTC), got)
	})

	t.Run("should keep the period boundary for a pin of zero rather than randomizing it", func(t *testing.T) {
		got := testx.MustT(fron.NextTick(from, "1h @ 00:00", ""))(t)
		require.Equal(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC), got)
	})

	t.Run("should be a wall clock time in the schedule's timezone", func(t *testing.T) {
		in := "1d @ 07:30 in America/New_York"
		// dst ends 2026-11-01: 07:30 is 11:30 utc before and 12:30 utc after.
		before := testx.MustT(fron.NextTick(time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC), in, ""))(t)
		after := testx.MustT(fron.NextTick(before, in, ""))(t)
		require.Equal(t, time.Date(2026, 10, 31, 11, 30, 0, 0, time.UTC), before)
		require.Equal(t, time.Date(2026, 11, 1, 12, 30, 0, 0, time.UTC), after)
	})

	t.Run("should combine with a window", func(t *testing.T) {
		got := testx.MustT(fron.NextTick(from, "1h @ 00:07:30 during 09:00-17:00", ""))(t)
		require.Equal(t, time.Date(2026, 10, 15, 9, 7, 30, 0, time.UTC), got)

		got = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 16, 7, 30, 0, time.UTC), "1h @ 00:07:30 during 09:00-17:00", ""))(t)
		require.Equal(t, time.Date(2026, 10, 16, 9, 7, 30, 0, time.UTC), got)
	})

	t.Run("should keep a cron aligned window with a pin of zero", func(t *testing.T) {
		got := testx.MustT(fron.NextTick(from, "15m @ 00:00 during 09:00-17:00", ""))(t)
		require.Equal(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC), got)
	})

	t.Run("should not be changed by IgnoreSeed", func(t *testing.T) {
		want := testx.MustT(fron.NextTick(from, "1h @ 00:07:30", "row-a"))(t)
		got := testx.MustT(fron.NextTick(from, "someone | 1h @ 00:07:30", "row-a", fron.Options().IgnoreSeed()...))(t)
		require.Equal(t, want, got)
	})

	t.Run("should be dropped by IgnorePin so callers keep control of the slot", func(t *testing.T) {
		seeded := testx.MustT(fron.NextTick(from, "1h", "row-a"))(t)
		ignored := testx.MustT(fron.NextTick(from, "1h @ 00:07:30", "row-a", fron.Options().IgnorePin()...))(t)
		require.Equal(t, seeded, ignored)
	})

	t.Run("should leave a typed seed in effect when only IgnorePin is used", func(t *testing.T) {
		seeded := testx.MustT(fron.NextTick(from, "someone | 1h", "row-a"))(t)
		ignored := testx.MustT(fron.NextTick(from, "someone | 1h @ 00:07:30", "row-a", fron.Options().IgnorePin()...))(t)
		require.Equal(t, seeded, ignored)
	})

	t.Run("should leave only the fallback seed when both are ignored", func(t *testing.T) {
		seeded := testx.MustT(fron.NextTick(from, "1h", "row-a"))(t)
		ignored := testx.MustT(fron.NextTick(from, "someone | 1h @ 00:07:30", "row-a", fron.Options().IgnoreSeed().IgnorePin()...))(t)
		require.Equal(t, seeded, ignored)
	})
}
