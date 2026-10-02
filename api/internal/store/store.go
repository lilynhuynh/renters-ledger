// Package store persists policies. It hides whether data lives in memory or in Postgres.
// It may import policy (the domain). It must NOT import httpapi or know about HTTP.
package store

import (
	"context"
	"errors"

	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
)

// Sentinel errors that callers check with errors.Is. httpapi maps them to status codes.
var (
	ErrNotFound        = errors.New("policy not found")                    // -> 404
	ErrVersionConflict = errors.New("policy was modified by someone else") // -> 409, like OptimisticLockException
)

// Store is the persistence contract, like a Spring Data repository interface.
// Both MemoryStore and PostgresStore satisfy it implicitly by having these methods.
//
// Every method takes a context.Context first. It carries the request's deadline and
// cancellation. There is no direct Java equivalent; it does what a request-scoped
// timeout plus thread interruption would do, passed explicitly.
type Store interface {
	// List returns all policies, newest first.
	List(ctx context.Context) ([]policy.Policy, error)

	// Get returns one policy or ErrNotFound.
	Get(ctx context.Context, id string) (policy.Policy, error)

	// Create saves a new policy. It assigns ID, Version=1, CreatedAt and UpdatedAt,
	// then returns the saved copy.
	Create(ctx context.Context, p policy.Policy) (policy.Policy, error)

	// Cancel moves a policy to cancelled using policy.Transition. It succeeds only if
	// the stored Version equals expectedVersion (otherwise ErrVersionConflict). It bumps
	// Version and, in Postgres, appends a policy_events row in the same transaction.
	Cancel(ctx context.Context, id string, expectedVersion int64) (policy.Policy, error)
}
