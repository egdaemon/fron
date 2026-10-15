package fron_test

import (
	"testing"
	"time"

	"github.com/egdaemon/fron"
	"github.com/egdaemon/fron/ast"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

// zero is the reserved seed selecting cron-equivalent mode (offset 0).
const zero = "fron.zero.offset | "

func TestCRON(t *testing.T) {
	t.Run("should render minute frequencies", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "1m":  "* * * * *",
			zero + "2m":  "*/2 * * * *",
			zero + "3m":  "*/3 * * * *",
			zero + "4m":  "*/4 * * * *",
			zero + "5m":  "*/5 * * * *",
			zero + "6m":  "*/6 * * * *",
			zero + "7m":  "*/7 * * * *",
			zero + "10m": "*/10 * * * *",
			zero + "12m": "*/12 * * * *",
			zero + "13m": "*/13 * * * *",
			zero + "15m": "*/15 * * * *",
			zero + "17m": "*/17 * * * *",
			zero + "20m": "*/20 * * * *",
			zero + "30m": "*/30 * * * *",
			zero + "45m": "*/45 * * * *",
			zero + "59m": "*/59 * * * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render hour frequencies", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "60m":  "0 * * * *",
			zero + "61m":  "0 * * * *",
			zero + "90m":  "0 * * * *",
			zero + "119m": "0 * * * *",
			zero + "120m": "0 */2 * * *",
			zero + "1h":   "0 * * * *",
			zero + "2h":   "0 */2 * * *",
			zero + "3h":   "0 */3 * * *",
			zero + "4h":   "0 */4 * * *",
			zero + "5h":   "0 */5 * * *",
			zero + "6h":   "0 */6 * * *",
			zero + "8h":   "0 */8 * * *",
			zero + "12h":  "0 */12 * * *",
			zero + "23h":  "0 */23 * * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render day frequencies", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "24h": "0 0 * * *",
			zero + "48h": "0 0 */2 * *",
			zero + "1d":  "0 0 * * *",
			zero + "2d":  "0 0 */2 * *",
			zero + "3d":  "0 0 */3 * *",
			zero + "5d":  "0 0 */5 * *",
			zero + "7d":  "0 0 */7 * *",
			zero + "10d": "0 0 */10 * *",
			zero + "14d": "0 0 */14 * *",
			zero + "15d": "0 0 */15 * *",
			zero + "30d": "0 0 */30 * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should reduce compound frequencies", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "1h30m":    "0 * * * *",
			zero + "2h30m":    "0 */2 * * *",
			zero + "1d12h":    "0 0 * * *",
			zero + "2d12h":    "0 0 */2 * *",
			zero + "5d1h5m3s": "0 0 */5 * *",
			zero + "15m30s":   "*/15 * * * *",
			zero + "1m30s":    "* * * * *",
			zero + "90s":      "* * * * *",
			zero + "30s":      "* * * * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render days of the week", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "15m @ on " + "weekdays":                    "*/15 * * * 1-5",
			zero + "15m @ on " + "weekends":                    "*/15 * * * 0,6",
			zero + "15m @ on " + "everyday":                    "*/15 * * * *",
			zero + "15m @ on " + "mon":                         "*/15 * * * 1",
			zero + "15m @ on " + "tue":                         "*/15 * * * 2",
			zero + "15m @ on " + "wed":                         "*/15 * * * 3",
			zero + "15m @ on " + "thu":                         "*/15 * * * 4",
			zero + "15m @ on " + "fri":                         "*/15 * * * 5",
			zero + "15m @ on " + "sat":                         "*/15 * * * 6",
			zero + "15m @ on " + "sun":                         "*/15 * * * 0",
			zero + "15m @ on " + "mon,wed,fri":                 "*/15 * * * 1,3,5",
			zero + "15m @ on " + "tue,thu":                     "*/15 * * * 2,4",
			zero + "15m @ on " + "sat,sun":                     "*/15 * * * 6,0",
			zero + "15m @ on " + "mon,tue,wed":                 "*/15 * * * 1-3",
			zero + "15m @ on " + "mon,tue":                     "*/15 * * * 1,2",
			zero + "15m @ on " + "sun,mon,tue,wed,thu,fri,sat": "*/15 * * * *",
			zero + "15m @ on " + "mon-fri":                     "*/15 * * * 1-5",
			zero + "15m @ on " + "mon-wed":                     "*/15 * * * 1-3",
			zero + "15m @ on " + "tue-thu":                     "*/15 * * * 2-4",
			zero + "15m @ on " + "sat-sun":                     "*/15 * * * 6,0",
			zero + "15m @ on " + "fri-mon":                     "*/15 * * * 5,6,0,1",
			zero + "15m @ on " + "sat-mon":                     "*/15 * * * 6,0,1",
			zero + "15m @ on " + "sun-sat":                     "*/15 * * * *",
			zero + "15m @ on " + "mon-sun":                     "*/15 * * * *",
			zero + "15m @ on " + "wed-wed":                     "*/15 * * * 3",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should combine days with other frequencies", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "1m @ on weekdays":    "* * * * 1-5",
			zero + "5m @ on everyday":    "*/5 * * * *",
			zero + "30m @ on sat":        "*/30 * * * 6",
			zero + "1h @ on weekdays":    "0 * * * 1-5",
			zero + "2h @ on weekends":    "0 */2 * * 0,6",
			zero + "1d @ on mon":         "0 0 * * 1",
			zero + "1d @ on mon,wed,fri": "0 0 * * 1,3,5",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render days of the month", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "1d @ on day 1":           "0 0 1 * *",
			zero + "1d @ on day 15":          "0 0 15 * *",
			zero + "1d @ on day 31":          "0 0 31 * *",
			zero + "1d @ on day 1,15":        "0 0 1,15 * *",
			zero + "1d @ on day 15,1":        "0 0 1,15 * *",
			zero + "1d @ on day 1,10,20,28":  "0 0 1,10,20,28 * *",
			zero + "1d @ on day 1,2,3":       "0 0 1-3 * *",
			zero + "1d @ on day 1,2":         "0 0 1,2 * *",
			zero + "12h @ on day 1":          "0 */12 1 * *",
			zero + "15m @ on day 1,15":       "*/15 * 1,15 * *",
			zero + "1h @ on day 28,29,30,31": "0 * 28-31 * *",
			zero + "7d @ on day 1":           "0 0 1 * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render time windows", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "15m @ during 09:00-17:00": "*/15 9-16 * * *",
			zero + "5m @ during 09:00-17:00":  "*/5 9-16 * * *",
			zero + "1m @ during 00:00-06:00":  "* 0-5 * * *",
			zero + "15m @ during 22:00-02:00": "*/15 22-23,0-1 * * *",
			zero + "20m @ during 18:00-00:00": "*/20 18-23 * * *",
			zero + "15m @ during 09:30-17:00": "*/15 9-16 * * *",
			zero + "15m @ during 09:00-17:30": "*/15 9-17 * * *",
			zero + "15m @ during 09:30-17:30": "*/15 9-17 * * *",
			zero + "10m @ during 00:00-24:00": "*/10 * * * *",
			zero + "10m @ during 12:00-13:00": "*/10 12 * * *",
			zero + "30m @ during 00:00-12:00": "*/30 0-11 * * *",
			zero + "30m @ during 12:00-24:00": "*/30 12-23 * * *",
			zero + "5m @ during 23:00-23:30":  "*/5 23 * * *",
			zero + "15m @ during 01:00-03:00": "*/15 1-2 * * *",
			zero + "45m @ during 08:00-18:00": "*/45 8-17 * * *",
			zero + "1h @ during 09:00-17:00":  "0 9-16 * * *",
			zero + "1h @ during 22:00-02:00":  "0 22-23,0-1 * * *",
			zero + "1h @ during 00:00-01:00":  "0 0 * * *",
			zero + "2h @ during 09:00-17:00":  "0 9,11,13,15 * * *",
			zero + "3h @ during 08:00-20:00":  "0 8,11,14,17 * * *",
			zero + "2h @ during 22:00-02:00":  "0 22,0 * * *",
			zero + "6h @ during 00:00-24:00":  "0 */6 * * *",
			zero + "1d @ during 09:00-17:00":  "0 9 * * *",
			zero + "1d @ during 22:00-02:00":  "0 22 * * *",
			zero + "3d @ during 06:00-12:00":  "0 6 */3 * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render timezones", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "15m @ in UTC":                 "CRON_TZ=UTC */15 * * * *",
			zero + "15m @ in America/New_York":    "CRON_TZ=America/New_York */15 * * * *",
			zero + "15m @ in Europe/London":       "CRON_TZ=Europe/London */15 * * * *",
			zero + "15m @ in Asia/Tokyo":          "CRON_TZ=Asia/Tokyo */15 * * * *",
			zero + "15m @ in Australia/Sydney":    "CRON_TZ=Australia/Sydney */15 * * * *",
			zero + "15m @ in America/Los_Angeles": "CRON_TZ=America/Los_Angeles */15 * * * *",
			zero + "15m @ in Europe/Berlin":       "CRON_TZ=Europe/Berlin */15 * * * *",
			zero + "15m @ in Asia/Kolkata":        "CRON_TZ=Asia/Kolkata */15 * * * *",
			zero + "1h @ in UTC":                  "CRON_TZ=UTC 0 * * * *",
			zero + "1d @ in America/Chicago":      "CRON_TZ=America/Chicago 0 0 * * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should accept conditions in any order", func(t *testing.T) {
		const expected = "CRON_TZ=America/New_York */15 9-16 * * 1-5"
		for _, in := range []string{
			"on weekdays during 09:00-17:00 in America/New_York",
			"in America/New_York on weekdays during 09:00-17:00",
			"during 09:00-17:00 on weekdays in America/New_York",
			"during 09:00-17:00 in America/New_York on weekdays",
			"on weekdays in America/New_York during 09:00-17:00",
			"in America/New_York during 09:00-17:00 on weekdays",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(zero + "15m @ " + in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should combine conditions", func(t *testing.T) {
		for in, expected := range map[string]string{
			zero + "1h @ on mon,wed,fri during 08:00-18:00 in Europe/London":     "CRON_TZ=Europe/London 0 8-17 * * 1,3,5",
			zero + "30m @ on weekends during 10:00-22:00 in Asia/Tokyo":          "CRON_TZ=Asia/Tokyo */30 10-21 * * 0,6",
			zero + "1d @ on day 1 during 03:00-04:00 in UTC":                     "CRON_TZ=UTC 0 3 1 * *",
			zero + "1d @ on day 1,15 during 06:00-12:00":                         "0 6 1,15 * *",
			zero + "2h @ on weekdays during 09:00-17:00":                         "0 9,11,13,15 * * 1-5",
			zero + "5m @ on weekdays during 22:00-02:00":                         "*/5 22-23,0-1 * * 1-5",
			zero + "10m @ on sat during 00:00-12:00":                             "*/10 0-11 * * 6",
			zero + "20m @ on mon-fri in UTC":                                     "CRON_TZ=UTC */20 * * * 1-5",
			zero + "4h @ on mon-fri during 00:00-24:00":                          "0 */4 * * 1-5",
			zero + "15m @ on everyday during 09:00-17:00":                        "*/15 9-16 * * *",
			zero + "30m @ on fri-mon during 08:00-20:00":                         "*/30 8-19 * * 5,6,0,1",
			zero + "3d @ on mon in UTC":                                          "CRON_TZ=UTC 0 0 */3 * 1",
			zero + "12h @ on weekends in Asia/Kolkata":                           "CRON_TZ=Asia/Kolkata 0 */12 * * 0,6",
			zero + "1h @ on day 1 in Europe/Berlin":                              "CRON_TZ=Europe/Berlin 0 * 1 * *",
			zero + "45m @ on weekdays during 07:00-19:00 in America/Los_Angeles": "CRON_TZ=America/Los_Angeles */45 7-18 * * 1-5",
			zero + "15m @ on tue,thu during 13:00-15:00":                         "*/15 13-14 * * 2,4",
			zero + "1h @ during 09:00-17:00 in Australia/Sydney":                 "CRON_TZ=Australia/Sydney 0 9-16 * * *",
			zero + "6h @ on weekdays during 06:00-24:00":                         "0 6,12,18 * * 1-5",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should tolerate whitespace variations", func(t *testing.T) {
		for in, expected := range map[string]string{
			"fron.zero.offset|15m":                                "*/15 * * * *",
			"fron.zero.offset |15m":                               "*/15 * * * *",
			"  fron.zero.offset  |  15m  ":                        "*/15 * * * *",
			"fron.zero.offset | 15m@ on weekdays":                 "*/15 * * * 1-5",
			"fron.zero.offset | 15m @on weekdays":                 "*/15 * * * 1-5",
			"fron.zero.offset|15m@on weekdays during 09:00-17:00": "*/15 9-16 * * 1-5",
			"fron.zero.offset | 15m @\ton\tweekdays":              "*/15 * * * 1-5",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render the aligned form when the seed is omitted", func(t *testing.T) {
		for in, expected := range map[string]string{
			"15m":                     "*/15 * * * *",
			"1d":                      "0 0 * * *",
			"1h @ on weekdays":        "0 * * * 1-5",
			"30m@ during 09:00-17:00": "*/30 9-16 * * *",
			"2h @ in UTC":             "CRON_TZ=UTC 0 */2 * * *",
			"15m @ on day 1":          "*/15 * 1 * *",
			"billing-job | 2h":        "0 */2 * * *",
			"auth-cleaner | 15m @ on weekdays during 09:00-17:00 in America/New_York": "CRON_TZ=America/New_York */15 9-16 * * 1-5",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should apply extra options after the tree", func(t *testing.T) {
		a, err := ast.Parse(zero + "15m")
		require.NoError(t, err)

		actual, err := ast.CRON(a, fron.Options().Weekdays().During(9*time.Hour, 17*time.Hour)...)
		require.NoError(t, err)
		require.Equal(t, "*/15 9-16 * * 1-5", actual)
	})

	t.Run("should let extra options override the tree", func(t *testing.T) {
		a, err := ast.Parse(zero + "15m @ on weekdays in UTC")
		require.NoError(t, err)

		actual, err := ast.CRON(a, fron.Options().Weekends().InTzName("Asia/Tokyo")...)
		require.NoError(t, err)
		require.Equal(t, "CRON_TZ=Asia/Tokyo */15 * * * 0,6", actual)
	})

	t.Run("should fail for unknown timezones", func(t *testing.T) {
		a, err := ast.Parse(zero + "15m @ in Mars/Olympus")
		require.NoError(t, err)

		_, err = ast.CRON(a)
		require.Error(t, err)
	})
}

func TestCRONSeeds(t *testing.T) {
	t.Run("should derive a deterministic offset within the frequency", func(t *testing.T) {
		for _, in := range []string{"billing-job | 2h", "auth-cleaner | 15m", "x | 1m", "30m"} {
			a, err := ast.Parse(in)
			require.NoError(t, err, in)

			first, err := ast.Compile(a)
			require.NoError(t, err, in)
			second, err := ast.Compile(a)
			require.NoError(t, err, in)

			s1, s2 := first.(interface{ Spec() fron.Spec }).Spec(), second.(interface{ Spec() fron.Spec }).Spec()
			require.Equal(t, s1.Offset, s2.Offset, in)
			require.GreaterOrEqual(t, s1.Offset, time.Duration(0), in)
			require.Less(t, s1.Offset, s1.Hz, in)
		}
	})

	t.Run("should yield a zero offset for the reserved seed", func(t *testing.T) {
		a, err := ast.Parse("fron.zero.offset | 15m")
		require.NoError(t, err)

		s, err := ast.Compile(a)
		require.NoError(t, err)
		require.Equal(t, time.Duration(0), s.(interface{ Spec() fron.Spec }).Spec().Offset)
	})

	t.Run("should spread different seeds", func(t *testing.T) {
		offsets := make(map[time.Duration]struct{})
		for _, seed := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
			a, err := ast.Parse(seed + " | 1h")
			require.NoError(t, err)
			s, err := ast.Compile(a)
			require.NoError(t, err)
			offsets[s.(interface{ Spec() fron.Spec }).Spec().Offset] = struct{}{}
		}
		require.Greater(t, len(offsets), 1)
	})
}

func TestParseRejects(t *testing.T) {
	for name, in := range map[string]string{
		"nth weekday":                zero + "1h @ on 2nd wed",
		"nth weekday spelled out":    zero + "1h @ on 2nd wednesday",
		"last weekday":               zero + "1h @ on last fri",
		"last day":                   zero + "1d @ on last day",
		"last day of month":          zero + "1d @ on day L",
		"nearest weekday":            zero + "1d @ on day 15W",
		"every nth weekday":          zero + "1h @ on every 3rd monday",
		"macro daily":                "@daily",
		"macro yearly":               "@yearly",
		"five field cron":            "*/15 * * * *",
		"six field cron":             "0 */15 * * * *",
		"seven field cron":           "0 */15 * * * * 2026",
		"cron numbers":               "0 0 1 1 1",
		"day of month zero":          zero + "1d @ on day 0",
		"day of month 32":            zero + "1d @ on day 32",
		"day of month text":          zero + "1d @ on day first",
		"day of month trailing":      zero + "1d @ on day 1,",
		"missing day of month":       zero + "1d @ on day",
		"time out of range":          zero + "1h @ during 25:00-26:00",
		"minute out of range":        zero + "1h @ during 09:60-17:00",
		"missing end time":           zero + "1h @ during 09:00",
		"empty time range":           zero + "1h @ during 09:00-09:00",
		"start at 24":                zero + "1h @ during 24:00-12:00",
		"time without colon":         zero + "1h @ during 0900-1700",
		"duplicate on":               zero + "1h @ on weekdays on weekends",
		"duplicate during":           zero + "1h @ during 09:00-10:00 during 11:00-12:00",
		"duplicate in":               zero + "1h @ in UTC in UTC",
		"unknown unit":               zero + "1w",
		"unit without number":        zero + "h",
		"number without unit":        zero + "15",
		"zero frequency":             zero + "0m",
		"empty frequency":            "",
		"blank frequency":            "   ",
		"missing frequency":          "billing-job |",
		"missing frequency before @": "billing-job | @ on weekdays",
		"missing seed":               "| 2h",
		"double pipe":                "a | b | 2h",
		"unknown keyword":            zero + "1h @ at weekdays",
		"missing condition":          zero + "1h @",
		"incomplete condition":       zero + "1h @ on",
		"unknown day":                zero + "1h @ on funday",
		"day range with list":        zero + "1h @ on mon-fri,sat",
		"uppercase day":              zero + "1h @ on MON",
		"double at":                  zero + "1h @ on weekdays @ in UTC",
		"trailing text":              zero + "1h extra",
		"frequency overflow":         zero + "99999999999999999999d",
	} {
		t.Run("should reject "+name, func(t *testing.T) {
			_, err := ast.Parse(in)
			require.Error(t, err, in)
		})
	}

	t.Run("should report the offset of the error", func(t *testing.T) {
		_, err := ast.Parse("1h @ on day 32")
		require.ErrorContains(t, err, "ast: offset 12:")
	})
}
