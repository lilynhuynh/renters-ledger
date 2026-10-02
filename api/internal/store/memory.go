package store

import (
	"context"
	"sync"

	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
)

// Compile-time check that *MemoryStore satisfies Store, the closest thing to `implements Store`.
// If a method is missing or has the wrong signature, the build fails right here.
var _ Store = (*MemoryStore)(nil)

// MemoryStore keeps policies in a map. It is used for Day 2 and in tests.
// Like a @Repository backed by a ConcurrentHashMap.
type MemoryStore struct {
	mu       sync.RWMutex // guards policies; HTTP handlers run concurrently (one goroutine per request)
	policies map[string]policy.Policy
	nextID   int
}

// NewMemoryStore is a constructor. Go has no `new ClassName()` with logic, so
// NewXxx functions are the convention.
//
// TODO(Day2): initialise the map. A nil map panics on write.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// TODO(Day2): read-lock, copy values into a slice, sort newest first (sort.Slice or slices.SortFunc).
func (m *MemoryStore) List(ctx context.Context) ([]policy.Policy, error) {
	panic("TODO(Day2): MemoryStore.List")
}

// TODO(Day2): read-lock and look up. Return ErrNotFound if missing.
// Hint: the two-value lookup `p, ok := m.policies[id]`.
func (m *MemoryStore) Get(ctx context.Context, id string) (policy.Policy, error) {
	panic("TODO(Day2): MemoryStore.Get")
}

// TODO(Day2): write-lock, assign an ID (e.g. fmt.Sprintf("pol_%d", m.nextID)), Version=1,
// set timestamps with time.Now().UTC(), store, and return the copy.
func (m *MemoryStore) Create(ctx context.Context, p policy.Policy) (policy.Policy, error) {
	panic("TODO(Day2): MemoryStore.Create")
}

// TODO(Day2): write-lock, Get, compare Version with expectedVersion, call
// policy.Transition(p.Status, policy.EventCancel), bump Version, save.
// (Transition itself is implemented on Day 4. Until then, set the status directly.)
func (m *MemoryStore) Cancel(ctx context.Context, id string, expectedVersion int64) (policy.Policy, error) {
	panic("TODO(Day2): MemoryStore.Cancel")
}
