package fron_test

import (
	"testing"
	"time"

	"github.com/egdaemon/fron"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

type seedname string

func offsetOf(t *testing.T, s fron.Schedule) time.Duration {
	t.Helper()
	return s.(interface{ Spec() fron.Spec }).Spec().Offset
}

func TestSeeded(t *testing.T) {
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	t.Run("should accept strings, byte slices and named string types alike", func(t *testing.T) {
		want := offsetOf(t, testx.MustT(fron.Seeded("foo").New(time.Hour))(t))
		require.Equal(t, want, offsetOf(t, testx.MustT(fron.Seeded([]byte("foo")).New(time.Hour))(t)))
		require.Equal(t, want, offsetOf(t, testx.MustT(fron.Seeded(seedname("foo")).New(time.Hour))(t)))
	})

	t.Run("should derive the offset like Seed within the frequency", func(t *testing.T) {
		got := offsetOf(t, testx.MustT(fron.Seeded("foo").New(time.Hour))(t))
		require.Equal(t, fron.Seed("foo")%time.Hour, got)
	})

	t.Run("should be stable per seed and differ across seeds", func(t *testing.T) {
		a := offsetOf(t, testx.MustT(fron.Seeded("a").New(time.Hour))(t))
		require.Equal(t, a, offsetOf(t, testx.MustT(fron.Seeded("a").New(time.Hour))(t)))
		require.NotEqual(t, a, offsetOf(t, testx.MustT(fron.Seeded("b").New(time.Hour))(t)))
	})

	t.Run("should be random for an empty seed, the zero value and the package functions", func(t *testing.T) {
		for name, next := range map[string]func() time.Time{
			"empty":   func() time.Time { return testx.MustT(fron.Seeded("").NextTick(from, "1h"))(t) },
			"zero":    func() time.Time { return testx.MustT(fron.Seeder{}.NextTick(from, "1h"))(t) },
			"package": func() time.Time { return testx.MustT(fron.NextTick(from, "1h"))(t) },
		} {
			seen := make(map[time.Time]struct{})
			for range 40 {
				got := next()
				require.True(t, got.After(from), name)
				require.False(t, got.After(from.Add(time.Hour)), name)
				seen[got] = struct{}{}
			}
			require.Greater(t, len(seen), 1, name)
		}
	})

	t.Run("should agree between Compile and NextTick", func(t *testing.T) {
		seeder := fron.Seeded("row-a")
		tree := testx.MustT(fron.Parse("1h @ on weekdays"))(t)
		sched := testx.MustT(seeder.Compile(tree))(t)
		require.Equal(t, testx.MustT(sched.Next(from))(t), testx.MustT(seeder.NextTick(from, "1h @ on weekdays"))(t))
	})

	t.Run("should not depend on the text of the schedule", func(t *testing.T) {
		seeder := fron.Seeded("row-a")
		a := testx.MustT(seeder.NextTick(from, "1h"))(t)
		b := testx.MustT(seeder.NextTick(from, "1h @ on everyday"))(t)
		require.Equal(t, a, b)
	})

	t.Run("should let a typed seed beat the seeder", func(t *testing.T) {
		typed := testx.MustT(fron.Seeded("row-a").NextTick(from, "someone | 1h"))(t)
		require.Equal(t, typed, testx.MustT(fron.Seeded("row-b").NextTick(from, "someone | 1h"))(t))
	})

	t.Run("should hand the slot to the seeder when the typed seed is ignored", func(t *testing.T) {
		seeder := fron.Seeded("row-a")
		plain := testx.MustT(seeder.NextTick(from, "1h"))(t)
		ignored := testx.MustT(seeder.NextTick(from, "someone | 1h", fron.Options().IgnoreSeed()...))(t)
		require.Equal(t, plain, ignored)
	})

	t.Run("should keep a pin exactly where it was written", func(t *testing.T) {
		got := testx.MustT(fron.Seeded("row-a").NextTick(from, "1h @ 00:07:30"))(t)
		require.Equal(t, time.Date(2026, 10, 15, 0, 7, 30, 0, time.UTC), got)
	})

	t.Run("should not write to the options slice it is given", func(t *testing.T) {
		opts := make([]fron.Option, 1, 8)
		opts[0] = fron.Options().Weekdays()[0]
		spare := opts[:cap(opts)]

		_, err := fron.Seeded("row-a").NextTick(from, "1h", opts...)
		require.NoError(t, err)
		_, err = fron.Seeded("row-a").New(time.Hour, opts...)
		require.NoError(t, err)

		for i := 1; i < len(spare); i++ {
			require.Nil(t, spare[i], i)
		}
	})
}
