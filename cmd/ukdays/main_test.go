package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// fixedNow is a Sunday evening in London, so "today" is 11 October 2026.
func fixedNow() time.Time {
	return time.Date(2026, time.October, 11, 20, 0, 0, 0, time.UTC)
}

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	t.Setenv("UKDAYS_DIVISION", "")
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut, fixedNow)
	return code, out.String(), errOut.String()
}

func TestCommands(t *testing.T) {
	tests := []struct {
		args []string
		code int
		want string
	}{
		{[]string{"check", "2026-12-24"}, 0, "2026-12-24 (Thursday) is a working day in England and Wales\n"},
		{[]string{"check", "2026-12-28"}, 1, "2026-12-28 (Monday) is not a working day in England and Wales: Boxing Day (substitute day)\n"},
		{[]string{"check", "2026-12-26"}, 1, "2026-12-26 is a Saturday\n"},
		{[]string{"check"}, 1, "2026-10-11 is a Sunday\n"},
		{[]string{"-division", "scotland", "check", "2027-01-04"}, 1, "2027-01-04 (Monday) is not a working day in Scotland: 2nd January (substitute day)\n"},
		{[]string{"check", "-division", "ni", "2026-03-17"}, 1, "2026-03-17 (Tuesday) is not a working day in Northern Ireland: St Patrick’s Day\n"},
		{[]string{"next"}, 0, "2026-10-12\n"},
		{[]string{"next", "2026-12-24"}, 0, "2026-12-29\n"},
		{[]string{"previous", "2026-12-29"}, 0, "2026-12-24\n"},
		{[]string{"add", "2026-12-24", "4"}, 0, "2027-01-04\n"},
		{[]string{"add", "2027-01-04", "-4"}, 0, "2026-12-24\n"},
		{[]string{"between", "2026-01-01", "2027-01-01"}, 0, "253\n"},
		{[]string{"-json", "next", "2026-12-24"}, 0, "{\n  \"date\": \"2026-12-29\"\n}\n"},
	}
	for _, tt := range tests {
		code, out, errOut := runCLI(t, tt.args...)
		if code != tt.code || out != tt.want {
			t.Errorf("ukdays %s: exit %d, output %q, stderr %q; want exit %d, output %q",
				strings.Join(tt.args, " "), code, out, errOut, tt.code, tt.want)
		}
	}
}

func TestHolidaysCommand(t *testing.T) {
	code, out, _ := runCLI(t, "holidays", "2026")
	if code != 0 || !strings.HasPrefix(out, "2026-01-01  Thu  New Year’s Day\n") || strings.Count(out, "\n") != 8 {
		t.Errorf("holidays 2026: exit %d, output:\n%s", code, out)
	}

	code, out, _ = runCLI(t, "holidays", "2040")
	if code != 0 || !strings.HasPrefix(out, "# 2040 is not in GOV.UK's list") {
		t.Errorf("holidays 2040: exit %d, output:\n%s", code, out)
	}

	code, _, errOut := runCLI(t, "holidays", "1990")
	if code != 1 || !strings.Contains(errOut, "no bank holidays known for 1990") {
		t.Errorf("holidays 1990: exit %d, stderr %q", code, errOut)
	}

	code, out, _ = runCLI(t, "-json", "holidays", "2026")
	var hs []map[string]string
	if err := json.Unmarshal([]byte(out), &hs); code != 0 || err != nil || len(hs) != 8 || hs[0]["source"] != "gov.uk" {
		t.Errorf("holidays -json: exit %d, err %v, output:\n%s", code, err, out)
	}
}

func TestCheckJSON(t *testing.T) {
	code, out, _ := runCLI(t, "-json", "check", "2026-12-25")
	var got struct {
		WorkingDay bool `json:"working_day"`
		Holiday    *struct {
			Title string `json:"title"`
		} `json:"holiday"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil || code != 1 || got.WorkingDay || got.Holiday == nil || got.Holiday.Title != "Christmas Day" {
		t.Errorf("check -json: exit %d, output %s", code, out)
	}
}

func TestDivisionFromEnvironment(t *testing.T) {
	t.Setenv("UKDAYS_DIVISION", "scotland")
	var out, errOut bytes.Buffer
	if code := run([]string{"check", "2026-08-03"}, &out, &errOut, fixedNow); code != 1 {
		t.Errorf("Scotland's summer bank holiday: exit %d, output %q", code, out.String())
	}
}

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"nonsense"},
		{"add", "2026-12-24"},
		{"add", "2026-12-24", "x"},
		{"between", "2026-12-24"},
		{"check", "24/12/2026"},
		{"check", "2026-12-24", "2026-12-25"},
		{"holidays", "twenty"},
		{"-division", "wales-and-england", "next"},
		{"-unknown"},
	} {
		if code, _, errOut := runCLI(t, args...); code != 2 || errOut == "" {
			t.Errorf("ukdays %s: exit %d, stderr %q; want exit 2 and a message", strings.Join(args, " "), code, errOut)
		}
	}
}
