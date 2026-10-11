package ukdays

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleJSON = `{
  "england-and-wales": {"division": "england-and-wales", "events": [
    {"title": "Boxing Day", "date": "2040-12-26", "notes": "", "bunting": true},
    {"title": "Christmas Day", "date": "2040-12-25", "notes": "", "bunting": true}
  ]},
  "scotland": {"division": "scotland", "events": [
    {"title": "Christmas Day", "date": "2040-12-25", "notes": "", "bunting": true}
  ]},
  "northern-ireland": {"division": "northern-ireland", "events": [
    {"title": "Christmas Day", "date": "2040-12-25", "notes": "", "bunting": true}
  ]}
}`

func TestLoad(t *testing.T) {
	data, err := Load(strings.NewReader(sampleJSON))
	if err != nil {
		t.Fatal(err)
	}
	if first, last := data.Years(); first != 2040 || last != 2040 {
		t.Errorf("Years = %d, %d", first, last)
	}
	c, err := data.Calendar(EnglandAndWales)
	if err != nil {
		t.Fatal(err)
	}
	hs := c.Holidays(2040)
	if len(hs) != 2 || hs[0].Title != "Christmas Day" || hs[0].Source != SourceGOVUK {
		t.Errorf("Holidays(2040) = %+v; want two, sorted, from GOV.UK", hs)
	}
	// Years outside the loaded data still come from the rules.
	if s, _ := c.Source(2039); s != SourceRules {
		t.Errorf("Source(2039) = %v, want rules", s)
	}
	if _, err := data.Calendar("wales"); err == nil {
		t.Error("Calendar accepted an unknown division")
	}
}

func TestLoadErrors(t *testing.T) {
	tests := map[string]string{
		"not JSON":         `<html>`,
		"missing division": `{"england-and-wales": {"events": [{"title": "x", "date": "2040-01-01"}]}}`,
		"bad date":         strings.Replace(sampleJSON, "2040-12-26", "2040-12-32", 1),
		"no title":         strings.Replace(sampleJSON, `"title": "Boxing Day"`, `"title": ""`, 1),
		"duplicate":        strings.Replace(sampleJSON, "2040-12-26", "2040-12-25", 1),
		"no events": strings.Replace(sampleJSON,
			`[
    {"title": "Christmas Day", "date": "2040-12-25", "notes": "", "bunting": true}
  ]},
  "northern-ireland"`, `[]},
  "northern-ireland"`, 1),
		"too large": `{"x":"` + strings.Repeat("a", maxDataSize) + `"}`,
	}
	for name, in := range tests {
		if _, err := Load(strings.NewReader(in)); err == nil {
			t.Errorf("%s: Load succeeded, want error", name)
		}
	}
}

func TestEmbedded(t *testing.T) {
	first, last := Embedded().Years()
	if first > 2019 || last < 2028 {
		t.Errorf("embedded data covers %d to %d", first, last)
	}
}

func TestFetch(t *testing.T) {
	var gotAgent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAgent = r.Header.Get("User-Agent")
		if r.URL.Path != "/bank-holidays.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleJSON))
	}))
	defer srv.Close()

	client := srv.Client()
	client.Transport = rewriteTransport{base: client.Transport, target: srv.URL}

	data, err := Fetch(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if first, _ := data.Years(); first != 2040 {
		t.Errorf("fetched data starts in %d", first)
	}
	if !strings.HasPrefix(gotAgent, "go-ukdays") {
		t.Errorf("User-Agent = %q", gotAgent)
	}
}

func TestFetchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = rewriteTransport{base: client.Transport, target: srv.URL}

	if _, err := Fetch(context.Background(), client); err == nil || !strings.Contains(err.Error(), "503") {
		t.Errorf("Fetch error = %v, want a 503", err)
	}
}

// rewriteTransport sends every request to a test server instead of GOV.UK.
type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (rt rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	u := *r.URL
	u.Scheme, u.Host = "http", strings.TrimPrefix(rt.target, "http://")
	r2 := r.Clone(r.Context())
	r2.URL = &u
	return rt.base.RoundTrip(r2)
}
