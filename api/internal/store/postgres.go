package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
)

var _ Store = (*PostgresStore)(nil)

// PostgresStore is the real implementation (Day 3). Like a @Repository using JdbcTemplate.
// Plain SQL, no ORM: you write the queries and Scan the columns into struct fields.
type PostgresStore struct {
	pool *pgxpool.Pool // pointer to the pool to share same pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) List(ctx context.Context) ([]policy.Policy, error) {
	const listSQL = `
		SELECT
			id,
			customer_id,
			unit,
			status,
			premium_cents,
			effective_date,
			version,
			created_at,
			updated_at
		FROM policies
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, listSQL)
	if err != nil {
		return nil, fmt.Errorf("policy list error: %w", err)
	}
	defer rows.Close() // runs when function returns

	var pList = make([]policy.Policy, 0)
	for rows.Next() {
		var p policy.Policy // new policy struct
		if err := rows.Scan(&p.ID, &p.CustomerID, &p.Unit, &p.Status, &p.PremiumCents, &p.EffectiveDate, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan of row error: %w", err)
		}
		pList = append(pList, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	return pList, nil
}

func (s *PostgresStore) Get(ctx context.Context, id string) (policy.Policy, error) {
	const getSQL = `
		SELECT
			id,
			customer_id,
			unit,
			status,
			premium_cents,
			effective_date,
			version,
			created_at,
			updated_at
		FROM policies
		WHERE id = $1
	`

	var p policy.Policy
	err := s.pool.QueryRow(ctx, getSQL, id).
		Scan(&p.ID, &p.CustomerID, &p.Unit, &p.Status, &p.PremiumCents, &p.EffectiveDate, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return policy.Policy{}, ErrNotFound
	}
	if err != nil {
		return policy.Policy{}, fmt.Errorf("let policy id %s: %w", id, err)
	}
	return p, nil
}

func (s *PostgresStore) Create(ctx context.Context, p policy.Policy) (policy.Policy, error) {
	const createSQL = `
		INSERT INTO policies (customer_id, unit, status, premium_cents, effective_date)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, version, created_at, updated_at
	`
	err := s.pool.QueryRow(ctx, createSQL, p.CustomerID, p.Unit, p.Status, p.PremiumCents, p.EffectiveDate).
		Scan(&p.ID, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("create policy id %s: %w", p.ID, err)
	}
	return p, nil
}

// TODO(Day4): in one transaction (pool.Begin, defer tx.Rollback, tx.Commit):
// UPDATE policies SET status=$1, version=version+1 WHERE id=$2 AND version=$3.
// 0 rows affected means ErrVersionConflict (or ErrNotFound). Then INSERT INTO policy_events.
func (s *PostgresStore) Cancel(ctx context.Context, id string, expectedVersion int64) (policy.Policy, error) {
	panic("TODO(Day4): PostgresStore.Cancel")
}
