package ukdays

import (
	"slices"
	"time"
)

// MinYear is the first year the package has bank holidays for. Scotland's
// St Andrew's Day bank holiday began in 2007, and the rules used for years
// before GOV.UK's published list are only checked from then on. For earlier
// dates, a [Calendar] treats every weekday as a working day.
const MinYear = 2007

// Holiday titles, spelt as GOV.UK spells them.
const (
	titleNewYear      = "New Year’s Day"
	title2ndJanuary   = "2nd January"
	titleStPatrick    = "St Patrick’s Day"
	titleGoodFriday   = "Good Friday"
	titleEasterMonday = "Easter Monday"
	titleEarlyMay     = "Early May bank holiday"
	titleSpring       = "Spring bank holiday"
	titleBoyne        = "Battle of the Boyne (Orangemen’s Day)"
	titleSummer       = "Summer bank holiday"
	titleStAndrew     = "St Andrew’s Day"
	titleChristmas    = "Christmas Day"
	titleBoxingDay    = "Boxing Day"
)

// oneOff is a holiday that the standing rules don't produce: an extra day,
// or a regular holiday moved to another date. These cover the years between
// MinYear and the start of GOV.UK's published list, and apply to the whole
// UK. Later ones come from the published list.
type oneOff struct {
	date  Date
	title string
	moves string // title of the regular holiday this replaces, if any
}

var oneOffs = map[int][]oneOff{
	2011: {
		{date: Date{2011, time.April, 29}, title: "Bank holiday for the wedding of Prince William and Catherine Middleton"},
	},
	2012: {
		{date: Date{2012, time.June, 4}, title: titleSpring, moves: titleSpring},
		{date: Date{2012, time.June, 5}, title: "Queen’s Diamond Jubilee bank holiday"},
	},
}

// calculated returns the bank holidays for div in year, worked out from the
// rules, sorted by date. It returns nil for years before MinYear.
func calculated(div Division, year int) []Holiday {
	if year < MinYear || !div.valid() {
		return nil
	}

	var hs []Holiday
	add := func(d Date, title string, substitute bool) {
		h := Holiday{Date: d, Title: title, Source: SourceRules}
		if substitute {
			h.Notes = substituteNote
		}
		hs = append(hs, h)
	}

	// New Year's Day moves to the following Monday.
	newYear, moved := nextMondayIfWeekend(Date{year, time.January, 1})
	add(newYear, titleNewYear, moved)

	if div == Scotland {
		// 2nd January moves to the next weekday that isn't already the New
		// Year holiday.
		d := Date{year, time.January, 2}
		for d.isWeekend() || d == newYear {
			d = d.AddDays(1)
		}
		add(d, title2ndJanuary, d.Day != 2)
	}

	if div == NorthernIreland {
		d, moved := nextMondayIfWeekend(Date{year, time.March, 17})
		add(d, titleStPatrick, moved)
	}

	easter := easterSunday(year)
	add(easter.AddDays(-2), titleGoodFriday, false)
	if div != Scotland {
		add(easter.AddDays(1), titleEasterMonday, false)
	}

	add(nthMonday(year, time.May, 1), titleEarlyMay, false)
	add(lastMonday(year, time.May), titleSpring, false)

	if div == NorthernIreland {
		d, moved := nextMondayIfWeekend(Date{year, time.July, 12})
		add(d, titleBoyne, moved)
	}

	if div == Scotland {
		add(nthMonday(year, time.August, 1), titleSummer, false)
		d, moved := nextMondayIfWeekend(Date{year, time.November, 30})
		add(d, titleStAndrew, moved)
	} else {
		add(lastMonday(year, time.August), titleSummer, false)
	}

	// At Christmas the day itself keeps its place if it can. If Christmas
	// Day is at a weekend it moves to 27 December; if Boxing Day is, it
	// moves to 28 December.
	christmas := Date{year, time.December, 25}
	if christmas.isWeekend() {
		add(Date{year, time.December, 27}, titleChristmas, true)
	} else {
		add(christmas, titleChristmas, false)
	}
	boxing := Date{year, time.December, 26}
	if boxing.isWeekend() {
		add(Date{year, time.December, 28}, titleBoxingDay, true)
	} else {
		add(boxing, titleBoxingDay, false)
	}

	for _, o := range oneOffs[year] {
		if o.moves != "" {
			hs = slices.DeleteFunc(hs, func(h Holiday) bool { return h.Title == o.moves })
		}
		hs = append(hs, Holiday{Date: o.date, Title: o.title, Source: SourceRules})
	}

	slices.SortFunc(hs, func(a, b Holiday) int { return a.Date.Compare(b.Date) })
	return hs
}

// nextMondayIfWeekend returns d, or the following Monday if d is a Saturday
// or Sunday, and whether it moved.
func nextMondayIfWeekend(d Date) (Date, bool) {
	switch d.Weekday() {
	case time.Saturday:
		return d.AddDays(2), true
	case time.Sunday:
		return d.AddDays(1), true
	}
	return d, false
}

// nthMonday returns the nth Monday of the month, counting from 1.
func nthMonday(year int, month time.Month, n int) Date {
	first := Date{year, month, 1}
	offset := (int(time.Monday) - int(first.Weekday()) + 7) % 7
	return first.AddDays(offset + 7*(n-1))
}

// lastMonday returns the last Monday of the month.
func lastMonday(year int, month time.Month) Date {
	last := NewDate(year, month+1, 0)
	offset := (int(last.Weekday()) - int(time.Monday) + 7) % 7
	return last.AddDays(-offset)
}

// easterSunday returns the date of Easter Sunday in the Gregorian calendar,
// using the anonymous Gregorian algorithm (Meeus/Jones/Butcher).
func easterSunday(year int) Date {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return Date{year, time.Month(month), day}
}
