// Package ingest turns raw tenant roster files into validated Tenant values.
// It only parses and validates. It does not save anything (no store import).
package ingest

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrEmptyRoster is returned when the input has no header line at all.
// A package-level error value like this is a "sentinel error". Callers check it with
// errors.Is(err, ingest.ErrEmptyRoster), similar to catching a specific exception type in Java.
var ErrEmptyRoster = errors.New("roster is empty: no header row")

// Tenant is one valid roster row. A plain struct like this is like a Java record or POJO.
// Fields starting with a capital letter are exported (public). Lowercase means package-private.
type Tenant struct {
	Name       string
	Email      string
	Unit       string
	MoveInDate time.Time // date only, parsed from YYYY-MM-DD and stored in UTC

	// PremiumCents is the monthly premium in integer cents ($12.50 -> 1250).
	// Money is never stored as float64, because binary floats cannot represent 0.10
	// exactly and rounding errors pile up when you add amounts. In Java you would
	// reach for BigDecimal. Here an int64 of cents is simpler and just as exact.
	PremiumCents int64
}

// RowError describes one invalid roster row. It is not fatal: parsing continues past it.
type RowError struct {
	Line   int    // 1-based line number in the file; the header is line 1
	Reason string // human-readable, e.g. "invalid email \"bob@\""
}

// Error makes RowError satisfy the built-in `error` interface. Go interfaces are
// satisfied implicitly: there is no `implements error`, the method alone is enough.
//
// TODO(Day1): return a string of the form "line <Line>: <Reason>",
// e.g. RowError{Line: 3, Reason: "bad date"}.Error() == "line 3: bad date".
// Hint: fmt.Sprintf.
func (e RowError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Reason)
}

// ParseRoster reads a CSV roster with the header
//
//	name,email,unit,move_in_date,monthly_premium_dollars
//
// and returns:
//   - []Tenant:   every valid row, in file order.
//   - []RowError: every invalid row, with its 1-based line number (header = line 1).
//     A row is invalid if it has the wrong number of columns, an empty value, an email
//     without a name and domain around the "@", a date that is not YYYY-MM-DD,
//     or a premium that is not a non-negative dollar amount with at most 2 decimals.
//   - error: only when the input cannot be read at all (an I/O error, or ErrEmptyRoster
//     when there is no header line). Bad rows NEVER produce this error.
//
// Blank lines are skipped and are not errors, but they still count toward line numbers.
// So a bad row after a blank line on line 3 is reported as line 4.
//
// TODO(Day1): implement. Hints:
//   - encoding/csv: csv.NewReader(r). Set FieldsPerRecord = -1 so a short row becomes
//     a RowError instead of stopping the whole read.
//   - (*csv.Reader).FieldPos(0) gives the real line number of the current record,
//     even when the reader skipped blank lines. A hand-written counter gets this wrong.
//   - time.Parse("2006-01-02", s). Go layouts use that exact reference date, not "yyyy-MM-dd".
//   - Do not use strconv.ParseFloat for money: ParseFloat("19.99")*100 truncates to 1998.
//     Split on "." and parse the whole-dollar and cents parts as integers.
//   - Consider small helpers (parseEmail, parseCents) in this file, each with its own error.
func ParseRoster(r io.Reader) ([]Tenant, []RowError, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // allows rows with wrong col count

	var tenants []Tenant
	var rowErr []RowError
	_, err := cr.Read()
	if err != nil || err == io.EOF {
		return nil, nil, fmt.Errorf("Empty file or invalid file: %w", ErrEmptyRoster)
	}

	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("Read Roster: %w", err)
		}
		line, _ := cr.FieldPos(0) // Want line number, _ = blank id
		if len(record) != 5 {
			rowErr = append(rowErr, RowError{Line: line, Reason: "Invalid record length, missing columns"})
			continue
		}
		name, email, unit, moveInDate, premium := record[0], record[1], record[2], record[3], record[4]
		name, nameErr := parseName(name)
		email, emailErr := parseEmail(email)
		unit, unitErr := parseUnit(unit)
		date, dateErr := parseDate(moveInDate)
		value, valErr := parseCents(premium)
		if nameErr != nil {
			rowErr = append(rowErr, RowError{Line: line, Reason: nameErr.Error()})
			continue
		}
		if emailErr != nil {
			rowErr = append(rowErr, RowError{Line: line, Reason: emailErr.Error()})
			continue
		}
		if unitErr != nil {
			rowErr = append(rowErr, RowError{Line: line, Reason: unitErr.Error()})
			continue
		}
		if dateErr != nil {
			rowErr = append(rowErr, RowError{Line: line, Reason: dateErr.Error()})
			continue
		}
		if valErr != nil {
			rowErr = append(rowErr, RowError{Line: line, Reason: valErr.Error()})
			continue
		}
		tenants = append(tenants, Tenant{Name: name, Email: email, Unit: unit, MoveInDate: date, PremiumCents: value})
	}
	return tenants, rowErr, nil
}

func parseUnit(unit string) (string, error) {
	unitPattern := regexp.MustCompile(`^[0-9]{1}[A-Z]{1}$`)
	if !unitPattern.MatchString(unit) {
		return unit, fmt.Errorf("invalid unit input %s", unit)
	}
	return unit, nil
}

func parseName(name string) (string, error) {
	if len(name) == 0 {
		return name, fmt.Errorf("empty name")
	}
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9]* ([a-zA-Z0-9]*)*$`)
	if !namePattern.MatchString(name) {
		return name, fmt.Errorf("invalid name %s", name)
	}
	return name, nil
}

func parseDate(date string) (time.Time, error) {
	dt, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format %s", date)
	}
	return dt, nil
}

func parseEmail(email string) (string, error) {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(email) {
		return email, fmt.Errorf("invalid email %s", email)
	}
	return email, nil
}

func parseCents(value string) (int64, error) {
	dol, cent, split := strings.Cut(value, ".")
	if len(dol) == 0 {
		return 0, fmt.Errorf("empty dollars %s", value)
	}
	if dol[0] == '-' || dol[0] == '+' {
		return 0, fmt.Errorf("invalid dol %s", value)
	}
	var dolInt, centInt int64
	var dolErr, centErr error
	if split {
		if cent[0] == '-' || cent[0] == '+' {
			return 0, fmt.Errorf("invalid cent %s", value)
		}
		if len(cent) > 2 { // greater than 2 cents
			return 0, fmt.Errorf("value %s has greater than 2 decimal points", value)
		}
		if len(cent) == 1 { // edge case 12.5
			cent = cent + "0"
		}
		centInt, centErr = strconv.ParseInt(cent, 10, 64)
		if centErr != nil {
			return 0, fmt.Errorf("invalid dollar amount %s", value)
		}
	}
	dolInt, dolErr = strconv.ParseInt(dol, 10, 64)
	if dolErr != nil {
		return 0, fmt.Errorf("invalid dollar amount %s", value)
	}

	return dolInt*100 + centInt, nil
}
