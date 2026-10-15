package ast_test

import (
	"strings"
	"testing"
	"time"

	"github.com/egdaemon/fron/ast"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("should parse a bare frequency", func(t *testing.T) {
		a, err := ast.Parse("30m")
		require.NoError(t, err)
		require.Equal(t, 30*time.Minute, a.Hz)
		require.False(t, a.HasSeed)
		require.Equal(t, time.Duration(-1), a.Pin)
	})

	t.Run("should parse compound frequencies", func(t *testing.T) {
		a, err := ast.Parse("5d1h5m3s")
		require.NoError(t, err)
		require.Equal(t, 5*24*time.Hour+time.Hour+5*time.Minute+3*time.Second, a.Hz)
	})

	t.Run("should parse a seed", func(t *testing.T) {
		a, err := ast.Parse("billing-job | 2h")
		require.NoError(t, err)
		require.True(t, a.HasSeed)
		require.Equal(t, "billing-job", a.Seed)
		require.Equal(t, 2*time.Hour, a.Hz)
	})

	t.Run("should parse a pin as the first clause", func(t *testing.T) {
		for in, expected := range map[string]time.Duration{
			"1h @ 00:07:30":          7*time.Minute + 30*time.Second,
			"1h @ 00:07":             7 * time.Minute,
			"1h @ 00:07:30.5":        7*time.Minute + 30*time.Second + 500*time.Millisecond,
			"1h @ 00:07:30.250":      7*time.Minute + 30*time.Second + 250*time.Millisecond,
			"1h@00:07:30":            7*time.Minute + 30*time.Second,
			"1d @ 07:30 on weekdays": 7*time.Hour + 30*time.Minute,
			"1h @ 00:00":             0,
		} {
			a, err := ast.Parse(in)
			require.NoError(t, err, in)
			require.Equal(t, expected, a.Pin, in)
		}
	})

	t.Run("should keep conditions after a pin", func(t *testing.T) {
		a, err := ast.Parse("1h @ 00:07:30 during 09:00-17:00 in UTC")
		require.NoError(t, err)
		require.True(t, a.HasWindow)
		require.Equal(t, "UTC", a.Tz)
		require.Equal(t, 7*time.Minute+30*time.Second, a.Pin)
	})

	t.Run("should accept seconds and milliseconds in a window", func(t *testing.T) {
		a, err := ast.Parse("15m @ 00:00 during 09:00:30-17:00:00.250")
		require.NoError(t, err)
		require.Equal(t, 9*time.Hour+30*time.Second, a.Start)
		require.Equal(t, 17*time.Hour+250*time.Millisecond, a.End)
	})

	t.Run("should report where a pin never falls inside the window", func(t *testing.T) {
		_, err := ast.Parse("1d @ 07:30 during 09:00-17:00")
		require.ErrorContains(t, err, "time 07:30 never falls inside during 09:00-17:00")
	})

	t.Run("should accept pins that land inside the window", func(t *testing.T) {
		for _, in := range []string{
			"1h @ 00:07:30 during 09:00-17:00",
			"15m @ 00:00 during 09:00-17:00",
			"1d @ 12:00 during 09:00-17:00",
			"15m @ 00:02:30 during 22:00-02:00",
			"1d @ 23:00 during 22:00-02:00",
			"1d @ 01:00 during 22:00-02:00",
			"6h @ 03:00 during 09:00-10:00",
		} {
			_, err := ast.Parse(in)
			require.NoError(t, err, in)
		}
	})

	t.Run("should reject pins that never land inside the window", func(t *testing.T) {
		for _, in := range []string{
			"1d @ 07:30 during 09:00-17:00",
			"6h @ 01:00 during 09:00-10:00",
			"1d @ 12:00 during 22:00-02:00",
			"2d @ 08:59:59 during 09:00-17:00",
			"1d @ 17:00 during 09:00-17:00",
		} {
			_, err := ast.Parse(in)
			require.Error(t, err, in)
		}
	})

	t.Run("should parse every condition", func(t *testing.T) {
		a, err := ast.Parse("auth-cleaner | 15m @ on weekdays during 09:00-17:00 in America/New_York")
		require.NoError(t, err)
		require.True(t, a.HasWindow)
		require.Equal(t, 9*time.Hour, a.Start)
		require.Equal(t, 17*time.Hour, a.End)
		require.Equal(t, "America/New_York", a.Tz)
	})

	t.Run("should parse day lists, ranges, and days of the month", func(t *testing.T) {
		a, err := ast.Parse("1h @ on mon,wed")
		require.NoError(t, err)
		require.Equal(t, []time.Weekday{time.Monday, time.Wednesday}, a.Days)

		a, err = ast.Parse("1h @ on fri-mon")
		require.NoError(t, err)
		require.Equal(t, time.Friday, a.From)
		require.Equal(t, time.Monday, a.To)

		a, err = ast.Parse("1d @ on day 1,15")
		require.NoError(t, err)
		require.Equal(t, []int{1, 15}, a.DaysOfMonth)
	})

	t.Run("should parse a 24:00 window end", func(t *testing.T) {
		a, err := ast.Parse("1h @ during 12:00-24:00")
		require.NoError(t, err)
		require.Equal(t, 24*time.Hour, a.End)
	})

	t.Run("should scale linearly as a single pass", func(t *testing.T) {
		// a long list of days reads the input once, so allocations grow with the
		// number of elements and never with repeated re-scans of the input.
		build := func(n int) string {
			parts := make([]string, n)
			for i := range parts {
				parts[i] = "mon"
			}
			return "1h @ on " + strings.Join(parts, ",")
		}

		allocs := func(n int) float64 {
			in := build(n)
			return testing.AllocsPerRun(10, func() {
				_, err := ast.Parse(in)
				require.NoError(t, err)
			})
		}

		small, large := allocs(100), allocs(1000)
		require.Less(t, large, small*20, "allocations grew faster than the input")
	})
}
