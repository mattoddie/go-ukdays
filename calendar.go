package ukdays

import "slices"

// Calendar answers working-day questions for one division. A working day is
// a Monday to Friday that is not a bank holiday.
//
// For years covered by GOV.UK's data the calendar uses the published list;
// for other years from [MinYear] on it uses the standing rules (see
// [SourceRules]). Before MinYear it knows no bank holidays, so every weekday
// counts as a working day.
//
// A Calendar is safe for concurrent use.
type Calendar struct {
	division  Division
	published map[int][]Holiday
}

// New returns a calendar for div using the data built into the package. It
// returns an error only if div is not one of the three divisions.
func New(div Division) (*Calendar, error) {
	return Embedded().Calendar(div)
}

func newCalendar(div Division, hs []Holiday) *Calendar {
	c := &Calendar{division: div, published: make(map[int][]Holiday)}
	for _, h := range hs {
		c.published[h.Date.Year] = append(c.published[h.Date.Year], h)
	}
	return c
}

// Division returns the division the calendar is for.
func (c *Calendar) Division() Division {
	return c.division
}

// Source reports where the holidays for year come from. It returns false
// for years before [MinYear], which have none.
func (c *Calendar) Source(year int) (Source, bool) {
	if _, ok := c.published[year]; ok {
		return SourceGOVUK, true
	}
	if year < MinYear {
		return 0, false
	}
	return SourceRules, true
}

// Holidays returns the bank holidays in year, sorted by date. It returns nil
// for years before [MinYear].
func (c *Calendar) Holidays(year int) []Holiday {
	if hs, ok := c.published[year]; ok {
		return slices.Clone(hs)
	}
	return calculated(c.division, year)
}

// HolidaysBetween returns the bank holidays from start up to, but not
// including, end, sorted by date.
func (c *Calendar) HolidaysBetween(start, end Date) []Holiday {
	var out []Holiday
	for y := start.Year; y <= end.Year; y++ {
		for _, h := range c.holidays(y) {
			if !h.Date.Before(start) && h.Date.Before(end) {
				out = append(out, h)
			}
		}
	}
	return out
}

// HolidayOn returns the bank holiday observed on d, if there is one.
func (c *Calendar) HolidayOn(d Date) (Holiday, bool) {
	for _, h := range c.holidays(d.Year) {
		if h.Date == d {
			return h, true
		}
	}
	return Holiday{}, false
}

// IsWorkingDay reports whether d is a weekday that is not a bank holiday.
func (c *Calendar) IsWorkingDay(d Date) bool {
	if d.isWeekend() {
		return false
	}
	_, holiday := c.HolidayOn(d)
	return !holiday
}

// NextWorkingDay returns the first working day after d.
func (c *Calendar) NextWorkingDay(d Date) Date {
	return c.AddWorkingDays(d, 1)
}

// PreviousWorkingDay returns the last working day before d.
func (c *Calendar) PreviousWorkingDay(d Date) Date {
	return c.AddWorkingDays(d, -1)
}

// AddWorkingDays returns the date n working days after d, so adding 1 to a
// Friday gives the following Monday when that Monday is not a bank holiday.
// A negative n counts back. d itself need not be a working day; when n is 0,
// d is returned unchanged.
func (c *Calendar) AddWorkingDays(d Date, n int) Date {
	step, remaining := 1, n
	if n < 0 {
		step, remaining = -1, -n
	}
	cur := d
	for remaining > 0 {
		if remaining > 300 {
			// Jump whole weeks that can't contain more than remaining
			// working days, then count how many they did contain.
			days := (remaining - 10) / 5 * 7
			next := cur.AddDays(step * days)
			var counted int
			if step > 0 {
				counted = c.WorkingDaysBetween(cur.AddDays(1), next.AddDays(1))
			} else {
				counted = c.WorkingDaysBetween(next, cur)
			}
			cur, remaining = next, remaining-counted
			continue
		}
		cur = cur.AddDays(step)
		if c.IsWorkingDay(cur) {
			remaining--
		}
	}
	return cur
}

// WorkingDaysBetween returns the number of working days from start up to,
// but not including, end. It is negative when end is before start.
func (c *Calendar) WorkingDaysBetween(start, end Date) int {
	if end.Before(start) {
		return -c.WorkingDaysBetween(end, start)
	}
	n := weekdaysBetween(start, end)
	for _, h := range c.HolidaysBetween(start, end) {
		if !h.Date.isWeekend() {
			n--
		}
	}
	return n
}

// holidays returns the holidays in year without copying published data.
// Callers must not modify the result.
func (c *Calendar) holidays(year int) []Holiday {
	if hs, ok := c.published[year]; ok {
		return hs
	}
	return calculated(c.division, year)
}

// weekdaysBetween counts Mondays to Fridays from start up to, but not
// including, end. start must not be after end.
func weekdaysBetween(start, end Date) int {
	days := end.DaysSince(start)
	n := days / 7 * 5
	d := start.AddDays(days / 7 * 7)
	for d.Before(end) {
		if !d.isWeekend() {
			n++
		}
		d = d.AddDays(1)
	}
	return n
}
