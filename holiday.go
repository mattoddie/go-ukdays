package ukdays

import (
	"fmt"
	"strings"
)

// Division is a part of the UK with its own set of bank holidays. The values
// match the keys GOV.UK uses in its bank holidays data.
type Division string

// The three divisions GOV.UK publishes bank holidays for.
const (
	EnglandAndWales Division = "england-and-wales"
	Scotland        Division = "scotland"
	NorthernIreland Division = "northern-ireland"
)

// Divisions returns all divisions, in the order GOV.UK lists them.
func Divisions() []Division {
	return []Division{EnglandAndWales, Scotland, NorthernIreland}
}

// ParseDivision returns the division named by s. It accepts the GOV.UK keys
// ("england-and-wales", "scotland", "northern-ireland") and the shorter
// forms "england", "wales", "ew", "sct" and "ni", in any case.
func ParseDivision(s string) (Division, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "england-and-wales", "england", "wales", "ew":
		return EnglandAndWales, nil
	case "scotland", "sct":
		return Scotland, nil
	case "northern-ireland", "ni":
		return NorthernIreland, nil
	}
	return "", fmt.Errorf("ukdays: unknown division %q: want england-and-wales, scotland or northern-ireland", s)
}

// String returns the division's name, such as "England and Wales".
func (d Division) String() string {
	switch d {
	case EnglandAndWales:
		return "England and Wales"
	case Scotland:
		return "Scotland"
	case NorthernIreland:
		return "Northern Ireland"
	}
	return string(d)
}

func (d Division) valid() bool {
	return d == EnglandAndWales || d == Scotland || d == NorthernIreland
}

// Source says where a holiday's date came from.
type Source int

const (
	// SourceGOVUK means the holiday is in the list published by GOV.UK.
	SourceGOVUK Source = iota + 1
	// SourceRules means the holiday was worked out from the standing rules,
	// for a year GOV.UK's list does not cover. One-off holidays announced
	// later, such as for a coronation, cannot be predicted.
	SourceRules
)

// String returns "gov.uk" or "rules".
func (s Source) String() string {
	switch s {
	case SourceGOVUK:
		return "gov.uk"
	case SourceRules:
		return "rules"
	}
	return fmt.Sprintf("Source(%d)", int(s))
}

// MarshalText implements [encoding.TextMarshaler], so a Source encodes as a
// JSON string such as "gov.uk".
func (s Source) MarshalText() ([]byte, error) {
	switch s {
	case SourceGOVUK, SourceRules:
		return []byte(s.String()), nil
	}
	return nil, fmt.Errorf("ukdays: invalid source %d", int(s))
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (s *Source) UnmarshalText(text []byte) error {
	switch string(text) {
	case "gov.uk":
		*s = SourceGOVUK
	case "rules":
		*s = SourceRules
	default:
		return fmt.Errorf("ukdays: invalid source %q", text)
	}
	return nil
}

// Holiday is a bank holiday as it is observed: when the day itself falls at
// a weekend, Date is the substitute weekday and Notes says so.
type Holiday struct {
	Date   Date   `json:"date"`
	Title  string `json:"title"`
	Notes  string `json:"notes,omitempty"`
	Source Source `json:"source"`
}

// substituteNote is the note GOV.UK gives a holiday moved off a weekend.
const substituteNote = "Substitute day"

// IsSubstitute reports whether the holiday was moved from a weekend.
func (h Holiday) IsSubstitute() bool {
	return h.Notes == substituteNote
}
