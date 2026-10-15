package ast

import (
	"fmt"
	"strings"
	"time"
)

type state uint8

const (
	stateWord      state = iota // first word, seed or frequency
	statePipe                   // after the first word, expecting | or @ or end
	stateFrequency              // frequency after the pipe
	stateAfterFreq              // after the frequency, expecting @ or end
	stateKeyword                // expecting on, during, or in
	stateDay                    // argument of on
	stateDayNumber              // argument of on day
	stateDuring                 // argument of during
	stateIn                     // argument of in
)

var weekdays = map[string]time.Weekday{
	"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday,
	"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
}

type parser struct {
	t     T
	input string

	// first word, parsed speculatively as a frequency while it is scanned.
	wordStart, wordEnd int
	word               freq

	freq freq

	state state
	seen  [3]bool // on, during, in
	// clauses counts every clause read after @, including a pin.
	clauses int
	pinAt   int
}

// freq incrementally parses a compact duration such as 5d1h5m3s.
type freq struct {
	total  time.Duration
	number int64
	digits int
	units  int
	err    string
	pos    int
}

func (f *freq) feed(c byte, pos int) {
	if f.err != "" {
		return
	}

	switch {
	case c >= '0' && c <= '9':
		f.number = f.number*10 + int64(c-'0')
		f.digits++
		if f.number > 1<<40 {
			f.err, f.pos = "duration overflow", pos
		}
	default:
		var unit time.Duration
		switch c {
		case 'd':
			unit = 24 * time.Hour
		case 'h':
			unit = time.Hour
		case 'm':
			unit = time.Minute
		case 's':
			unit = time.Second
		default:
			f.err, f.pos = fmt.Sprintf("unexpected character %q in frequency", c), pos
			return
		}

		if f.digits == 0 {
			f.err, f.pos = fmt.Sprintf("unit %q without a number", c), pos
			return
		}

		f.total += time.Duration(f.number) * unit
		f.number, f.digits = 0, 0
		f.units++
	}
}

func (f *freq) done(pos int) (time.Duration, error) {
	if f.err != "" {
		return 0, perr(f.pos, f.err)
	}

	if f.digits != 0 {
		return 0, perr(pos, "number without a unit")
	}

	if f.units == 0 {
		return 0, perr(pos, "missing frequency")
	}

	if f.total <= 0 {
		return 0, perr(pos, "frequency must be positive")
	}

	return f.total, nil
}

func perr(offset int, msg string) error {
	return fmt.Errorf("ast: offset %d: %s", offset, msg)
}

// Parse reads input exactly once, left to right, building the tree as it goes.
func Parse(input string) (T, error) {
	p := &parser{input: input}
	p.t.Pin = -1

	var (
		i   int
		n   = len(input)
		tok = -1
	)

	for ; i <= n; i++ {
		var c byte = ' '
		if i < n {
			c = input[i]
		}

		boundary := c == ' ' || c == '\t' || c == '\n' || c == '\r'
		special := c == '|' || c == '@'

		// accumulate characters of the current token.
		if !boundary && !special && i < n {
			if tok < 0 {
				tok = i
			}

			switch p.state {
			case stateWord:
				p.word.feed(c, i)
			case stateFrequency:
				p.freq.feed(c, i)
			}
			continue
		}

		// token boundary.
		if tok >= 0 {
			if err := p.token(tok, i); err != nil {
				return T{}, err
			}
			tok = -1
		}

		if special {
			if err := p.special(c, i); err != nil {
				return T{}, err
			}
		}
	}

	return p.finish(n)
}

// token handles a complete whitespace delimited word.
func (p *parser) token(start, end int) error {
	text := p.input[start:end]

	switch p.state {
	case stateWord:
		p.wordStart, p.wordEnd = start, end
		p.state = statePipe
	case stateFrequency:
		p.state = stateAfterFreq
	case stateKeyword:
		switch text {
		case "on":
			if p.seen[0] {
				return perr(start, "duplicate on clause")
			}
			p.seen[0], p.state = true, stateDay
			p.clauses++
		case "during":
			if p.seen[1] {
				return perr(start, "duplicate during clause")
			}
			p.seen[1], p.state = true, stateDuring
			p.clauses++
		case "in":
			if p.seen[2] {
				return perr(start, "duplicate in clause")
			}
			p.seen[2], p.state = true, stateIn
			p.clauses++
		default:
			if text[0] >= '0' && text[0] <= '9' {
				return p.pin(text, start)
			}

			return perr(start, fmt.Sprintf("unexpected %q, expected a time, on, during, or in", text))
		}
	case stateDay:
		return p.day(text, start)
	case stateDayNumber:
		return p.dayNumbers(text, start)
	case stateDuring:
		return p.during(text, start)
	case stateIn:
		p.t.Tz = text
		p.state = stateKeyword
	case statePipe, stateAfterFreq:
		return perr(start, fmt.Sprintf("unexpected %q", text))
	}

	return nil
}

