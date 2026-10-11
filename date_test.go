package ukdays

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	d, err := ParseDate("2028-02-29")
	if err != nil || d != (Date{2028, time.February, 29}) {
		t.Fatalf("ParseDate(2028-02-29) = %v, %v", d, err)
	}
	for _, bad := range []string{"", "2027-02-29", "2026-13-01", "26-12-25", "2026/12/25", "2026-12-25T00:00:00Z", " 2026-12-25"} {
		if _, err := ParseDate(bad); err == nil {
			t.Errorf("ParseDate(%q) succeeded, want error", bad)
		}
	}
}

func TestDateArithmetic(t *testing.T) {
	d := NewDate(2026, time.December, 31)
	if got := d.AddDays(1); got != (Date{2027, time.January, 1}) {
		t.Errorf("AddDays(1) = %v", got)
	}
	if got := d.AddDays(-365); got != (Date{2025, time.December, 31}) {
		t.Errorf("AddDays(-365) = %v", got)
	}
	if got := NewDate(2026, time.December, 32); got != (Date{2027, time.January, 1}) {
		t.Errorf("NewDate normalises to %v", got)
	}
	if got := (Date{2027, time.January, 1}).DaysSince(Date{2026, time.January, 1}); got != 365 {
		t.Errorf("DaysSince = %d, want 365", got)
	}
	if got := (Date{2026, time.January, 1}).DaysSince(Date{2027, time.January, 1}); got != -365 {
		t.Errorf("DaysSince = %d, want -365", got)
	}
	if wd := d.Weekday(); wd != time.Thursday {
		t.Errorf("Weekday = %v, want Thursday", wd)
	}
}

func TestDateCompare(t *testing.T) {
	a, b := Date{2026, time.May, 4}, Date{2026, time.May, 25}
	if !a.Before(b) || a.After(b) || a.Compare(Date{2026, time.May, 4}) != 0 || b.Compare(a) != 1 {
		t.Error("Compare gives the wrong order")
	}
	if !(Date{2025, time.December, 31}).Before(Date{2026, time.January, 1}) {
		t.Error("year boundary ordered wrongly")
	}
}

func TestDateIsValid(t *testing.T) {
	if !(Date{2028, time.February, 29}).IsValid() {
		t.Error("29 February 2028 should be valid")
	}
	if (Date{2027, time.February, 29}).IsValid() || (Date{}).IsValid() {
		t.Error("invalid dates reported as valid")
	}
}

func TestDateIn(t *testing.T) {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Skip("no time zone data:", err)
	}
	got := NewDate(2026, time.July, 1).In(london)
	if got.Hour() != 0 || got.Location() != london || DateOf(got) != NewDate(2026, time.July, 1) {
		t.Errorf("In(london) = %v", got)
	}
}

func TestDateJSON(t *testing.T) {
	h := Holiday{Date: NewDate(2026, time.December, 28), Title: "Boxing Day", Notes: substituteNote, Source: SourceGOVUK}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"date":"2026-12-28","title":"Boxing Day","notes":"Substitute day","source":"gov.uk"}`
	if string(b) != want {
		t.Errorf("Marshal = %s, want %s", b, want)
	}
	var back Holiday
	if err := json.Unmarshal(b, &back); err != nil || back != h {
		t.Errorf("Unmarshal = %+v, %v", back, err)
	}
	if err := json.Unmarshal([]byte(`{"date":"2026-02-30"}`), &back); err == nil {
		t.Error("Unmarshal accepted 30 February")
	}
	if err := json.Unmarshal([]byte(`{"source":"guess"}`), &back); err == nil {
		t.Error("Unmarshal accepted an unknown source")
	}
}

func TestParseDivision(t *testing.T) {
	tests := map[string]Division{
		"england-and-wales": EnglandAndWales,
		"England":           EnglandAndWales,
		"wales":             EnglandAndWales,
		"EW":                EnglandAndWales,
		"Scotland":          Scotland,
		"sct":               Scotland,
		"NI":                NorthernIreland,
		"northern-ireland":  NorthernIreland,
	}
	for in, want := range tests {
		if got, err := ParseDivision(in); err != nil || got != want {
			t.Errorf("ParseDivision(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if got, err := ParseDivision(" scotland "); err != nil || got != Scotland {
		t.Errorf("ParseDivision does not trim spaces: %v, %v", got, err)
	}
	if _, err := ParseDivision("ireland"); err == nil {
		t.Error("ParseDivision(ireland) succeeded")
	}
	if Scotland.String() != "Scotland" || NorthernIreland.String() != "Northern Ireland" {
		t.Error("Division.String gives the wrong name")
	}
}
