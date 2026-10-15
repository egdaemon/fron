package fron_test

import (
	"strings"
	"testing"
	"time"

	"github.com/egdaemon/fron"
	"github.com/egdaemon/fron/ast"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

// pin00 pins a schedule to the period boundary, which is what cron does: the
// time 00:00 goes first after the @.
func pin00(s string) string {
	if !strings.Contains(s, "@") {
		return s + " @ 00:00"
	}

	return strings.Replace(s, "@ ", "@ 00:00 ", 1)
}

func TestCRON(t *testing.T) {
	t.Run("should render minute frequencies", func(t *testing.T) {
		for in, expected := range map[string]string{
			pin00("1m"):  "* * * * *",
			pin00("2m"):  "*/2 * * * *",
			pin00("3m"):  "*/3 * * * *",
			pin00("4m"):  "*/4 * * * *",
			pin00("5m"):  "*/5 * * * *",
			pin00("6m"):  "*/6 * * * *",
			pin00("7m"):  "*/7 * * * *",
			pin00("10m"): "*/10 * * * *",
			pin00("12m"): "*/12 * * * *",
			pin00("13m"): "*/13 * * * *",
			pin00("15m"): "*/15 * * * *",
			pin00("17m"): "*/17 * * * *",
			pin00("20m"): "*/20 * * * *",
			pin00("30m"): "*/30 * * * *",
			pin00("45m"): "*/45 * * * *",
			pin00("59m"): "*/59 * * * *",
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
			pin00("60m"):  "0 * * * *",
			pin00("61m"):  "0 * * * *",
			pin00("90m"):  "0 * * * *",
			pin00("119m"): "0 * * * *",
			pin00("120m"): "0 */2 * * *",
			pin00("1h"):   "0 * * * *",
			pin00("2h"):   "0 */2 * * *",
			pin00("3h"):   "0 */3 * * *",
			pin00("4h"):   "0 */4 * * *",
			pin00("5h"):   "0 */5 * * *",
			pin00("6h"):   "0 */6 * * *",
			pin00("8h"):   "0 */8 * * *",
			pin00("12h"):  "0 */12 * * *",
			pin00("23h"):  "0 */23 * * *",
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
			pin00("24h"): "0 0 * * *",
			pin00("48h"): "0 0 */2 * *",
			pin00("1d"):  "0 0 * * *",
			pin00("2d"):  "0 0 */2 * *",
			pin00("3d"):  "0 0 */3 * *",
			pin00("5d"):  "0 0 */5 * *",
			pin00("7d"):  "0 0 */7 * *",
			pin00("10d"): "0 0 */10 * *",
			pin00("14d"): "0 0 */14 * *",
			pin00("15d"): "0 0 */15 * *",
			pin00("30d"): "0 0 */30 * *",
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
			pin00("1h30m"):    "0 * * * *",
			pin00("2h30m"):    "0 */2 * * *",
			pin00("1d12h"):    "0 0 * * *",
			pin00("2d12h"):    "0 0 */2 * *",
			pin00("5d1h5m3s"): "0 0 */5 * *",
			pin00("15m30s"):   "*/15 * * * *",
			pin00("1m30s"):    "* * * * *",
			pin00("90s"):      "* * * * *",
			pin00("30s"):      "* * * * *",
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
			pin00("15m @ on " + "weekdays"):                    "*/15 * * * 1-5",
			pin00("15m @ on " + "weekends"):                    "*/15 * * * 0,6",
			pin00("15m @ on " + "everyday"):                    "*/15 * * * *",
			pin00("15m @ on " + "mon"):                         "*/15 * * * 1",
			pin00("15m @ on " + "tue"):                         "*/15 * * * 2",
			pin00("15m @ on " + "wed"):                         "*/15 * * * 3",
			pin00("15m @ on " + "thu"):                         "*/15 * * * 4",
			pin00("15m @ on " + "fri"):                         "*/15 * * * 5",
			pin00("15m @ on " + "sat"):                         "*/15 * * * 6",
			pin00("15m @ on " + "sun"):                         "*/15 * * * 0",
			pin00("15m @ on " + "mon,wed,fri"):                 "*/15 * * * 1,3,5",
			pin00("15m @ on " + "tue,thu"):                     "*/15 * * * 2,4",
			pin00("15m @ on " + "sat,sun"):                     "*/15 * * * 6,0",
			pin00("15m @ on " + "mon,tue,wed"):                 "*/15 * * * 1-3",
			pin00("15m @ on " + "mon,tue"):                     "*/15 * * * 1,2",
			pin00("15m @ on " + "sun,mon,tue,wed,thu,fri,sat"): "*/15 * * * *",
			pin00("15m @ on " + "mon-fri"):                     "*/15 * * * 1-5",
			pin00("15m @ on " + "mon-wed"):                     "*/15 * * * 1-3",
			pin00("15m @ on " + "tue-thu"):                     "*/15 * * * 2-4",
			pin00("15m @ on " + "sat-sun"):                     "*/15 * * * 6,0",
			pin00("15m @ on " + "fri-mon"):                     "*/15 * * * 5,6,0,1",
			pin00("15m @ on " + "sat-mon"):                     "*/15 * * * 6,0,1",
			pin00("15m @ on " + "sun-sat"):                     "*/15 * * * *",
			pin00("15m @ on " + "mon-sun"):                     "*/15 * * * *",
			pin00("15m @ on " + "wed-wed"):                     "*/15 * * * 3",
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
			pin00("1m @ on weekdays"):    "* * * * 1-5",
			pin00("5m @ on everyday"):    "*/5 * * * *",
			pin00("30m @ on sat"):        "*/30 * * * 6",
			pin00("1h @ on weekdays"):    "0 * * * 1-5",
			pin00("2h @ on weekends"):    "0 */2 * * 0,6",
			pin00("1d @ on mon"):         "0 0 * * 1",
			pin00("1d @ on mon,wed,fri"): "0 0 * * 1,3,5",
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
			pin00("1d @ on day 1"):           "0 0 1 * *",
			pin00("1d @ on day 15"):          "0 0 15 * *",
			pin00("1d @ on day 31"):          "0 0 31 * *",
			pin00("1d @ on day 1,15"):        "0 0 1,15 * *",
			pin00("1d @ on day 15,1"):        "0 0 1,15 * *",
			pin00("1d @ on day 1,10,20,28"):  "0 0 1,10,20,28 * *",
			pin00("1d @ on day 1,2,3"):       "0 0 1-3 * *",
			pin00("1d @ on day 1,2"):         "0 0 1,2 * *",
			pin00("12h @ on day 1"):          "0 */12 1 * *",
			pin00("15m @ on day 1,15"):       "*/15 * 1,15 * *",
			pin00("1h @ on day 28,29,30,31"): "0 * 28-31 * *",
			pin00("7d @ on day 1"):           "0 0 1 * *",
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
			pin00("15m @ during 09:00-17:00"): "*/15 9-16 * * *",
			pin00("5m @ during 09:00-17:00"):  "*/5 9-16 * * *",
			pin00("1m @ during 00:00-06:00"):  "* 0-5 * * *",
			pin00("15m @ during 22:00-02:00"): "*/15 22-23,0-1 * * *",
			pin00("20m @ during 18:00-00:00"): "*/20 18-23 * * *",
			pin00("15m @ during 09:30-17:00"): "*/15 9-16 * * *",
			pin00("15m @ during 09:00-17:30"): "*/15 9-17 * * *",
			pin00("15m @ during 09:30-17:30"): "*/15 9-17 * * *",
			pin00("10m @ during 00:00-24:00"): "*/10 * * * *",
			pin00("10m @ during 12:00-13:00"): "*/10 12 * * *",
			pin00("30m @ during 00:00-12:00"): "*/30 0-11 * * *",
			pin00("30m @ during 12:00-24:00"): "*/30 12-23 * * *",
			pin00("5m @ during 23:00-23:30"):  "*/5 23 * * *",
			pin00("15m @ during 01:00-03:00"): "*/15 1-2 * * *",
			pin00("45m @ during 08:00-18:00"): "*/45 8-17 * * *",
			pin00("1h @ during 09:00-17:00"):  "0 9-16 * * *",
			pin00("1h @ during 22:00-02:00"):  "0 22-23,0-1 * * *",
			pin00("1h @ during 00:00-01:00"):  "0 0 * * *",
			pin00("2h @ during 09:00-17:00"):  "0 9,11,13,15 * * *",
			pin00("3h @ during 08:00-20:00"):  "0 8,11,14,17 * * *",
			pin00("2h @ during 22:00-02:00"):  "0 22,0 * * *",
			pin00("6h @ during 00:00-24:00"):  "0 */6 * * *",
			"1d @ 09:00 during 09:00-17:00":   "0 9 * * *",
			"1d @ 22:00 during 22:00-02:00":   "0 22 * * *",
			"3d @ 06:00 during 06:00-12:00":   "0 6 */3 * *",
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
			pin00("15m @ in UTC"):                 "CRON_TZ=UTC */15 * * * *",
			pin00("15m @ in America/New_York"):    "CRON_TZ=America/New_York */15 * * * *",
			pin00("15m @ in Europe/London"):       "CRON_TZ=Europe/London */15 * * * *",
			pin00("15m @ in Asia/Tokyo"):          "CRON_TZ=Asia/Tokyo */15 * * * *",
			pin00("15m @ in Australia/Sydney"):    "CRON_TZ=Australia/Sydney */15 * * * *",
			pin00("15m @ in America/Los_Angeles"): "CRON_TZ=America/Los_Angeles */15 * * * *",
			pin00("15m @ in Europe/Berlin"):       "CRON_TZ=Europe/Berlin */15 * * * *",
			pin00("15m @ in Asia/Kolkata"):        "CRON_TZ=Asia/Kolkata */15 * * * *",
			pin00("1h @ in UTC"):                  "CRON_TZ=UTC 0 * * * *",
			pin00("1d @ in America/Chicago"):      "CRON_TZ=America/Chicago 0 0 * * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should render a time pinned within the frequency", func(t *testing.T) {
		for in, expected := range map[string]string{
			"1h @ 00:02:30":                     "2 * * * *",
			"1h @ 00:00:30":                     "0 * * * *",
			"1h @ 00:07:30.250":                 "7 * * * *",
			"1h @ 00:07:30.5":                   "7 * * * *",
			"15m @ 00:02:30":                    "2-59/15 * * * *",
			"15m @ 00:00:30":                    "*/15 * * * *",
			"2h @ 01:07:30":                     "7 1-23/2 * * *",
			"2h @ 00:07:30":                     "7 */2 * * *",
			"1d @ 09:07:30":                     "7 9 * * *",
			"1d @ 07:30":                        "30 7 * * *",
			"1d @ 09:07:30.250":                 "7 9 * * *",
			"1d @ 00:00:00.001":                 "0 0 * * *",
			"1h @ 00:02 in UTC":                 "CRON_TZ=UTC 2 * * * *",
			"1h @ 00:02 on weekdays":            "2 * * * 1-5",
			"1d @ 07:30 on weekdays":            "30 7 * * 1-5",
			"1d @ 07:30 in America/New_York":    "CRON_TZ=America/New_York 30 7 * * *",
			"1h @ 00:07:30 during 09:00-17:00":  "7 9-16 * * *",
			"15m @ 00:00 during 09:00-17:00":    "*/15 9-16 * * *",
			"15m @ 00:02:30 during 22:00-02:00": "2-59/15 22-23,0-1 * * *",
			"1d @ 12:00 during 09:00-17:00":     "0 12 * * *",
			"6h @ 03:00 during 09:00-10:00":     "0 9 * * *",
		} {
			t.Run(in, func(t *testing.T) {
				a := testx.MustT(ast.Parse(in))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should accept seconds and milliseconds in a window", func(t *testing.T) {
		for in, expected := range map[string]string{
			pin00("15m @ during 09:00:30-17:00:00"):     "*/15 9-16 * * *",
			pin00("15m @ during 09:00:30-17:00:00.500"): "*/15 9-17 * * *",
			pin00("5m @ during 00:00:00.250-06:00:00"):  "*/5 0-5 * * *",
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
				a := testx.MustT(ast.Parse(pin00("15m @ " + in)))(t)
				actual := testx.MustT(ast.CRON(a))(t)
				require.Equal(t, expected, actual, in)
			})
		}
	})

	t.Run("should combine conditions", func(t *testing.T) {
		for in, expected := range map[string]string{
			pin00("1h @ on mon,wed,fri during 08:00-18:00 in Europe/London"):     "CRON_TZ=Europe/London 0 8-17 * * 1,3,5",
			pin00("30m @ on weekends during 10:00-22:00 in Asia/Tokyo"):          "CRON_TZ=Asia/Tokyo */30 10-21 * * 0,6",
			"1d @ 03:00 on day 1 during 03:00-04:00 in UTC":                      "CRON_TZ=UTC 0 3 1 * *",
			"1d @ 06:00 on day 1,15 during 06:00-12:00":                          "0 6 1,15 * *",
			pin00("2h @ on weekdays during 09:00-17:00"):                         "0 9,11,13,15 * * 1-5",
			pin00("5m @ on weekdays during 22:00-02:00"):                         "*/5 22-23,0-1 * * 1-5",
			pin00("10m @ on sat during 00:00-12:00"):                             "*/10 0-11 * * 6",
			pin00("20m @ on mon-fri in UTC"):                                     "CRON_TZ=UTC */20 * * * 1-5",
			pin00("4h @ on mon-fri during 00:00-24:00"):                          "0 */4 * * 1-5",
			pin00("15m @ on everyday during 09:00-17:00"):                        "*/15 9-16 * * *",
			pin00("30m @ on fri-mon during 08:00-20:00"):                         "*/30 8-19 * * 5,6,0,1",
			pin00("3d @ on mon in UTC"):                                          "CRON_TZ=UTC 0 0 */3 * 1",
			pin00("12h @ on weekends in Asia/Kolkata"):                           "CRON_TZ=Asia/Kolkata 0 */12 * * 0,6",
			pin00("1h @ on day 1 in Europe/Berlin"):                              "CRON_TZ=Europe/Berlin 0 * 1 * *",
			pin00("45m @ on weekdays during 07:00-19:00 in America/Los_Angeles"): "CRON_TZ=America/Los_Angeles */45 7-18 * * 1-5",
			pin00("15m @ on tue,thu during 13:00-15:00"):                         "*/15 13-14 * * 2,4",
			pin00("1h @ during 09:00-17:00 in Australia/Sydney"):                 "CRON_TZ=Australia/Sydney 0 9-16 * * *",
			pin00("6h @ on weekdays during 06:00-24:00"):                         "0 6,12,18 * * 1-5",
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
			"15m@00:00":                                "*/15 * * * *",
			"  15m  @  00:00  ":                        "*/15 * * * *",
			"15m@ 00:00 on weekdays":                   "*/15 * * * 1-5",
			"15m @00:00 on weekdays":                   "*/15 * * * 1-5",
			"15m@00:00 on weekdays during 09:00-17:00": "*/15 9-16 * * 1-5",
			"15m @\t00:00\ton\tweekdays":               "*/15 * * * 1-5",
			"billing | 15m @ 00:02:30":                 "2-59/15 * * * *",
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
		a, err := ast.Parse(pin00("15m"))
		require.NoError(t, err)

		actual, err := ast.CRON(a, fron.Options().Weekdays().During(9*time.Hour, 17*time.Hour)...)
		require.NoError(t, err)
		require.Equal(t, "*/15 9-16 * * 1-5", actual)
	})

	t.Run("should let extra options override the tree", func(t *testing.T) {
		a, err := ast.Parse(pin00("15m @ on weekdays in UTC"))
		require.NoError(t, err)

		actual, err := ast.CRON(a, fron.Options().Weekends().InTzName("Asia/Tokyo")...)
		require.NoError(t, err)
		require.Equal(t, "CRON_TZ=Asia/Tokyo */15 * * * 0,6", actual)
	})

	t.Run("should fail for unknown timezones", func(t *testing.T) {
		a, err := ast.Parse(pin00("15m @ in Mars/Olympus"))
		require.NoError(t, err)

		_, err = ast.CRON(a)
		require.Error(t, err)
	})
}

func TestCRONSeeds(t *testing.T) {
	t.Run("should pick a random offset for a schedule without a seed and an empty fallback", func(t *testing.T) {
		a, err := ast.Parse("30m")
		require.NoError(t, err)

		seen := make(map[time.Duration]struct{})
		for range 40 {
			s := testx.MustT(ast.Compile(a))(t).(interface{ Spec() fron.Spec }).Spec()
			require.GreaterOrEqual(t, s.Offset, time.Duration(0))
			require.Less(t, s.Offset, s.Hz)
			seen[s.Offset] = struct{}{}
		}
		require.Greater(t, len(seen), 1)
	})

	t.Run("should derive a deterministic offset within the frequency", func(t *testing.T) {
		// a typed seed, or a seeder for a schedule without one.
		for _, c := range [][2]string{{"billing-job | 2h", ""}, {"auth-cleaner | 15m", ""}, {"x | 1m", ""}, {"30m", "row-1"}} {
			in := c[0]
			a, err := ast.Parse(in)
			require.NoError(t, err, in)

			first, err := fron.Seeded(c[1]).Compile(a)
			require.NoError(t, err, in)
			second, err := fron.Seeded(c[1]).Compile(a)
			require.NoError(t, err, in)

			s1, s2 := first.(interface{ Spec() fron.Spec }).Spec(), second.(interface{ Spec() fron.Spec }).Spec()
			require.Equal(t, s1.Offset, s2.Offset, in)
			require.GreaterOrEqual(t, s1.Offset, time.Duration(0), in)
			require.Less(t, s1.Offset, s1.Hz, in)
		}
	})

	t.Run("should yield a zero offset for a pin of 00:00", func(t *testing.T) {
		a, err := ast.Parse("15m @ 00:00")
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
		"nth weekday":                      "1h @ on 2nd wed",
		"nth weekday spelled out":          "1h @ on 2nd wednesday",
		"last weekday":                     "1h @ on last fri",
		"last day":                         "1d @ on last day",
		"last day of month":                "1d @ on day L",
		"nearest weekday":                  "1d @ on day 15W",
		"every nth weekday":                "1h @ on every 3rd monday",
		"macro daily":                      "@daily",
		"macro yearly":                     "@yearly",
		"five field cron":                  "*/15 * * * *",
		"six field cron":                   "0 */15 * * * *",
		"seven field cron":                 "0 */15 * * * * 2026",
		"cron numbers":                     "0 0 1 1 1",
		"day of month zero":                "1d @ on day 0",
		"day of month 32":                  "1d @ on day 32",
		"day of month text":                "1d @ on day first",
		"day of month trailing":            "1d @ on day 1,",
		"missing day of month":             "1d @ on day",
		"time out of range":                "1h @ during 25:00-26:00",
		"minute out of range":              "1h @ during 09:60-17:00",
		"seconds out of range in a window": "1h @ during 09:00:60-17:00",
		"too many millisecond digits":      "1h @ during 09:00-17:00:30.1234",
		"dangling millisecond point":       "1h @ during 09:00-17:00:30.",
		"milliseconds without seconds":     "1h @ during 09:00-17:00.500",
		"one digit hour":                   "1h @ during 9:00-17:00",
		"single time in during":            "1d @ during 07:30",
		"pin as long as the frequency":     "1h @ 01:00",
		"pin longer than the frequency":    "15m @ 00:15",
		"pin of a day":                     "1d @ 24:00",
		"pin hour out of range":            "1h @ 25:00",
		"pin seconds out of range":         "1h @ 00:07:60",
		"pin milliseconds too long":        "1h @ 00:07:30.1234",
		"pin with one digit hour":          "1h @ 0:07",
		"pin after another clause":         "1d @ on weekdays 07:30",
		"two pins":                         "1d @ 07:30 07:45",
		"pin never inside the window":      "1d @ 07:30 during 09:00-17:00",
		"pin steps over a narrow window":   "6h @ 01:00 during 09:00-10:00",
		"pin before a wrapped window":      "1d @ 12:00 during 22:00-02:00",
		"pin without a frequency":          "@ 00:07:30",
		"missing end time":                 "1h @ during 09:00-",
		"missing start time":               "1h @ during -17:00",
		"empty time range":                 "1h @ during 09:00-09:00",
		"start at 24":                      "1h @ during 24:00-12:00",
		"time without colon":               "1h @ during 0900-1700",
		"duplicate on":                     "1h @ on weekdays on weekends",
		"duplicate during":                 "1h @ during 09:00-10:00 during 11:00-12:00",
		"duplicate in":                     "1h @ in UTC in UTC",
		"unknown unit":                     "1w",
		"unit without number":              "h",
		"number without unit":              "15",
		"zero frequency":                   "0m",
		"empty frequency":                  "",
		"blank frequency":                  "   ",
		"missing frequency":                "billing-job |",
		"missing frequency before @":       "billing-job | @ on weekdays",
		"missing seed":                     "| 2h",
		"double pipe":                      "a | b | 2h",
		"unknown keyword":                  "1h @ at weekdays",
		"missing condition":                "1h @",
		"incomplete condition":             "1h @ on",
		"unknown day":                      "1h @ on funday",
		"day range with list":              "1h @ on mon-fri,sat",
		"uppercase day":                    "1h @ on MON",
		"double at":                        "1h @ on weekdays @ in UTC",
		"trailing text":                    "1h extra",
		"frequency overflow":               "99999999999999999999d",
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
