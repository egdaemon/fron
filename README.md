# fron

A frequency-first scheduler for Go. Instead of cron's calendar grid, a schedule is
an elapsed interval (`15m`, `2h`, `1d`) plus optional conditions. Each job gets a
stable or random slot within its interval, which avoids the thundering herd when
many jobs share the same schedule.

```go
import "github.com/egdaemon/fron"
```

## Quick start

```go
// first tick strictly after now
next, err := fron.NextTick(time.Now(), "1h @ on weekdays during 09:00-17:00 in America/New_York")

// stable slot per row: the same id always lands in the same slot
next, err = fron.Seeded(rowID).NextTick(time.Now(), "30m")

// parse once, compile, then iterate
tree, err := fron.Parse("billing-job | 2h")
sched, err := fron.Compile(tree)
next, err = sched.Next(time.Now())

// code-first, no DSL
sched, err = fron.New(15*time.Minute, 0, fron.Options().Weekdays().During(9*time.Hour, 17*time.Hour)...)
```

`Next` returns `fron.ErrNoTick` if nothing matches within the horizon (128 years by default).

## DSL

```text
[seed |] frequency [@ [pin] [conditions]]
```

```text
30m                                   # random slot, or stable with fron.Seeded(v)
billing-job | 2h                      # seed gives a deterministic slot
15m @ 00:00                           # pin to 00:00: cron-aligned, like */15 * * * *
1h @ 00:07:30                         # 7m30s past every hour
1d @ 07:30 on weekdays                # every weekday at 07:30
auth-cleaner | 15m @ on weekdays during 09:00-17:00 in America/New_York
12h @ on day 1,15
15m @ during 22:00-02:00              # ranges may wrap past midnight
```

| Clause | Example | Option |
| --- | --- | --- |
| frequency | `5d1h5m3s` | `New`'s `hz` argument |
| seed | `billing-job \|` | `Seed(s)` / `AutoSeed(v)` |
| pin | `@ 00:07:30` | `At(d)` |
| days | `on weekdays`, `on mon-fri`, `on mon,wed,fri` | `Weekdays()`, `DaysBetween()`, `Days()` |
| day of month | `on day 1,15` | `DaysOfMonth(...)` |
| window | `during 09:00-17:00` | `During(start, end)` |
| timezone | `in America/New_York` | `InTz(loc)`, `InTzName(name)` |

A pin overrides any seed and is a wall-clock time in the schedule's timezone (UTC by default).

## Cron output

```go
expr, err := fron.CRON(tree) // 5 field cron expression
```

## Packages

| Package | Role |
| --- | --- |
| `fron` | facade; one import is enough |
| `fron/ast` | DSL front end: `Parse`, `Compile`, `CRON` |
| `fron/engine` | scheduling engine: `Schedule`, `New`, `Seed`, options |

## Design

See [.Design.Document.md](.Design.Document.md) for the full grammar, option semantics, window and pin behaviour.

## License

See [LICENSE](LICENSE).
