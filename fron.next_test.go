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
			actual := testx.MustT(fron.NextTick(from, pin00("15m")))(t)
			require.Equal(t, expected, actual, from)
		}
	})

	t.Run("generate a proper sequence", func(t *testing.T) {
		t.Run("1d", func(t *testing.T) {
			s := fron.Seeded("d1a9e71c-40bd-4eed-861b-25e4c452bb0c")
			ts0 := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
			ts1, err := s.NextTick(ts0, "1d")
			require.NoError(t, err)
			require.Equal(t,
				time.Date(2026, time.January, 1, 12, 34, 56, 418220144, time.UTC),
				ts1,
			)
			ts2, err := s.NextTick(ts1, "1d")
			require.NoError(t, err)
			require.Equal(
				t,
				time.Date(2026, time.January, 2, 12, 34, 56, 418220144, time.UTC),
				ts2,
			)
		})

		t.Run("1h", func(t *testing.T) {
			s := fron.Seeded("d1a9e71c-40bd-4eed-861b-25e4c452bb0c")
			ts0 := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
			ts1, err := s.NextTick(ts0, "1h")
			require.NoError(t, err)
			require.Equal(t,
				time.Date(2026, time.January, 1, 0, 34, 56, 418220144, time.UTC),
				ts1,
			)
			ts2, err := s.NextTick(ts1, "1h")
			require.NoError(t, err)
			require.Equal(
				t,
				time.Date(2026, time.January, 1, 1, 34, 56, 418220144, time.UTC),
				ts2,
			)
		})

		t.Run("14d @ on sat", func(t *testing.T) {
			// the options the terraform provider uses for scheduled workloads.
			opts := fron.Options().IgnoreSeed().IgnorePin().InTz(time.UTC)
			s := fron.Seeded("row-a")
			ts0 := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
			ts1, err := s.NextTick(ts0, "14d @ on sat", opts...)
			require.NoError(t, err)
			require.Equal(t,
				time.Date(2026, time.October, 31, 10, 24, 53, 721551607, time.UTC),
				ts1,
			)
			require.Equal(t, time.Saturday, ts1.Weekday())
			ts2, err := s.NextTick(ts1, "14d @ on sat", opts...)
			require.NoError(t, err)
			require.Equal(t,
				time.Date(2026, time.November, 14, 10, 24, 53, 721551607, time.UTC),
				ts2,
			)
			require.Equal(t, time.Saturday, ts2.Weekday())
			ts3, err := s.NextTick(ts2, "14d @ on sat", opts...)
			require.NoError(t, err)
			require.Equal(t,
				time.Date(2026, time.November, 28, 10, 24, 53, 721551607, time.UTC),
				ts3,
			)
			require.Equal(t, time.Saturday, ts3.Weekday())
			ts4, err := s.NextTick(ts3, "14d @ on sat", opts...)
			require.NoError(t, err)
			require.Equal(t,
				time.Date(2026, time.December, 12, 10, 24, 53, 721551607, time.UTC),
				ts4,
			)
			require.Equal(t, time.Saturday, ts4.Weekday())
		})
	})

	t.Run("should accept the reference time when inclusive", func(t *testing.T) {
		opts := fron.Options().Inclusive()

		actual := testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), pin00("15m"), opts...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 8, 16, 0, 0, time.UTC), pin00("15m"), opts...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC), actual)
	})

	t.Run("should honour conditions", func(t *testing.T) {
		// 2026-10-16 is a friday.
		actual := testx.MustT(fron.NextTick(time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC), pin00("1h @ on weekdays during 09:00-17:00")))(t)
		require.Equal(t, time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), pin00("1d @ on day 1,15")))(t)
		require.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 23, 0, 0, 0, time.UTC), pin00("1h @ during 22:00-02:00")))(t)
		require.Equal(t, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), actual)
	})

	t.Run("should handle daylight saving transitions", func(t *testing.T) {
		// dst ended 2026-11-01 in new york, 09:00 local moves from 13:00 to 14:00 utc.
		actual := testx.MustT(fron.NextTick(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), pin00("1h @ during 09:00-10:00 in America/New_York")))(t)
		require.Equal(t, time.Date(2026, 11, 1, 14, 0, 0, 0, time.UTC), actual)
	})

	t.Run("should return the result in the reference location", func(t *testing.T) {
		loc := time.FixedZone("plus2", 2*60*60)
		from := time.Date(2026, 10, 15, 8, 0, 0, 0, loc)
		actual := testx.MustT(fron.NextTick(from, pin00("15m")))(t)
		require.Equal(t, loc, actual.Location())
		require.True(t, from.Add(15*time.Minute).Equal(actual))
	})

	t.Run("should be deterministic for a typed seed or a fallback seed", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)
		require.Equal(t, testx.MustT(fron.NextTick(from, "billing-job | 2h"))(t), testx.MustT(fron.NextTick(from, "billing-job | 2h"))(t))
		require.Equal(t, testx.MustT(fron.Seeded("row-1").NextTick(from, "30m"))(t), testx.MustT(fron.Seeded("row-1").NextTick(from, "30m"))(t))
	})

	t.Run("should fail for invalid input", func(t *testing.T) {
		for _, in := range []string{"", "1h @ on 2nd wed", "*/15 * * * *", "1h @ in Mars/Olympus"} {
			_, err := fron.NextTick(time.Now(), in)
			require.Error(t, err, in)
		}
	})

	t.Run("should fail when no tick exists", func(t *testing.T) {
		// 28 day multiples from the epoch always land on a thursday.
		_, err := fron.NextTick(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), pin00("28d @ on mon"))
		require.ErrorIs(t, err, fron.ErrNoTick)
	})

	t.Run("seeds", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		t.Run("should give every id its own stable slot", func(t *testing.T) {
			a := testx.MustT(fron.NextTick(from, "1h", fron.Options().IgnoreSeed().AutoSeed("row-a")...))(t)
			b := testx.MustT(fron.NextTick(from, "1h", fron.Options().IgnoreSeed().AutoSeed("row-b")...))(t)
			again := testx.MustT(fron.NextTick(from, "1h", fron.Options().IgnoreSeed().AutoSeed("row-a")...))(t)

			require.NotEqual(t, a, b)
			require.Equal(t, a, again)
			require.True(t, a.After(from))
		})

		t.Run("should ignore a seed typed in the string", func(t *testing.T) {
			opts := fron.Options().IgnoreSeed().AutoSeed("row-a")
			plain := testx.MustT(fron.NextTick(from, "1h", opts...))(t)
			typed := testx.MustT(fron.NextTick(from, "someone-else | 1h", opts...))(t)

			require.Equal(t, plain, typed)
		})

		t.Run("should keep conditions when seeding", func(t *testing.T) {
			// 2026-10-17 is a saturday.
			got := testx.MustT(fron.NextTick(time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC), "1h @ on weekdays", fron.Options().IgnoreSeed().AutoSeed("row-a")...))(t)
			require.Equal(t, time.Monday, got.Weekday())
		})
	})

	t.Run("window snap", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		t.Run("should smear a daily schedule within the window by seed", func(t *testing.T) {
			seen := make(map[time.Time]struct{})
			for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
				opts := fron.Options().IgnoreSeed().AutoSeed(id)
				got := testx.MustT(fron.NextTick(from, "1d @ during 09:00-17:00", opts...))(t)
				require.False(t, got.Before(time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)), id)
				require.Less(t, got, time.Date(2026, 10, 15, 17, 0, 0, 0, time.UTC), id)
				require.Equal(t, got, testx.MustT(fron.NextTick(from, "1d @ during 09:00-17:00", opts...))(t), id)
				seen[got] = struct{}{}
			}
			require.Greater(t, len(seen), 1)
		})

		t.Run("should keep a daily tick that already falls inside the window", func(t *testing.T) {
			got := testx.MustT(fron.NextTick(from, "1d @ 00:00 during 00:00-10:00"))(t)
			require.Equal(t, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), got)
		})

		t.Run("should render the same time the cron form fires", func(t *testing.T) {
			tree := testx.MustT(fron.Parse("1d @ 12:00 during 09:00-17:00"))(t)
			require.Equal(t, "0 12 * * *", testx.MustT(fron.CRON(tree))(t))
			require.Equal(t, time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC), testx.MustT(fron.NextTick(from, "1d @ 12:00 during 09:00-17:00"))(t))
		})
	})

	t.Run("fallback seed", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		t.Run("should give each fallback seed its own stable slot", func(t *testing.T) {
			a := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h"))(t)
			b := testx.MustT(fron.Seeded("row-b").NextTick(from, "1h"))(t)
			require.NotEqual(t, a, b)
			require.Equal(t, a, testx.MustT(fron.Seeded("row-a").NextTick(from, "1h"))(t))
		})

		t.Run("should not depend on the text of the schedule", func(t *testing.T) {
			// two different strings with the same fallback land on the same offset.
			a := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h"))(t)
			b := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h @ on everyday"))(t)
			require.Equal(t, a, b)
		})

		t.Run("should pick a random slot for an empty fallback seed", func(t *testing.T) {
			seen := make(map[time.Time]struct{})
			for range 40 {
				got := testx.MustT(fron.NextTick(from, "1h"))(t)
				require.True(t, got.After(from))
				require.False(t, got.After(from.Add(time.Hour)))
				seen[got] = struct{}{}
			}
			require.Greater(t, len(seen), 1)
		})

		t.Run("should let a typed seed beat the fallback", func(t *testing.T) {
			typed := testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h"))(t)
			require.Equal(t, typed, testx.MustT(fron.Seeded("row-b").NextTick(from, "someone | 1h"))(t))
		})

		t.Run("should let a pin beat both the typed seed and the fallback", func(t *testing.T) {
			want := time.Date(2026, 10, 15, 0, 7, 30, 0, time.UTC)
			require.Equal(t, want, testx.MustT(fron.Seeded("row-a").NextTick(from, "1h @ 00:07:30"))(t))
			require.Equal(t, want, testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h @ 00:07:30"))(t))
		})

		t.Run("should let IgnoreSeed hand the schedule to the fallback", func(t *testing.T) {
			ignore := fron.Options().IgnoreSeed()
			plain := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h"))(t)
			require.Equal(t, plain, testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h", ignore...))(t))
		})
	})

	t.Run("pinned time", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 8, 20, 0, 0, time.UTC)

		t.Run("should tick at the same time past every hour", func(t *testing.T) {
			in := "1h @ 00:07:30"
			first := testx.MustT(fron.NextTick(from, in))(t)
			second := testx.MustT(fron.NextTick(first, in))(t)
			require.Equal(t, time.Date(2026, 10, 15, 9, 7, 30, 0, time.UTC), first)
			require.Equal(t, time.Date(2026, 10, 15, 10, 7, 30, 0, time.UTC), second)
		})

		t.Run("should repeat within the day for a shorter frequency", func(t *testing.T) {
			in := "15m @ 00:02:30"
			first := testx.MustT(fron.NextTick(from, in))(t)
			second := testx.MustT(fron.NextTick(first, in))(t)
			require.Equal(t, time.Date(2026, 10, 15, 8, 32, 30, 0, time.UTC), first)
			require.Equal(t, time.Date(2026, 10, 15, 8, 47, 30, 0, time.UTC), second)
		})

		t.Run("should pin a daily time to the millisecond", func(t *testing.T) {
			got := testx.MustT(fron.NextTick(from, "1d @ 09:07:30.250"))(t)
			require.Equal(t, time.Date(2026, 10, 15, 9, 7, 30, 250_000_000, time.UTC), got)
		})

		t.Run("should pin a daily time on weekdays", func(t *testing.T) {
			// 2026-10-16 is a friday.
			got := testx.MustT(fron.NextTick(time.Date(2026, 10, 16, 7, 30, 0, 0, time.UTC), "1d @ 07:30 on weekdays"))(t)
			require.Equal(t, time.Date(2026, 10, 19, 7, 30, 0, 0, time.UTC), got)
		})

		t.Run("should keep the period boundary for a pin of zero rather than randomizing it", func(t *testing.T) {
			got := testx.MustT(fron.NextTick(from, "1h @ 00:00"))(t)
			require.Equal(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC), got)
		})

		t.Run("should be a wall clock time in the schedule's timezone", func(t *testing.T) {
			in := "1d @ 07:30 in America/New_York"
			// dst ends 2026-11-01: 07:30 is 11:30 utc before and 12:30 utc after.
			before := testx.MustT(fron.NextTick(time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC), in))(t)
			after := testx.MustT(fron.NextTick(before, in))(t)
			require.Equal(t, time.Date(2026, 10, 31, 11, 30, 0, 0, time.UTC), before)
			require.Equal(t, time.Date(2026, 11, 1, 12, 30, 0, 0, time.UTC), after)
		})

		t.Run("should combine with a window", func(t *testing.T) {
			got := testx.MustT(fron.NextTick(from, "1h @ 00:07:30 during 09:00-17:00"))(t)
			require.Equal(t, time.Date(2026, 10, 15, 9, 7, 30, 0, time.UTC), got)

			got = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 16, 7, 30, 0, time.UTC), "1h @ 00:07:30 during 09:00-17:00"))(t)
			require.Equal(t, time.Date(2026, 10, 16, 9, 7, 30, 0, time.UTC), got)
		})

		t.Run("should keep a cron aligned window with a pin of zero", func(t *testing.T) {
			got := testx.MustT(fron.NextTick(from, "15m @ 00:00 during 09:00-17:00"))(t)
			require.Equal(t, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC), got)
		})

		t.Run("should not be changed by IgnoreSeed", func(t *testing.T) {
			want := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h @ 00:07:30"))(t)
			got := testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h @ 00:07:30", fron.Options().IgnoreSeed()...))(t)
			require.Equal(t, want, got)
		})

		t.Run("should be dropped by IgnorePin so callers keep control of the slot", func(t *testing.T) {
			seeded := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h"))(t)
			ignored := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h @ 00:07:30", fron.Options().IgnorePin()...))(t)
			require.Equal(t, seeded, ignored)
		})

		t.Run("should leave a typed seed in effect when only IgnorePin is used", func(t *testing.T) {
			seeded := testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h"))(t)
			ignored := testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h @ 00:07:30", fron.Options().IgnorePin()...))(t)
			require.Equal(t, seeded, ignored)
		})

		t.Run("should leave only the fallback seed when both are ignored", func(t *testing.T) {
			seeded := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h"))(t)
			ignored := testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h @ 00:07:30", fron.Options().IgnoreSeed().IgnorePin()...))(t)
			require.Equal(t, seeded, ignored)
		})
	})
}
