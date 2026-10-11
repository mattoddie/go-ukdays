package ukdays_test

import (
	"fmt"
	"time"

	"github.com/mattoddie/go-ukdays"
)

func Example() {
	cal, err := ukdays.New(ukdays.EnglandAndWales)
	if err != nil {
		panic(err)
	}

	christmasEve := ukdays.NewDate(2026, time.December, 24)
	fmt.Println(cal.AddWorkingDays(christmasEve, 1))
	fmt.Println(cal.WorkingDaysBetween(christmasEve, ukdays.NewDate(2027, time.January, 1)))
	// Output:
	// 2026-12-29
	// 4
}

func ExampleCalendar_IsWorkingDay() {
	cal, _ := ukdays.New(ukdays.Scotland)

	for _, s := range []string{"2027-01-04", "2027-01-05"} {
		d, _ := ukdays.ParseDate(s)
		fmt.Println(d, cal.IsWorkingDay(d))
	}
	// Output:
	// 2027-01-04 false
	// 2027-01-05 true
}

func ExampleCalendar_Holidays() {
	cal, _ := ukdays.New(ukdays.NorthernIreland)

	for _, h := range cal.Holidays(2026)[:3] {
		fmt.Println(h.Date, h.Title)
	}
	// Output:
	// 2026-01-01 New Year’s Day
	// 2026-03-17 St Patrick’s Day
	// 2026-04-03 Good Friday
}

func ExampleCalendar_HolidayOn() {
	cal, _ := ukdays.New(ukdays.EnglandAndWales)

	h, ok := cal.HolidayOn(ukdays.NewDate(2027, time.December, 27))
	fmt.Println(ok, h.Title, h.IsSubstitute())
	// Output:
	// true Christmas Day true
}

func ExampleCalendar_Source() {
	cal, _ := ukdays.New(ukdays.EnglandAndWales)

	for _, year := range []int{2026, 2035, 2000} {
		src, ok := cal.Source(year)
		fmt.Println(year, src, ok)
	}
	// Output:
	// 2026 gov.uk true
	// 2035 rules true
	// 2000 Source(0) false
}
