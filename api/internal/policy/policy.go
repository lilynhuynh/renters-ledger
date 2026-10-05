// Package policy is the domain core: the Policy type and its lifecycle rules.
// It is pure business logic. It must NOT import store, httpapi, payments or any database or HTTP package.
// Think of it as the domain/service layer in Spring with no framework annotations at all.
package policy

import "time"

// Status is a named string type, Go's lightweight stand-in for a Java enum.
// The compiler will not let you pass a plain string where a Status is expected
// without an explicit conversion, which gives you some type safety.
type Status string

// The full lifecycle. See docs/ARCHITECTURE.md, "Policy lifecycle".
const (
	StatusQuoted     Status = "quoted"     // price offered, nothing signed yet
	StatusPending    Status = "pending"    // submitted and waiting for the first payment
	StatusActive     Status = "active"     // in force and paid up
	StatusPastDue    Status = "past_due"   // a payment failed; grace period running
	StatusCancelled  Status = "cancelled"  // ended on purpose (terminal)
	StatusLapsed     Status = "lapsed"     // grace period ran out unpaid
	StatusReinstated Status = "reinstated" // brought back from lapsed; awaiting next payment
)

// Policy is one renters-insurance policy. Like a JPA @Entity, but with no annotations.
// Mapping to the database happens by hand in the store package.
type Policy struct {
	ID            string    `json:"id"` // UUID as text
	CustomerID    string    `json:"customer_id"`
	Unit          string    `json:"unit"`
	Status        Status    `json:"status"`
	PremiumCents  int64     `json:"premium_cents"`  // monthly premium in integer cents (see ingest.Tenant)
	EffectiveDate time.Time `json:"effective_date"` // coverage start date
	Version       int64     `json:"version"`        // optimistic-locking counter, like JPA @Version
	CreatedAt     time.Time `json:"created_at"`     // stored as timestamptz (always UTC in Go)
	UpdatedAt     time.Time `json:"updated_at"`
}
