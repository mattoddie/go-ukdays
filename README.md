# go-ukdays

[![Test](https://github.com/mattoddie/go-ukdays/actions/workflows/test.yaml/badge.svg)](https://github.com/mattoddie/go-ukdays/actions/workflows/test.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mattoddie/go-ukdays.svg)](https://pkg.go.dev/github.com/mattoddie/go-ukdays)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/mattoddie/go-ukdays/badge)](https://scorecard.dev/viewer/?uri=github.com/mattoddie/go-ukdays)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

UK bank holidays and working days for Go, with a command-line tool. It
answers questions such as "is this a working day?", "what is five working
days after Christmas Eve?" and "how many working days are there in March?"
for England and Wales, Scotland and Northern Ireland.

- Uses the official list from [GOV.UK](https://www.gov.uk/bank-holidays),
  built in so it works offline, with an option to fetch the latest.
- Works out holidays for years GOV.UK hasn't published yet from the standing
  rules, including Easter and substitute days. The rules are tested against
  every published year.
- No dependencies outside the standard library.

## Installation

```sh
go get github.com/mattoddie/go-ukdays
```

Requires Go 1.22 or later.

## Usage

```go
import "github.com/mattoddie/go-ukdays"

cal, err := ukdays.New(ukdays.EnglandAndWales)
if err != nil {
	// only possible for an unknown division
}

christmasEve := ukdays.NewDate(2026, time.December, 24)

cal.IsWorkingDay(christmasEve)        // true
cal.NextWorkingDay(christmasEve)      // 2026-12-29: Christmas Day, the weekend and Boxing Day's substitute
cal.AddWorkingDays(christmasEve, 4)   // 2027-01-04
cal.WorkingDaysBetween(christmasEve,  // 4: the 24th, 29th, 30th and 31st
	ukdays.NewDate(2027, time.January, 1))

for _, h := range cal.Holidays(2027) {
	fmt.Println(h.Date, h.Title, h.IsSubstitute())
}
```

`WorkingDaysBetween` counts from the first date up to, but not including, the
second. `AddWorkingDays` takes negative numbers to count back.

### Divisions

Each part of the UK has its own bank holidays, so a calendar is for one
division:

| Division           | Constant          | Holidays the others don't have                   |
| ------------------ | ----------------- | ------------------------------------------------ |
| England and Wales  | `EnglandAndWales` |                                                  |
| Scotland           | `Scotland`        | 2nd January, St Andrew's Day; summer holiday in early August; no Easter Monday |
| Northern Ireland   | `NorthernIreland` | St Patrick's Day, Battle of the Boyne            |

`ParseDivision` accepts the GOV.UK names (`england-and-wales`, `scotland`,
`northern-ireland`) and short forms such as `ni`.

### Where the dates come from

| Years              | Source                                                   |
| ------------------ | -------------------------------------------------------- |
| In GOV.UK's list   | The published list (currently 2019 to 2028)              |
| 2007 onwards       | The standing rules, plus known one-offs from 2011 and 2012 |
| Before 2007        | None: every weekday is a working day                     |

The rules can't predict one-off holidays such as a coronation; those appear
once GOV.UK publishes them. `Holiday.Source` and `Calendar.Source` say which
source a date came from. The built-in copy is checked weekly and updated in
new releases. To use the live list instead:

```go
data, err := ukdays.Fetch(ctx, http.DefaultClient)
if err != nil {
	// handle error
}
cal, err := data.Calendar(ukdays.Scotland)
```

### Dates

`ukdays.Date` is a plain calendar date with no time zone. Make one with
`NewDate`, `ParseDate("2026-12-25")` or `DateOf(t)`, which takes the date of a
`time.Time` in its own location. `Today(loc)` gives today's date; for UK
working days that's usually `Europe/London`. Dates encode to JSON as
`"2026-12-25"`.

## Command-line tool

```sh
go install github.com/mattoddie/go-ukdays/cmd/ukdays@latest
```

```console
$ ukdays check 2026-12-28
2026-12-28 (Monday) is not a working day in England and Wales: Boxing Day (substitute day)

$ ukdays add 2026-12-24 4
2027-01-04

$ ukdays between 2026-01-01 2027-01-01
253

$ ukdays -division scotland holidays 2027
2027-01-01  Fri  New Year’s Day
2027-01-04  Mon  2nd January (substitute day)
...
```

| Command           | Prints                                                          |
| ----------------- | --------------------------------------------------------------- |
| `check [DATE]`    | Whether DATE is a working day; exits with status 1 if it isn't  |
| `next [DATE]`     | The first working day after DATE                                |
| `previous [DATE]` | The last working day before DATE                                |
| `add DATE N`      | The date N working days after DATE; negative N counts back      |
| `between FROM TO` | Working days from FROM up to, but not including, TO             |
| `holidays [YEAR]` | The bank holidays in YEAR                                       |

DATE defaults to today in the UK. Flags go before the arguments:
`-division` (or the `UKDAYS_DIVISION` environment variable), `-json` for JSON
output, and `-live` to fetch the latest list from GOV.UK.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Report security issues as described
in [SECURITY.md](SECURITY.md).

## Licence

[MIT](LICENSE). Bank holiday data is from [GOV.UK](https://www.gov.uk/bank-holidays)
and is available under the
[Open Government Licence v3.0](https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/).
