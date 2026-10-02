// Command rostercli parses a tenant roster CSV and prints a summary (Day 1).
//
// Usage (from api/):  go run ./cmd/rostercli internal/ingest/testdata/roster_bad_rows.csv
//
// A `package main` with `func main()` is a runnable program, like a class with
// `public static void main`. Each folder under cmd/ builds into its own binary.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/lilynhuynh/renters-ledger/api/internal/ingest"
)

func main() {
	// os.Args[0] is the program name, so the file path is os.Args[1].
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: rostercli <path-to-roster.csv>")
		os.Exit(2)
	}

	// main stays tiny and run returns an error. That keeps the logic testable and
	// replaces Java's "throw and let main crash".
	if err := run(os.Args[1], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run opens the file, parses it, and writes the summary to out.
// io.Writer is an interface (like java.io.Writer), so tests can pass a bytes.Buffer.
func run(path string, out io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		// %w wraps the error so callers can still errors.Is() it, like a Java exception cause.
		return fmt.Errorf("open roster: %w", err)
	}
	// defer runs when run() returns, similar to try-with-resources.
	defer f.Close()

	tenants, rowErrs, err := ingest.ParseRoster(f)
	if err != nil {
		return fmt.Errorf("parse roster %s: %w", path, err)
	}

	var totalCents int64
	for _, t := range tenants {
		totalCents += t.PremiumCents
	}

	fmt.Fprintf(out, "roster:        %s\n", path)
	fmt.Fprintf(out, "valid rows:    %d\n", len(tenants))
	fmt.Fprintf(out, "invalid rows:  %d\n", len(rowErrs))
	for _, re := range rowErrs {
		fmt.Fprintf(out, "  - %s\n", re.Error())
	}
	fmt.Fprintf(out, "total monthly premium: %s\n", formatDollars(totalCents))
	return nil
}

// formatDollars renders cents as dollars using integer math only: 123456 -> "$1234.56".
func formatDollars(cents int64) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}