// special handles | and @.
func (p *parser) special(c byte, at int) error {
	switch {
	case c == '|' && p.state == statePipe:
		p.t.Seed, p.t.HasSeed = p.input[p.wordStart:p.wordEnd], true
		p.state = stateFrequency
	case c == '@' && (p.state == statePipe || p.state == stateAfterFreq || p.state == stateFrequency):
		if p.state == stateFrequency && p.freq.units == 0 && p.freq.digits == 0 {
			return perr(at, "missing frequency")
		}

		if err := p.commitFrequency(at); err != nil {
			return err
		}
		p.state = stateKeyword
	case c == '|':
		return perr(at, "unexpected |")
	default:
		return perr(at, "unexpected @")
	}

	return nil
}

// commitFrequency settles the frequency when the conditions delimiter or end is reached.
func (p *parser) commitFrequency(at int) error {
	var (
		hz  time.Duration
		err error
	)

	switch p.state {
	case statePipe:
		// no pipe: the first word was the frequency.
		hz, err = p.word.done(p.wordEnd)
	default:
		hz, err = p.freq.done(at)
	}

	if err != nil {
		return err
	}

	p.t.Hz = hz
	return nil
}

func (p *parser) finish(n int) (T, error) {
	switch p.state {
	case stateWord:
		return T{}, perr(n, "missing frequency")
	case statePipe, stateFrequency, stateAfterFreq:
		if err := p.commitFrequency(n); err != nil {
			return T{}, err
		}
	case stateKeyword:
		if p.clauses == 0 {
			return T{}, perr(n, "missing condition after @")
		}
	default:
		return T{}, perr(n, "incomplete condition")
	}

	if p.t.Pin >= 0 && p.t.HasWindow && !ticksInWindow(p.t) {
		return T{}, perr(p.pinAt, fmt.Sprintf("time %s never falls inside during %s-%s", clockString(p.t.Pin), clockString(p.t.Start), clockString(p.t.End)))
	}

	return p.t, nil
}

func (p *parser) day(text string, at int) error {
	switch text {
	case "weekdays":
		p.t.day = dayWeekdays
	case "weekends":
		p.t.day = dayWeekends
	case "everyday":
		p.t.day = dayEveryday
	case "day":
		p.state = stateDayNumber
		return nil
	default:
		return p.dayNames(text, at)
	}

	p.state = stateKeyword
	return nil
}

func (p *parser) dayNames(text string, at int) error {
	if from, to, ok := strings.Cut(text, "-"); ok {
		f, fok := weekdays[from]
		t, tok := weekdays[to]
		if !fok || !tok {
			return perr(at, fmt.Sprintf("invalid day range %q", text))
		}
		p.t.day, p.t.From, p.t.To = dayRange, f, t
		p.state = stateKeyword
		return nil
	}

	pos := at
	for _, name := range strings.Split(text, ",") {
		d, ok := weekdays[name]
		if !ok {
			return perr(pos, fmt.Sprintf("invalid day %q", name))
		}
		p.t.Days = append(p.t.Days, d)
		pos += len(name) + 1
	}

	p.t.day = dayList
	p.state = stateKeyword
	return nil
}

func (p *parser) dayNumbers(text string, at int) error {
	pos := at
	for _, part := range strings.Split(text, ",") {
		if part == "" || len(part) > 2 {
			return perr(pos, fmt.Sprintf("invalid day of month %q", part))
		}

		n := 0
		for i := 0; i < len(part); i++ {
			if part[i] < '0' || part[i] > '9' {
				return perr(pos+i, fmt.Sprintf("invalid day of month %q", part))
			}
			n = n*10 + int(part[i]-'0')
		}

		if n < 1 || n > 31 {
			return perr(pos, fmt.Sprintf("day of month out of range: %d", n))
		}

		p.t.DaysOfMonth = append(p.t.DaysOfMonth, n)
		pos += len(part) + 1
	}

	p.t.day = dayOfMonth
	p.state = stateKeyword
	return nil
}

