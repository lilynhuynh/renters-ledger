package store

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

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
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{policies: make(map[string]policy.Policy)}
	// Dont need to init mu because it is set to unlocked, and int = 0
}

func (m *MemoryStore) List(ctx context.Context) ([]policy.Policy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	policies := make([]policy.Policy, 0, len(m.policies)) // return empty list
	for _, p := range m.policies {
		policies = append(policies, p)
	}
	slices.SortFunc(policies, func(a, b policy.Policy) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	}) // descending order - newest first
	return policies, nil
}

func (m *MemoryStore) Get(ctx context.Context, id string) (policy.Policy, error) {
	m.mu.RLock() // Lock
	defer m.mu.RUnlock()

	p, ok := m.policies[id] // where p = value, ok = true (key exists)
	if !ok {
		return policy.Policy{}, ErrNotFound
	}
	return p, nil
}

func (m *MemoryStore) Create(ctx context.Context, p policy.Policy) (policy.Policy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++ // increment id
	id := fmt.Sprintf("pol_%d", m.nextID)
	timestamp := time.Now().UTC()

	p.CreatedAt = timestamp
	p.UpdatedAt = timestamp
	p.ID = id
	p.Version = 1

	m.policies[p.ID] = p

	return p, nil
}

func (m *MemoryStore) Cancel(ctx context.Context, id string, expectedVersion int64) (policy.Policy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.policies[id]
	if !ok {
		return policy.Policy{}, ErrNotFound
	}

	if p.Version != expectedVersion {
		return policy.Policy{}, ErrVersionConflict
	}
	policy.Transition(p.Status, policy.EventCancel)
	// Update the copy if found different version
	p.Version++
	p.Status = policy.StatusCancelled
	p.UpdatedAt = time.Now().UTC()
	m.policies[p.ID] = p
	return p, nil
}
