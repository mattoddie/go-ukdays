package ukdays

import "testing"

func FuzzParseDate(f *testing.F) {
	for _, s := range []string{"2026-12-25", "2028-02-29", "2027-02-29", "", "0000-01-01", "9999-12-31"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		d, err := ParseDate(s)
		if err != nil {
			return
		}
		if !d.IsValid() {
			t.Fatalf("ParseDate(%q) = %v, which is not a valid date", s, d)
		}
		back, err := ParseDate(d.String())
		if err != nil || back != d {
			t.Fatalf("round trip of %q: %v, %v", s, back, err)
		}
	})
}

// FuzzWorkingDays checks that adding working days and counting them agree:
// n working days after a working day is exactly n working days away.
func FuzzWorkingDays(f *testing.F) {
	f.Add(uint8(0), int64(0), int16(1))
	f.Add(uint8(1), int64(365), int16(-20))
	f.Add(uint8(2), int64(-4000), int16(700))
	cals := map[uint8]*Calendar{}
	for i, div := range Divisions() {
		cals[uint8(i)] = mustCalendar(f, div)
	}
	f.Fuzz(func(t *testing.T, which uint8, offset int64, n int16) {
		c := cals[which%3]
		// Keep within about 200 years of 2026 so each run stays quick.
		start := NewDate(2026, 1, 1).AddDays(int(offset % 73000))
		if !c.IsWorkingDay(start) {
			start = c.NextWorkingDay(start)
		}
		end := c.AddWorkingDays(start, int(n))
		if !c.IsWorkingDay(end) {
			t.Fatalf("AddWorkingDays(%v, %d) = %v, not a working day", start, n, end)
		}
		var got int
		if n >= 0 {
			got = c.WorkingDaysBetween(start, end)
		} else {
			got = -c.WorkingDaysBetween(end, start)
		}
		if got != int(n) {
			t.Fatalf("%s: AddWorkingDays(%v, %d) = %v, but WorkingDaysBetween gives %d", c.Division(), start, n, end, got)
		}
	})
}
