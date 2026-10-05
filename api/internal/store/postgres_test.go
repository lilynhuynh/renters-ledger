package store

// Integration tests for PostgresStore. They run against a real database and skip when
// DATABASE_URL is unset, like a JUnit test guarded by Assumptions.assumeTrue(...).
//
//	DATABASE_URL="postgres://ledger:ledger@localhost:5432/renters_ledger?sslmode=disable" go test ./internal/store/ -v

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
)

// ---------- helpers ----------

// createStore connects to DATABASE_URL, or skips the test when it is unset.
func createStore(t *testing.T) (*PostgresStore, *pgxpool.Pool) {
	t.Helper()
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close) // registered first, so it runs last (cleanups run in reverse)

	if err := pool.Ping(ctx); err != nil { // Check pool status
		t.Fatalf("ping database: %v", err)
	}
	return NewPostgresStore(pool), pool
}

// createCustomerTest inserts a customer (policies need one for the FK) and deletes it after the test.
func createCustomerTest(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()) // generate unique email
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING id`,
		"Test Customer", email).Scan(&id)
	if err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	t.Cleanup(func() {
		// Clean up tables (policies first bc of FK)
		pool.Exec(ctx, `DELETE FROM policies WHERE customer_id = $1`, id)
		pool.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	})
	return id
}

// newTestPolicy builds a policy for the given customer with fixed, easy-to-check values.
func newTestPolicy(customerID string) policy.Policy {
	return policy.Policy{
		CustomerID:    customerID,
		Unit:          "4B",
		Status:        policy.StatusQuoted,
		PremiumCents:  1500,
		EffectiveDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ---------- tests ----------

func TestPostgresStore_CreateAndGet(t *testing.T) {
	st, pool := createStore(t)
	ctx := context.Background()
	in := newTestPolicy(createCustomerTest(t, pool))

	created, err := st.Create(ctx, in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Fields the database fills in (DEFAULT gen_random_uuid(), DEFAULT 1, DEFAULT now()).
	if created.ID == "" {
		t.Errorf("created.ID is empty, want a generated uuid")
	}
	if created.Version != 1 {
		t.Errorf("created.Version = %d, want 1", created.Version)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Errorf("timestamps not set: CreatedAt=%v UpdatedAt=%v", created.CreatedAt, created.UpdatedAt)
	}

	got, err := st.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get(%s): %v", created.ID, err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %s, want %s", got.ID, created.ID)
	}
	if got.CustomerID != in.CustomerID {
		t.Errorf("CustomerID = %s, want %s", got.CustomerID, in.CustomerID)
	}
	if got.Unit != in.Unit {
		t.Errorf("Unit = %s, want %s", got.Unit, in.Unit)
	}
	if got.Status != in.Status {
		t.Errorf("Status = %s, want %s", got.Status, in.Status)
	}
	if got.PremiumCents != in.PremiumCents {
		t.Errorf("PremiumCents = %d, want %d", got.PremiumCents, in.PremiumCents)
	}
	if !got.EffectiveDate.Equal(in.EffectiveDate) { // time.Time: use Equal, not ==
		t.Errorf("EffectiveDate = %v, want %v", got.EffectiveDate, in.EffectiveDate)
	}
	if got.Version != 1 {
		t.Errorf("Version = %d, want 1", got.Version)
	}
}

func TestPostgresStore_GetErrors(t *testing.T) {
	st, _ := createStore(t)
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{"unknown id", "00000000-0000-0000-0000-000000000000", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.Get(context.Background(), tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get(%s) err = %v, want %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestPostgresStore_List(t *testing.T) {
	st, pool := createStore(t)
	ctx := context.Background()

	created, err := st.Create(ctx, newTestPolicy(createCustomerTest(t, pool)))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := st.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Don't check len(): rows from manual curl testing may also be in the table.
	found := false
	for _, p := range list {
		if p.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List() does not contain created policy %s (got %d policies)", created.ID, len(list))
	}
}
