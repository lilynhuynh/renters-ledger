package store

import (
	"context"

	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
)

var _ Store = (*PostgresStore)(nil)

// PostgresStore is the real implementation (Day 3). Like a @Repository using JdbcTemplate.
// Plain SQL, no ORM: you write the queries and Scan the columns into struct fields.
//
// TODO(Day3): `go get github.com/jackc/pgx/v5` and add a field
//
//	pool *pgxpool.Pool // connection pool, like a HikariCP DataSource
type PostgresStore struct{}

// TODO(Day3): accept a *pgxpool.Pool (created in cmd/api/main.go from DATABASE_URL).
func NewPostgresStore() *PostgresStore {
	return &PostgresStore{}
}

// TODO(Day3): SELECT ... FROM policies ORDER BY created_at DESC; rows.Next()/rows.Scan loop.
func (s *PostgresStore) List(ctx context.Context) ([]policy.Policy, error) {
	panic("TODO(Day3): PostgresStore.List")
}

// TODO(Day3): QueryRow + Scan. Map pgx.ErrNoRows to ErrNotFound with errors.Is.
func (s *PostgresStore) Get(ctx context.Context, id string) (policy.Policy, error) {
	panic("TODO(Day3): PostgresStore.Get")
}

// TODO(Day3): INSERT ... RETURNING id, version, created_at, updated_at.
func (s *PostgresStore) Create(ctx context.Context, p policy.Policy) (policy.Policy, error) {
	panic("TODO(Day3): PostgresStore.Create")
}

// TODO(Day4): in one transaction (pool.Begin, defer tx.Rollback, tx.Commit):
// UPDATE policies SET status=$1, version=version+1 WHERE id=$2 AND version=$3.
// 0 rows affected means ErrVersionConflict (or ErrNotFound). Then INSERT INTO policy_events.
func (s *PostgresStore) Cancel(ctx context.Context, id string, expectedVersion int64) (policy.Policy, error) {
	panic("TODO(Day4): PostgresStore.Cancel")
}