func (p *parser) during(text string, at int) error {
	from, to, ok := strings.Cut(text, "-")
	if !ok {
		return perr(at, fmt.Sprintf("invalid time range %q, expected HH:MM[:SS[.mmm]]-HH:MM[:SS[.mmm]]", text))
	}

	start, err := clock(from, at)
	if err != nil {
		return err
	}

	end, err := clock(to, at+len(from)+1)
	if err != nil {
		return err
	}

	if start >= 24*time.Hour {
		return perr(at, "start time must be before 24:00")
	}

	if start == end {
		return perr(at, "empty time range")
	}

	p.t.HasWindow, p.t.Start, p.t.End = true, start, end
	p.state = stateKeyword
	return nil
}

// pin reads the time that pins the frequency's slot. it must be the first
// clause after @ and shorter than the frequency.
func (p *parser) pin(text string, at int) error {
	if p.clauses > 0 {
		return perr(at, "a time must be the first clause after @")
	}

	d, err := clock(text, at)
	if err != nil {
		return err
	}

	if d >= 24*time.Hour {
		return perr(at, "time must be before 24:00")
	}

	if d >= p.t.Hz {
		return perr(at, fmt.Sprintf("time %s must be shorter than the frequency", text))
	}

	p.t.Pin, p.pinAt = d, at
	p.clauses++
	return nil
}

// ticksInWindow reports whether any tick of a pinned schedule can land inside
// its during window. the ticks of a local day are pin, pin+hz, ... below 24h,
// or just the pin for a frequency of a day or more.
func ticksInWindow(t T) bool {
	const day = 24 * time.Hour

	// any tick in [lo, hi) of the day.
	has := func(lo, hi time.Duration) bool {
		if lo >= hi {
			return false
		}

		var k time.Duration
		if lo > t.Pin {
			k = (lo - t.Pin + t.Hz - 1) / t.Hz
		}

		if t.Hz >= day && k > 0 {
			return false
		}

		off := t.Pin + k*t.Hz
		return off < hi && off < day
	}

	if t.Start < t.End {
		return has(t.Start, t.End)
	}

	return has(t.Start, day) || has(0, t.End)
}

// clockString renders a duration as HH:MM[:SS[.mmm]].
func clockString(d time.Duration) string {
	h, m, s, ms := int(d/time.Hour), int(d%time.Hour/time.Minute), int(d%time.Minute/time.Second), int(d%time.Second/time.Millisecond)
	switch {
	case ms != 0:
		return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
	case s != 0:
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	default:
		return fmt.Sprintf("%02d:%02d", h, m)
	}
}

// clock parses HH:MM[:SS[.mmm]], at most millisecond precision.
func clock(s string, at int) (time.Duration, error) {
	bad := func() error {
		return perr(at, fmt.Sprintf("invalid time %q, expected HH:MM[:SS[.mmm]]", s))
	}

	two := func(i int) (int, bool) {
		if i+2 > len(s) || s[i] < '0' || s[i] > '9' || s[i+1] < '0' || s[i+1] > '9' {
			return 0, false
		}
		return int(s[i]-'0')*10 + int(s[i+1]-'0'), true
	}

	if len(s) < 5 || s[2] != ':' {
		return 0, bad()
	}

	h, hok := two(0)
	m, mok := two(3)
	if !hok || !mok {
		return 0, bad()
	}

	var sec, ms int
	if rest := s[5:]; rest != "" {
		if rest[0] != ':' {
			return 0, bad()
		}

		var ok bool
		if sec, ok = two(6); !ok {
			return 0, bad()
		}

		if rest = rest[3:]; rest != "" {
			if rest[0] != '.' || len(rest) < 2 || len(rest) > 4 {
				return 0, bad()
			}

			scale := 100
			for _, c := range []byte(rest[1:]) {
				if c < '0' || c > '9' {
					return 0, bad()
				}
				ms += int(c-'0') * scale
				scale /= 10
			}
		}
	}

	if m > 59 || sec > 59 || h > 24 || (h == 24 && (m != 0 || sec != 0 || ms != 0)) {
		return 0, perr(at, fmt.Sprintf("time out of range %q", s))
	}

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second + time.Duration(ms)*time.Millisecond, nil
}
