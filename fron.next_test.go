package fron_test

import (
	"testing"
	"time"

	"github.com/egdaemon/fron"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestNextTick(t *testing.T) {
	const zero = "fron.zero.offset | "

	t.Run("should be strictly after the reference time by default", func(t *testing.T) {
		for from, expected := range map[time.Time]time.Time{
			time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC):  time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC),
			time.Date(2026, 10, 15, 8, 14, 0, 0, time.UTC): time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC),
			time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC): time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC),
		} {
			actual := testx.MustT(fron.NextTick(from, zero+"15m"))(t)
			require.Equal(t, expected, actual, from)
		}
	})

	t.Run("should accept the reference time when inclusive", func(t *testing.T) {
		opts := fron.Options().Inclusive()

		actual := testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), zero+"15m", opts...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 15, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 8, 16, 0, 0, time.UTC), zero+"15m", opts...))(t)
		require.Equal(t, time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC), actual)
	})

	t.Run("should honour conditions", func(t *testing.T) {
		// 2026-10-16 is a friday.
		actual := testx.MustT(fron.NextTick(time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC), zero+"1h @ on weekdays during 09:00-17:00"))(t)
		require.Equal(t, time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), zero+"1d @ on day 1,15"))(t)
		require.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), actual)

		actual = testx.MustT(fron.NextTick(time.Date(2026, 10, 15, 23, 0, 0, 0, time.UTC), zero+"1h @ during 22:00-02:00"))(t)
		require.Equal(t, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), actual)
	})

	t.Run("should handle daylight saving transitions", func(t *testing.T) {
		// dst ended 2026-11-01 in new york, 09:00 local moves from 13:00 to 14:00 utc.
		actual := testx.MustT(fron.NextTick(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), zero+"1h @ during 09:00-10:00 in America/New_York"))(t)
		require.Equal(t, time.Date(2026, 11, 1, 14, 0, 0, 0, time.UTC), actual)
	})

	t.Run("should return the result in the reference location", func(t *testing.T) {
		loc := time.FixedZone("plus2", 2*60*60)
		from := time.Date(2026, 10, 15, 8, 0, 0, 0, loc)
		actual := testx.MustT(fron.NextTick(from, zero+"15m"))(t)
		require.Equal(t, loc, actual.Location())
		require.True(t, from.Add(15*time.Minute).Equal(actual))
	})

	t.Run("should be deterministic for seeded and unseeded input", func(t *testing.T) {
		from := time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)
		for _, in := range []string{"billing-job | 2h", "30m"} {
			require.Equal(t, testx.MustT(fron.NextTick(from, in))(t), testx.MustT(fron.NextTick(from, in))(t), in)
		}
	})

	t.Run("should fail for invalid input", func(t *testing.T) {
		for _, in := range []string{"", "1h @ on 2nd wed", "*/15 * * * *", "1h @ in Mars/Olympus"} {
			_, err := fron.NextTick(time.Now(), in)
			require.Error(t, err, in)
		}
	})

	t.Run("should fail when no tick exists", func(t *testing.T) {
		// 28 day multiples from the epoch always land on a thursday.
		_, err := fron.NextTick(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), zero+"28d @ on mon")
		require.ErrorIs(t, err, fron.ErrNoTick)
	})
}
