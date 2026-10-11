// Command ukdays answers questions about UK bank holidays and working days.
//
// Usage:
//
//	ukdays [flags] <command> [arguments]
//
// Commands:
//
//	check [DATE]       say whether DATE is a working day (exit status 1 if not)
//	next [DATE]        print the first working day after DATE
//	previous [DATE]    print the last working day before DATE
//	add DATE N         print the date N working days after DATE (negative N counts back)
//	between FROM TO    count working days from FROM up to, but not including, TO
//	holidays [YEAR]    list the bank holidays in YEAR
//
// DATE is in YYYY-MM-DD form and defaults to today in the UK.
//
// Flags:
//
//	-division string   england-and-wales, scotland or northern-ireland
//	                   (default from $UKDAYS_DIVISION, else england-and-wales)
//	-live              download the latest list from GOV.UK instead of using the built-in copy
//	-json              print JSON
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
	_ "time/tzdata" // so "today" works where the system has no zone data

	"github.com/mattoddie/go-ukdays"
)

// errNotWorkingDay makes "check" exit with status 1 without printing an error.
var errNotWorkingDay = errors.New("not a working day")

type options struct {
	division string
	live     bool
	json     bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, time.Now))
}

func run(args []string, stdout, stderr io.Writer, now func() time.Time) int {
	var opts options
	fs := flag.NewFlagSet("ukdays", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }
	defaultDivision := os.Getenv("UKDAYS_DIVISION")
	if defaultDivision == "" {
		defaultDivision = string(ukdays.EnglandAndWales)
	}
	fs.StringVar(&opts.division, "division", defaultDivision, "")
	fs.BoolVar(&opts.live, "live", false, "")
	fs.BoolVar(&opts.json, "json", false, "")

	// Flags may come before or after the command name.
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return 2
	}
	cmd := fs.Arg(0)
	if err := fs.Parse(fs.Args()[1:]); err != nil {
		return 2
	}
	cmdArgs := fs.Args()

	err := execute(cmd, cmdArgs, opts, stdout, now)
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errNotWorkingDay):
		return 1
	case errors.As(err, new(usageError)):
		fmt.Fprintf(stderr, "ukdays: %v\n\n", err)
		fs.Usage()
		return 2
	default:
		fmt.Fprintf(stderr, "ukdays: %v\n", err)
		return 1
	}
}

type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

