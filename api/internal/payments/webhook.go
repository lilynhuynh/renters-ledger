// Package payments handles inbound payment-processor webhooks: verify they are genuine,
// drop duplicates, and turn them into policy events.
// It must NOT import httpapi (httpapi calls payments, never the reverse) and must NOT
// change policy status directly. It goes through policy.Transition and a Store.
package payments

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBadSignature = errors.New("webhook signature invalid") // -> 401
	ErrDuplicate    = errors.New("webhook event already processed")
)

// Event is the decoded webhook body we care about.
type Event struct {
	ID          string    `json:"id"`   // processor's unique event id, the dedupe key
	Type        string    `json:"type"` // e.g. "payment.succeeded", "payment.failed"
	PolicyID    string    `json:"policy_id"`
	AmountCents int64     `json:"amount_cents"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// The backtick `json:"..."` parts are struct tags, Go's version of @JsonProperty.

// VerifySignature checks that header == hex(HMAC-SHA256(secret, payload)).
//
// TODO(Day4): crypto/hmac + crypto/sha256 + encoding/hex. Compare with hmac.Equal,
// NOT ==, because hmac.Equal takes constant time and does not leak timing information.
// Return ErrBadSignature on mismatch or a malformed header.
// The stub fails closed (rejects everything). Security checks should never default to "allow".
func VerifySignature(payload []byte, header string, secret []byte) error {
	return ErrBadSignature
}

// Deduper records event IDs that have been processed. The Postgres version inserts into
// webhook_events (event_id UNIQUE). A unique-violation error means "seen before".
// Idempotency comes from the database constraint, not from a read-then-write check.
type Deduper interface {
	// MarkSeen returns ErrDuplicate if eventID was already recorded.
	MarkSeen(ctx context.Context, eventID string) error
}

// Process verifies, dedupes and applies one webhook.
//
// TODO(Day4): VerifySignature -> json.Unmarshal into Event -> d.MarkSeen (ErrDuplicate means
// return nil, because a duplicate is a success for the sender) -> map Type to a policy.Event ->
// apply through the store. Ideally MarkSeen and the status update share one DB transaction.
func Process(ctx context.Context, d Deduper, payload []byte, sigHeader string, secret []byte) error {
	return nil
}
