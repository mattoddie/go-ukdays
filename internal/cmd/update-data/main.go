// Command update-data downloads GOV.UK's bank holidays and, if they parse,
// writes them to the package's embedded data file. Run it with
// `go generate ./...` from the repository root.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mattoddie/go-ukdays"
)

func main() {
	out := flag.String("o", "data/bank-holidays.json", "file to write")
	flag.Parse()
	if err := update(*out); err != nil {
		fmt.Fprintln(os.Stderr, "update-data:", err)
		os.Exit(1)
	}
}

func update(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ukdays.GOVUKURL, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", ukdays.GOVUKURL, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	// Refuse anything the package couldn't load.
	data, err := ukdays.Load(bytes.NewReader(body))
	if err != nil {
		return err
	}

	old, err := os.ReadFile(path)
	if err == nil && bytes.Equal(old, body) {
		fmt.Println("bank holidays unchanged")
		return nil
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return err
	}
	first, last := data.Years()
	fmt.Printf("wrote %s: bank holidays for %d to %d\n", path, first, last)
	return nil
}