func execute(cmd string, args []string, opts options, out io.Writer, now func() time.Time) error {
	div, err := ukdays.ParseDivision(opts.division)
	if err != nil {
		return usageError{err.Error()}
	}
	cal, err := calendar(div, opts.live)
	if err != nil {
		return err
	}
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		return err
	}
	today := ukdays.DateOf(now().In(london))

	switch cmd {
	case "check":
		d, err := dateArg(args, 0, today, 1)
		if err != nil {
			return err
		}
		h, isHoliday := cal.HolidayOn(d)
		working := cal.IsWorkingDay(d)
		if opts.json {
			result := struct {
				Date       ukdays.Date     `json:"date"`
				Division   ukdays.Division `json:"division"`
				WorkingDay bool            `json:"working_day"`
				Holiday    *ukdays.Holiday `json:"holiday,omitempty"`
			}{Date: d, Division: div, WorkingDay: working}
			if isHoliday {
				result.Holiday = &h
			}
			if err := writeJSON(out, result); err != nil {
				return err
			}
		} else {
			switch {
			case working:
				fmt.Fprintf(out, "%s (%s) is a working day in %s\n", d, d.Weekday(), div)
			case isHoliday:
				fmt.Fprintf(out, "%s (%s) is not a working day in %s: %s\n", d, d.Weekday(), div, describe(h))
			default:
				fmt.Fprintf(out, "%s is a %s\n", d, d.Weekday())
			}
		}
		if !working {
			return errNotWorkingDay
		}
		return nil

	case "next", "previous":
		d, err := dateArg(args, 0, today, 1)
		if err != nil {
			return err
		}
		result := cal.NextWorkingDay(d)
		if cmd == "previous" {
			result = cal.PreviousWorkingDay(d)
		}
		return printDate(out, result, opts.json)

	case "add":
		if len(args) != 2 {
			return usageError{"add needs a date and a number of working days"}
		}
		d, err := ukdays.ParseDate(args[0])
		if err != nil {
			return usageError{err.Error()}
		}
		n, err := strconv.Atoi(args[1])
		if err != nil {
			return usageError{fmt.Sprintf("invalid number of days %q", args[1])}
		}
		return printDate(out, cal.AddWorkingDays(d, n), opts.json)

	case "between":
		if len(args) != 2 {
			return usageError{"between needs two dates"}
		}
		from, err := ukdays.ParseDate(args[0])
		if err != nil {
			return usageError{err.Error()}
		}
		to, err := ukdays.ParseDate(args[1])
		if err != nil {
			return usageError{err.Error()}
		}
		n := cal.WorkingDaysBetween(from, to)
		if opts.json {
			return writeJSON(out, map[string]int{"working_days": n})
		}
		fmt.Fprintln(out, n)
		return nil

	case "holidays":
		if len(args) > 1 {
			return usageError{"holidays takes at most one year"}
		}
		year := today.Year
		if len(args) == 1 {
			if year, err = strconv.Atoi(args[0]); err != nil {
				return usageError{fmt.Sprintf("invalid year %q", args[0])}
			}
		}
		hs := cal.Holidays(year)
		if opts.json {
			if hs == nil {
				hs = []ukdays.Holiday{}
			}
			return writeJSON(out, hs)
		}
		if len(hs) == 0 {
			return fmt.Errorf("no bank holidays known for %d (the first year covered is %d)", year, ukdays.MinYear)
		}
		if src, _ := cal.Source(year); src == ukdays.SourceRules {
			fmt.Fprintf(out, "# %d is not in GOV.UK's list, so these dates come from the standing rules.\n", year)
		}
		for _, h := range hs {
			fmt.Fprintf(out, "%s  %s  %s\n", h.Date, h.Date.Weekday().String()[:3], describe(h))
		}
		return nil
	}
	return usageError{fmt.Sprintf("unknown command %q", cmd)}
}

func calendar(div ukdays.Division, live bool) (*ukdays.Calendar, error) {
	if !live {
		return ukdays.New(div)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	data, err := ukdays.Fetch(ctx, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	return data.Calendar(div)
}

// dateArg returns args[i] as a date, or def when there are no arguments.
func dateArg(args []string, i int, def ukdays.Date, maxArgs int) (ukdays.Date, error) {
	if len(args) > maxArgs {
		return ukdays.Date{}, usageError{"too many arguments"}
	}
	if len(args) <= i {
		return def, nil
	}
	d, err := ukdays.ParseDate(args[i])
	if err != nil {
		return ukdays.Date{}, usageError{err.Error()}
	}
	return d, nil
}

func describe(h ukdays.Holiday) string {
	if h.IsSubstitute() {
		return h.Title + " (substitute day)"
	}
	return h.Title
}

func printDate(out io.Writer, d ukdays.Date, asJSON bool) error {
	if asJSON {
		return writeJSON(out, map[string]ukdays.Date{"date": d})
	}
	fmt.Fprintln(out, d)
	return nil
}

func writeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

const usage = `Usage: ukdays [flags] <command> [arguments]

Commands:
  check [DATE]       say whether DATE is a working day (exit status 1 if not)
  next [DATE]        print the first working day after DATE
  previous [DATE]    print the last working day before DATE
  add DATE N         print the date N working days after DATE (negative N counts back)
  between FROM TO    count working days from FROM up to, but not including, TO
  holidays [YEAR]    list the bank holidays in YEAR

DATE is in YYYY-MM-DD form and defaults to today in the UK.

Flags:
  -division string   england-and-wales, scotland or northern-ireland
                     (default from $UKDAYS_DIVISION, else england-and-wales)
  -live              download the latest list from GOV.UK instead of using the built-in copy
  -json              print JSON
`
