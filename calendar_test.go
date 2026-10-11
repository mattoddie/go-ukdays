package ukdays

import (
	"testing"
	"time"
)

func mustCalendar(t testing.TB, div Division) *Calendar {
	t.Helper()
	c, err := New(div)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func date(t testing.TB, s string) Date {
	t.Helper()
	d, err := ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestNewUnknownDivision(t *testing.T) {
	if _, err := New("wales"); err == nil {
		t.Error("New accepted an unknown division")
	}
}

func TestIsWorkingDay(t *testing.T) {
	tests := []struct {
		div  Division
		date string
		want bool
	}{
		{EnglandAndWales, "2026-12-24", true},
		{EnglandAndWales, "2026-12-25", false}, // Christmas Day
		{EnglandAndWales, "2026-12-26", false}, // Saturday
		{EnglandAndWales, "2026-12-28", false}, // Boxing Day substitute
		{EnglandAndWales, "2027-01-04", true},
		{Scotland, "2027-01-04", false},        // 2nd January substitute
		{EnglandAndWales, "2026-04-06", false}, // Easter Monday
		{Scotland, "2026-04-06", true},         // not a holiday in Scotland
		{NorthernIreland, "2026-03-17", false}, // St Patrick's Day
		{EnglandAndWales, "2026-03-17", true},
		{NorthernIreland, "2026-07-13", false}, // Battle of the Boyne substitute
		{Scotland, "2026-06-15", false},        // World Cup bank holiday
		{Scotland, "2026-08-03", false},        // Scottish summer bank holiday
		{EnglandAndWales, "2026-08-03", true},
		{EnglandAndWales, "2035-12-25", false}, // from the rules
		{EnglandAndWales, "2000-12-25", true},  // before MinYear
	}
	for _, tt := range tests {
		c := mustCalendar(t, tt.div)
		if got := c.IsWorkingDay(date(t, tt.date)); got != tt.want {
			t.Errorf("%s IsWorkingDay(%s) = %v, want %v", tt.div, tt.date, got, tt.want)
		}
	}
}

func TestNextAndPreviousWorkingDay(t *testing.T) {
	c := mustCalendar(t, EnglandAndWales)
	if got := c.NextWorkingDay(date(t, "2026-12-24")); got.String() != "2026-12-29" {
		t.Errorf("NextWorkingDay = %v, want 2026-12-29", got)
	}
	if got := c.PreviousWorkingDay(date(t, "2026-12-29")); got.String() != "2026-12-24" {
		t.Errorf("PreviousWorkingDay = %v, want 2026-12-24", got)
	}
	if got := c.NextWorkingDay(date(t, "2026-04-02")); got.String() != "2026-04-07" {
		t.Errorf("NextWorkingDay over Easter = %v, want 2026-04-07", got)
	}
}

func TestAddWorkingDays(t *testing.T) {
	c := mustCalendar(t, EnglandAndWales)
	tests := []struct {
		from string
		n    int
		want string
	}{
		{"2026-12-24", 0, "2026-12-24"},
		{"2026-12-25", 0, "2026-12-25"},
		{"2026-12-24", 3, "2026-12-31"},
		{"2026-12-24", 4, "2027-01-04"},
		{"2027-01-04", -4, "2026-12-24"},
		{"2026-10-09", 1, "2026-10-12"},
		{"2026-10-10", 1, "2026-10-12"}, // from a Saturday
		{"2026-10-10", -1, "2026-10-09"},
	}
	for _, tt := range tests {
		if got := c.AddWorkingDays(date(t, tt.from), tt.n); got.String() != tt.want {
			t.Errorf("AddWorkingDays(%s, %d) = %v, want %s", tt.from, tt.n, got, tt.want)
		}
	}
}

// TestAddWorkingDaysLarge checks the week-skipping path against stepping one
// day at a time.
func TestAddWorkingDaysLarge(t *testing.T) {
	for _, div := range Divisions() {
		c := mustCalendar(t, div)
		start := date(t, "2024-02-29")
		for _, n := range []int{299, 301, 1000, 2500, -301, -2500} {
			got := c.AddWorkingDays(start, n)
			want := stepWorkingDays(c, start, n)
			if got != want {
				t.Errorf("%s AddWorkingDays(%v, %d) = %v, want %v", div, start, n, got, want)
			}
		}
	}
}

func stepWorkingDays(c *Calendar, d Date, n int) Date {
	step := 1
	if n < 0 {
		step, n = -1, -n
	}
	for n > 0 {
		d = d.AddDays(step)
		if c.IsWorkingDay(d) {
			n--
		}
	}
	return d
}

func TestWorkingDaysBetween(t *testing.T) {
	c := mustCalendar(t, EnglandAndWales)
	tests := []struct {
		start, end string
		want       int
	}{
		{"2026-12-24", "2026-12-24", 0},
		{"2026-12-24", "2026-12-25", 1},
		{"2026-12-24", "2027-01-01", 4},
		{"2027-01-01", "2026-12-24", -4},
		{"2026-01-01", "2027-01-01", 253}, // 261 weekdays less 8 bank holidays
	}
	for _, tt := range tests {
		if got := c.WorkingDaysBetween(date(t, tt.start), date(t, tt.end)); got != tt.want {
			t.Errorf("WorkingDaysBetween(%s, %s) = %d, want %d", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestHolidays(t *testing.T) {
	c := mustCalendar(t, Scotland)
	hs := c.Holidays(2026)
	if len(hs) != 10 {
		t.Fatalf("Scotland 2026 has %d holidays %v, want 10", len(hs), dates(hs))
	}
	hs[0].Title = "changed"
	if c.Holidays(2026)[0].Title == "changed" {
		t.Error("Holidays returned the calendar's own slice")
	}
	if got := c.Holidays(MinYear - 1); got != nil {
		t.Errorf("Holidays(%d) = %v, want nil", MinYear-1, got)
	}

	between := c.HolidaysBetween(date(t, "2026-12-01"), date(t, "2027-01-05"))
	if got := dates(between); len(got) != 4 || got[0] != "2026-12-25" || got[3] != "2027-01-04" {
		t.Errorf("HolidaysBetween = %v", got)
	}
}

func TestHolidayOn(t *testing.T) {
	c := mustCalendar(t, NorthernIreland)
	h, ok := c.HolidayOn(date(t, "2026-07-13"))
	if !ok || h.Title != titleBoyne || !h.IsSubstitute() || h.Source != SourceGOVUK {
		t.Errorf("HolidayOn(2026-07-13) = %+v, %v", h, ok)
	}
	if _, ok := c.HolidayOn(date(t, "2026-07-14")); ok {
		t.Error("HolidayOn found a holiday on an ordinary day")
	}
}

func TestSource(t *testing.T) {
	c := mustCalendar(t, EnglandAndWales)
	if s, ok := c.Source(2026); !ok || s != SourceGOVUK {
		t.Errorf("Source(2026) = %v, %v", s, ok)
	}
	if s, ok := c.Source(2012); !ok || s != SourceRules {
		t.Errorf("Source(2012) = %v, %v", s, ok)
	}
	if _, ok := c.Source(1999); ok {
		t.Error("Source(1999) reported holidays")
	}
	if c.Division() != EnglandAndWales {
		t.Error("Division is wrong")
	}
}

func TestWeekdaysBetween(t *testing.T) {
	start := Date{2026, time.January, 1}
	for days := 0; days < 30; days++ {
		end := start.AddDays(days)
		want := 0
		for d := start; d.Before(end); d = d.AddDays(1) {
			if !d.isWeekend() {
				want++
			}
		}
		if got := weekdaysBetween(start, end); got != want {
			t.Errorf("weekdaysBetween over %d days = %d, want %d", days, got, want)
		}
	}
}

func BenchmarkAddWorkingDays(b *testing.B) {
	c := mustCalendar(b, EnglandAndWales)
	d := NewDate(2026, time.October, 12)
	for i := 0; i < b.N; i++ {
		c.AddWorkingDays(d, 20)
	}
}
