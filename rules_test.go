package ukdays

import (
	"testing"
	"time"
)

// publishedOnly lists dates in GOV.UK's data that the rules can't produce:
// one-off holidays, and regular holidays moved to make room for them.
var publishedOnly = map[Division][]string{
	EnglandAndWales: {"2020-05-08", "2022-06-02", "2022-06-03", "2022-09-19", "2023-05-08"},
	Scotland:        {"2020-05-08", "2022-06-02", "2022-06-03", "2022-09-19", "2023-05-08", "2026-06-15"},
	NorthernIreland: {"2020-05-08", "2022-06-02", "2022-06-03", "2022-09-19", "2023-05-08"},
}

// rulesOnly lists the regular dates those moves replaced.
var rulesOnly = []string{"2020-05-04", "2022-05-30"}

// TestRulesMatchPublished checks the rules against every year GOV.UK
// publishes. Apart from the one-offs above, the dates, titles and
// substitute-day notes must match exactly.
func TestRulesMatchPublished(t *testing.T) {
	data := Embedded()
	for _, div := range Divisions() {
		skipPublished := set(publishedOnly[div])
		skipRules := set(rulesOnly)

		published := map[Date]Holiday{}
		for _, h := range data.divisions[div] {
			if !skipPublished[h.Date.String()] {
				published[h.Date] = h
			}
		}
		first, last := data.Years()
		if first > 2019 || last < 2028 {
			t.Fatalf("embedded data covers %d to %d; expected at least 2019 to 2028", first, last)
		}

		for year := first; year <= last; year++ {
			for _, h := range calculated(div, year) {
				if skipRules[h.Date.String()] {
					continue
				}
				p, ok := published[h.Date]
				if !ok {
					t.Errorf("%s %s: rules give %q, GOV.UK has no holiday", div, h.Date, h.Title)
					continue
				}
				if p.Title != h.Title || p.Notes != h.Notes {
					t.Errorf("%s %s: rules give %q (%q), GOV.UK has %q (%q)", div, h.Date, h.Title, h.Notes, p.Title, p.Notes)
				}
				delete(published, h.Date)
			}
		}
		for d, h := range published {
			t.Errorf("%s %s: GOV.UK has %q, rules give nothing", div, d, h.Title)
		}
	}
}

func TestCalculatedBeforeGOVUKData(t *testing.T) {
	tests := []struct {
		div   Division
		year  int
		dates []string
	}{
		{EnglandAndWales, 2018, []string{
			"2018-01-01", "2018-03-30", "2018-04-02", "2018-05-07", "2018-05-28",
			"2018-08-27", "2018-12-25", "2018-12-26",
		}},
		// The royal wedding.
		{EnglandAndWales, 2011, []string{
			"2011-01-03", "2011-04-22", "2011-04-25", "2011-04-29", "2011-05-02",
			"2011-05-30", "2011-08-29", "2011-12-26", "2011-12-27",
		}},
		// The Diamond Jubilee moved the spring bank holiday to 4 June.
		{Scotland, 2012, []string{
			"2012-01-02", "2012-01-03", "2012-04-06", "2012-05-07", "2012-06-04",
			"2012-06-05", "2012-08-06", "2012-11-30", "2012-12-25", "2012-12-26",
		}},
		{NorthernIreland, 2017, []string{
			"2017-01-02", "2017-03-17", "2017-04-14", "2017-04-17", "2017-05-01",
			"2017-05-29", "2017-07-12", "2017-08-28", "2017-12-25", "2017-12-26",
		}},
	}
	for _, tt := range tests {
		got := calculated(tt.div, tt.year)
		if len(got) != len(tt.dates) {
			t.Errorf("%s %d: got %d holidays %v, want %d", tt.div, tt.year, len(got), dates(got), len(tt.dates))
			continue
		}
		for i, h := range got {
			if h.Date.String() != tt.dates[i] {
				t.Errorf("%s %d: holiday %d is %s (%s), want %s", tt.div, tt.year, i, h.Date, h.Title, tt.dates[i])
			}
			if h.Source != SourceRules {
				t.Errorf("%s %s: source %v, want rules", tt.div, h.Date, h.Source)
			}
		}
	}
}

func TestCalculatedBeforeMinYear(t *testing.T) {
	if got := calculated(EnglandAndWales, MinYear-1); got != nil {
		t.Errorf("calculated(%d) = %v, want nil", MinYear-1, got)
	}
}

func TestCalculatedFutureYearsAreWeekdays(t *testing.T) {
	for _, div := range Divisions() {
		for year := 2029; year <= 2400; year++ {
			hs := calculated(div, year)
			seen := map[Date]bool{}
			for _, h := range hs {
				if h.Date.isWeekend() {
					t.Fatalf("%s %s: %q falls on a %s", div, h.Date, h.Title, h.Date.Weekday())
				}
				if seen[h.Date] {
					t.Fatalf("%s %s: two holidays on the same day", div, h.Date)
				}
				seen[h.Date] = true
			}
		}
	}
}

func TestEasterSunday(t *testing.T) {
	tests := map[int]string{
		2008: "2008-03-23", // earliest in the period
		2011: "2011-04-24",
		2019: "2019-04-21",
		2024: "2024-03-31",
		2026: "2026-04-05",
		2038: "2038-04-25", // latest possible date
		2100: "2100-03-28",
	}
	for year, want := range tests {
		if got := easterSunday(year).String(); got != want {
			t.Errorf("easterSunday(%d) = %s, want %s", year, got, want)
		}
	}
}

func TestMondayHelpers(t *testing.T) {
	if got := nthMonday(2026, time.May, 1).String(); got != "2026-05-04" {
		t.Errorf("first Monday of May 2026 = %s", got)
	}
	if got := lastMonday(2026, time.May).String(); got != "2026-05-25" {
		t.Errorf("last Monday of May 2026 = %s", got)
	}
	if got := lastMonday(2027, time.August).String(); got != "2027-08-30" {
		t.Errorf("last Monday of August 2027 = %s", got)
	}
}

func set(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func dates(hs []Holiday) []string {
	out := make([]string, len(hs))
	for i, h := range hs {
		out[i] = h.Date.String()
	}
	return out
}
