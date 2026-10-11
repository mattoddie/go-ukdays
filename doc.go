// Package ukdays works with UK bank holidays and working days: whether a
// date is a working day, the next or previous working day, adding working
// days to a date, and counting working days between two dates.
//
// England and Wales, Scotland and Northern Ireland each have their own bank
// holidays, so every question is asked of a [Calendar] for one [Division]:
//
//	cal, err := ukdays.New(ukdays.Scotland)
//	if err != nil {
//		// only possible for an unknown division
//	}
//	due := cal.AddWorkingDays(ukdays.NewDate(2026, time.December, 24), 3)
//	fmt.Println(due) // 2026-12-31
//
// # Where the dates come from
//
// The package includes a copy of the list GOV.UK publishes at [GOVUKURL],
// which covers a range of past and future years. [Fetch] downloads the
// latest list, and [Load] reads one from any source.
//
// For years outside the list, from [MinYear] on, holidays are worked out
// from the standing rules: Easter, the fixed-date and Monday holidays of each
// division, and the substitute weekdays used when a holiday falls at a
// weekend. One-off holidays, such as those for coronations and jubilees,
// cannot be predicted and appear only once GOV.UK publishes them.
// [Holiday.Source] and [Calendar.Source] say which applies.
//
// Dates are plain calendar dates with no time zone; see [Date].
package ukdays
