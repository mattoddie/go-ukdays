package ukdays

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
)

//go:generate go run ./internal/cmd/update-data -o data/bank-holidays.json

// GOVUKURL is where GOV.UK publishes the bank holidays for each division.
const GOVUKURL = "https://www.gov.uk/bank-holidays.json"

// maxDataSize limits how much of a response [Fetch] and [Load] will read.
// GOV.UK's file is around 20 KB.
const maxDataSize = 1 << 20

//go:embed data/bank-holidays.json
var embeddedJSON []byte

// Data is a set of published bank holidays for every division, as loaded
// from GOV.UK's JSON. Use [Data.Calendar] to work with one division.
type Data struct {
	divisions map[Division][]Holiday
}

// Embedded returns the copy of GOV.UK's data built into the package. It is
// refreshed in each release; use [Fetch] for the latest list.
var Embedded = sync.OnceValue(func() *Data {
	data, err := parse(embeddedJSON)
	if err != nil {
		panic("ukdays: embedded data is invalid: " + err.Error())
	}
	return data
})

// Load reads bank holidays in the format GOV.UK publishes at [GOVUKURL].
// The data must include all three divisions.
func Load(r io.Reader) (*Data, error) {
	b, err := io.ReadAll(io.LimitReader(r, maxDataSize+1))
	if err != nil {
		return nil, fmt.Errorf("ukdays: reading bank holidays: %w", err)
	}
	if len(b) > maxDataSize {
		return nil, fmt.Errorf("ukdays: bank holidays data is larger than %d bytes", maxDataSize)
	}
	return parse(b)
}

// Fetch downloads the latest bank holidays from GOV.UK. If client is nil,
// [http.DefaultClient] is used; set a timeout on ctx or the client.
func Fetch(ctx context.Context, client *http.Client) (*Data, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GOVUKURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("ukdays: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-ukdays (+https://github.com/mattoddie/go-ukdays)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ukdays: fetching bank holidays: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ukdays: fetching bank holidays: %s", resp.Status)
	}
	return Load(resp.Body)
}

// Calendar returns a calendar for div that uses this data for the years it
// covers and the standing rules for other years from [MinYear] on.
func (d *Data) Calendar(div Division) (*Calendar, error) {
	if !div.valid() {
		return nil, fmt.Errorf("ukdays: unknown division %q", string(div))
	}
	return newCalendar(div, d.divisions[div]), nil
}

// Years returns the first and last years the data covers.
func (d *Data) Years() (first, last int) {
	for _, hs := range d.divisions {
		if len(hs) == 0 {
			continue
		}
		if f := hs[0].Date.Year; first == 0 || f < first {
			first = f
		}
		if l := hs[len(hs)-1].Date.Year; l > last {
			last = l
		}
	}
	return first, last
}

type govukFile map[string]struct {
	Division string `json:"division"`
	Events   []struct {
		Title string `json:"title"`
		Date  string `json:"date"`
		Notes string `json:"notes"`
	} `json:"events"`
}

func parse(b []byte) (*Data, error) {
	var file govukFile
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, fmt.Errorf("ukdays: parsing bank holidays: %w", err)
	}
	data := &Data{divisions: make(map[Division][]Holiday, 3)}
	for _, div := range Divisions() {
		entry, ok := file[string(div)]
		if !ok {
			return nil, fmt.Errorf("ukdays: bank holidays data has no %q division", string(div))
		}
		if len(entry.Events) == 0 {
			return nil, fmt.Errorf("ukdays: bank holidays data has no events for %q", string(div))
		}
		hs := make([]Holiday, 0, len(entry.Events))
		for _, e := range entry.Events {
			d, err := ParseDate(e.Date)
			if err != nil {
				return nil, fmt.Errorf("ukdays: %s: %w", div, err)
			}
			if e.Title == "" {
				return nil, fmt.Errorf("ukdays: %s: holiday on %s has no title", div, d)
			}
			hs = append(hs, Holiday{Date: d, Title: e.Title, Notes: e.Notes, Source: SourceGOVUK})
		}
		slices.SortFunc(hs, func(a, b Holiday) int { return a.Date.Compare(b.Date) })
		for i := 1; i < len(hs); i++ {
			if hs[i].Date == hs[i-1].Date {
				return nil, fmt.Errorf("ukdays: %s: two holidays on %s", div, hs[i].Date)
			}
		}
		data.divisions[div] = hs
	}
	return data, nil
}
