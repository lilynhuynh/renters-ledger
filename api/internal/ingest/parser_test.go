package ingest

// Test files end in _test.go and live in the same package, so they can see unexported names.
// `go test ./...` finds every func TestXxx(t *testing.T). This is like JUnit, minus annotations.

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const header = "name,email,unit,move_in_date,monthly_premium_dollars\n"

// date is a test helper. Calling t.Helper() makes failures point at the caller's line.
func date(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("bad test date %q: %v", s, err)
	}
	return d
}

// TestParseRoster is a table-driven test: one slice of cases and one loop.
// Comparable to JUnit 5 @ParameterizedTest with @MethodSource.
func TestParseRoster(t *testing.T) {
	alice := Tenant{Name: "Alice Nguyen", Email: "alice@example.com", Unit: "1A",
		MoveInDate: date(t, "2024-01-15"), PremiumCents: 1250}
	bob := Tenant{Name: "Bob Smith", Email: "bob@example.com", Unit: "2B",
		MoveInDate: date(t, "2024-02-01"), PremiumCents: 1999} // 19.99 catches float bugs
	carmen := Tenant{Name: "Carmen Diaz", Email: "carmen@example.com", Unit: "3C",
		MoveInDate: date(t, "2024-03-10"), PremiumCents: 900} // "9" with no decimals

	aliceRow := "Alice Nguyen,alice@example.com,1A,2024-01-15,12.50\n"

	tests := []struct {
		name         string
		input        string
		wantTenants  []Tenant
		wantErrLines []int // line numbers of expected RowErrors, in order
		wantErr      error // non-nil only for whole-file failures
	}{
		{
			name: "valid file",
			input: header + aliceRow +
				"Bob Smith,bob@example.com,2B,2024-02-01,19.99\n" +
				"Carmen Diaz,carmen@example.com,3C,2024-03-10,9\n",
			wantTenants: []Tenant{alice, bob, carmen},
		},
		{
			name:         "bad email",
			input:        header + aliceRow + "Bob Smith,not-an-email,2B,2024-02-01,19.99\n",
			wantTenants:  []Tenant{alice},
			wantErrLines: []int{3},
		},
		{
			name:         "bad date",
			input:        header + aliceRow + "Bob Smith,bob@example.com,2B,02/01/2024,19.99\n",
			wantTenants:  []Tenant{alice},
			wantErrLines: []int{3},
		},
		{
			name:         "negative premium",
			input:        header + aliceRow + "Bob Smith,bob@example.com,2B,2024-02-01,-19.99\n",
			wantTenants:  []Tenant{alice},
			wantErrLines: []int{3},
		},
		{
			name:         "missing column value",
			input:        header + ",bob@example.com,2B,2024-02-01,19.99\n" + aliceRow,
			wantTenants:  []Tenant{alice},
			wantErrLines: []int{2},
		},
		{
			name:         "row with too few columns",
			input:        header + "Bob Smith,bob@example.com,2B\n" + aliceRow,
			wantTenants:  []Tenant{alice},
			wantErrLines: []int{2},
		},
		{
			// Line 3 is blank and skipped, so the bad email row is on line 4, not line 3.
			name:         "blank row is skipped but still counts as a line",
			input:        header + aliceRow + "\n" + "Bob Smith,not-an-email,2B,2024-02-01,19.99\n",
			wantTenants:  []Tenant{alice},
			wantErrLines: []int{4},
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: ErrEmptyRoster,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenants, rowErrs, err := ParseRoster(strings.NewReader(tt.input))

			// 1. Whole-file error. errors.Is is like `instanceof`/catch for sentinel errors.
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			// 2. Valid rows.
			if len(tenants) != len(tt.wantTenants) {
				t.Fatalf("got %d tenants, want %d: %+v", len(tenants), len(tt.wantTenants), tenants)
			}
			for i, want := range tt.wantTenants {
				got := tenants[i]
				// Compare time.Time with Equal, not ==, because == also compares the location pointer.
				if got.Name != want.Name || got.Email != want.Email || got.Unit != want.Unit ||
					!got.MoveInDate.Equal(want.MoveInDate) || got.PremiumCents != want.PremiumCents {
					t.Errorf("tenant[%d] = %+v, want %+v", i, got, want)
				}
			}

			// 3. Invalid rows. Only the line numbers are checked, so you can word Reason however you like.
			if len(rowErrs) != len(tt.wantErrLines) {
				t.Fatalf("got %d row errors, want %d: %+v", len(rowErrs), len(tt.wantErrLines), rowErrs)
			}
			for i, line := range tt.wantErrLines {
				if rowErrs[i].Line != line {
					t.Errorf("rowErrs[%d].Line = %d, want %d", i, rowErrs[i].Line, line)
				}
				if rowErrs[i].Reason == "" {
					t.Errorf("rowErrs[%d].Reason is empty; say what was wrong", i)
				}
			}
		})
	}
}

func TestRowErrorError(t *testing.T) {
	got := RowError{Line: 3, Reason: "bad date"}.Error()
	want := "line 3: bad date"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
