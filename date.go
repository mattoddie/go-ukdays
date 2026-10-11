package ukdays

import (
	"fmt"
	"time"
)

// Date is a calendar date with no time of day or time zone, such as the date
// of a bank holiday.
//
// The zero value is not a valid date. Use [NewDate], [DateOf] or [ParseDate]
// to make one.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// NewDate returns the date for the given year, month and day. Out-of-range
// values are normalised in the same way as [time.Date], so NewDate(2026,
// time.December, 32) is 1 January 2027.
func NewDate(year int, month time.Month, day int) Date {
	return DateOf(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

// DateOf returns the calendar date of t in t's own location.
func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return Date{Year: y, Month: m, Day: d}
}

// Today returns the current date in loc. UK callers usually want
// Europe/London, where the date changes at local midnight.
func Today(loc *time.Location) Date {
	return DateOf(time.Now().In(loc))
}

// ParseDate parses a date in ISO 8601 form, such as "2026-12-25".
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("ukdays: invalid date %q: want YYYY-MM-DD", s)
	}
	return DateOf(t), nil
}

// String returns the date in ISO 8601 form, such as "2026-12-25".
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}

// IsValid reports whether d is a real calendar date, such as 29 February in
// a leap year but not in other years.
func (d Date) IsValid() bool {
	return d == NewDate(d.Year, d.Month, d.Day)
}

// In returns the time at the start of d in loc.
func (d Date) In(loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

// AddDays returns the date n calendar days after d. A negative n goes back.
func (d Date) AddDays(n int) Date {
	return NewDate(d.Year, d.Month, d.Day+n)
}

// Weekday returns the day of the week d falls on.
func (d Date) Weekday() time.Weekday {
	return d.utc().Weekday()
}

// DaysSince returns the number of calendar days from e to d. It is negative
// when d is before e.
func (d Date) DaysSince(e Date) int {
	return int((d.utc().Unix() - e.utc().Unix()) / (24 * 60 * 60))
}

// Compare returns -1 if d is before e, +1 if d is after e, and 0 if they are
// the same date.
func (d Date) Compare(e Date) int {
	switch {
	case d.Year != e.Year:
		return sign(d.Year - e.Year)
	case d.Month != e.Month:
		return sign(int(d.Month) - int(e.Month))
	default:
		return sign(d.Day - e.Day)
	}
}

// Before reports whether d is before e.
func (d Date) Before(e Date) bool { return d.Compare(e) < 0 }

// After reports whether d is after e.
func (d Date) After(e Date) bool { return d.Compare(e) > 0 }

// MarshalText implements [encoding.TextMarshaler]. The date is written in the
// same form as [Date.String], so it encodes as a JSON string.
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]. It accepts the same
// form as [ParseDate].
func (d *Date) UnmarshalText(text []byte) error {
	parsed, err := ParseDate(string(text))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Date) utc() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func (d Date) isWeekend() bool {
	wd := d.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
