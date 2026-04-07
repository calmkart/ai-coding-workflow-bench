package leaderelection

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Lease represents a leader lease.
type Lease struct {
	HolderID      string
	AcquireTime   time.Time
	RenewTime     time.Time
	LeaseDuration time.Duration
	Version       int64
}

// IsExpired returns true if the lease has expired.
func (l *Lease) IsExpired() bool {
	return time.Now().After(l.RenewTime.Add(l.LeaseDuration))
}

// LeaseStore persists lease information.
type LeaseStore interface {
	Get(name string) (*Lease, error)
	Create(name string, lease *Lease) error
	Update(name string, lease *Lease) error // must check version for optimistic lock
}

// Errors
var (
	ErrLeaseNotFound    = errors.New("lease not found")
	ErrVersionConflict  = errors.New("version conflict")
)

// InMemoryLeaseStore is an in-memory lease store.
// TODO: Implement with optimistic locking.
type InMemoryLeaseStore struct {
	mu     sync.Mutex
	leases map[string]*Lease
}

func NewInMemoryLeaseStore() *InMemoryLeaseStore {
	return &InMemoryLeaseStore{
		leases: make(map[string]*Lease),
	}
}

func (s *InMemoryLeaseStore) Get(name string) (*Lease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.leases[name]
	if !ok {
		return nil, ErrLeaseNotFound
	}
	// Return a copy
	copy := *l
	return &copy, nil
}

// TODO: Implement Create and Update with version checking.
func (s *InMemoryLeaseStore) Create(name string, lease *Lease) error {
	return errors.New("not implemented")
}

func (s *InMemoryLeaseStore) Update(name string, lease *Lease) error {
	return errors.New("not implemented")
}

// LeaderOpts configures the leader elector.
type LeaderOpts struct {
	LeaseDuration    time.Duration
	RenewDeadline    time.Duration
	RetryPeriod      time.Duration
	OnStartedLeading func()
	OnStoppedLeading func()
}

// LeaderElector implements leader election using leases.
// TODO: Implement Run, tryAcquireOrRenew, IsLeader.
type LeaderElector struct {
	id       string
	leaseName string
	store    LeaseStore
	opts     LeaderOpts
	isLeader bool
	mu       sync.RWMutex
}

// NewLeaderElector creates a new leader elector.
func NewLeaderElector(id string, store LeaseStore, opts LeaderOpts) *LeaderElector {
	return &LeaderElector{
		id:        id,
		leaseName: "leader-lease",
		store:     store,
		opts:      opts,
	}
}

// IsLeader returns whether this elector currently holds the leader lease.
func (le *LeaderElector) IsLeader() bool {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.isLeader
}

// Run starts the leader election loop.
// TODO: Implement — try to acquire, renew if leader, retry if not.
func (le *LeaderElector) Run(ctx context.Context) {
	// stub
}
