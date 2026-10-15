package ast

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/egdaemon/fron/engine"
)

type spec interface {
	Spec() engine.Spec
}

// CRON renders the tree as a 5 field cron expression. Extra options are
// applied after the options derived from the tree, as in Compile. Cron is a
// clock grid and cannot express a seed offset so the aligned form is rendered.
func CRON(t T, options ...engine.Option) (string, error) {
	s, err := Compile(t, "", options...)
	if err != nil {
		return "", err
	}

	d, ok := s.(spec)
	if !ok {
		return "", errors.New("ast: schedule does not describe itself")
	}

	return render(d.Spec(), t.Tz != ""), nil
}

func render(s engine.Spec, tz bool) string {
	var (
		minute = "*"
		hour   = "*"
		dom    = "*"
		dow    = "*"
	)

	hours := windowHours(s)
	minutes := int64(s.Hz / time.Minute)

	switch {
	case minutes < 60:
		minute = step(max(minutes, 1))
		if len(hours) > 0 {
			hour = field(hours, 2)
		}
	case minutes < 24*60:
		minute = "0"
		n := minutes / 60
		switch {
		case len(hours) > 0 && n > 1:
			hour = field(every(hours, int(n)), 3)
		case len(hours) > 0:
			hour = field(hours, 2)
		default:
			hour = step(n)
		}
	default:
		minute = "0"
		hour = "0"
		if s.Window {
			hour = strconv.Itoa(int(s.Start / time.Hour))
		}
		dom = step(minutes / (24 * 60))
	}

	if len(s.DaysOfMonth) > 0 {
		dom = field(s.DaysOfMonth, 3)
	}

	if len(s.Days) > 0 && len(s.Days) < 7 {
		days := make([]int, len(s.Days))
		for i, d := range s.Days {
			days[i] = int(d)
		}
		dow = field(days, 3)
	}

	expr := strings.Join([]string{minute, hour, dom, "*", dow}, " ")

	// an omitted timezone is UTC, so no prefix is rendered for it.
	if tz && s.Location != nil {
		return "CRON_TZ=" + s.Location.String() + " " + expr
	}

	return expr
}

// step renders a */n field, n of 1 is every value.
func step(n int64) string {
	if n <= 1 {
		return "*"
	}

	return "*/" + strconv.FormatInt(n, 10)
}

// windowHours lists the whole hours touched by the window in window order,
// a window that wraps midnight lists the late hours first.
func windowHours(s engine.Spec) []int {
	if !s.Window {
		return nil
	}

	first := int(s.Start / time.Hour)
	last := int((s.End+time.Hour-1)/time.Hour) - 1 // inclusive, end bound is exclusive

	var hours []int
	switch {
	case s.Start < s.End:
		for h := first; h <= last; h++ {
			hours = append(hours, h)
		}
	default:
		for h := first; h < 24; h++ {
			hours = append(hours, h)
		}
		for h := 0; h <= last; h++ {
			hours = append(hours, h)
		}
	}

	if len(hours) >= 24 {
		return nil
	}

	return hours
}

// every keeps each nth hour counting from the start of the window.
func every(hours []int, n int) []int {
	var out []int
	for i := 0; i < len(hours); i += n {
		out = append(out, hours[i])
	}

	return out
}

// field renders values in order, collapsing ascending runs of at least min into ranges.
func field(values []int, min int) string {
	var parts []string
	for i := 0; i < len(values); {
		j := i
		for j+1 < len(values) && values[j+1] == values[j]+1 {
			j++
		}

		if j-i+1 >= min {
			parts = append(parts, strconv.Itoa(values[i])+"-"+strconv.Itoa(values[j]))
		} else {
			for k := i; k <= j; k++ {
				parts = append(parts, strconv.Itoa(values[k]))
			}
		}

		i = j + 1
	}

	return strings.Join(parts, ",")
}
