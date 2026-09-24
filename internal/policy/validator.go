package policy

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidPolicy = errors.New("invalid routing policy")
	ErrPolicyExpired = errors.New("routing policy is expired")
)

type Validator struct {
	backendIDs map[string]struct{}
	maxTTL     time.Duration
}

func NewValidator(backendIDs []string, maxTTL time.Duration) (*Validator, error) {
	if len(backendIDs) == 0 || maxTTL <= 0 {
		return nil, fmt.Errorf("%w: backend IDs and positive max TTL are required", ErrInvalidPolicy)
	}
	ids := make(map[string]struct{}, len(backendIDs))
	for _, id := range backendIDs {
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("%w: backend ID cannot be empty", ErrInvalidPolicy)
		}
		if _, exists := ids[id]; exists {
			return nil, fmt.Errorf("%w: duplicate backend ID %q", ErrInvalidPolicy, id)
		}
		ids[id] = struct{}{}
	}
	return &Validator{backendIDs: ids, maxTTL: maxTTL}, nil
}

func (v *Validator) Validate(candidate Policy, now time.Time) (ValidatedPolicy, error) {
	if strings.TrimSpace(candidate.Version) == "" {
		return ValidatedPolicy{}, fmt.Errorf("%w: version is required", ErrInvalidPolicy)
	}
	if !candidate.ExpiresAt.After(now) {
		return ValidatedPolicy{}, ErrPolicyExpired
	}
	if candidate.ExpiresAt.After(now.Add(v.maxTTL)) {
		return ValidatedPolicy{}, fmt.Errorf("%w: expiry exceeds maximum TTL", ErrInvalidPolicy)
	}
	if len(candidate.Weights) != len(v.backendIDs) {
		return ValidatedPolicy{}, fmt.Errorf("%w: weights must include every configured backend exactly once", ErrInvalidPolicy)
	}
	var total uint32
	for id, weight := range candidate.Weights {
		if _, known := v.backendIDs[id]; !known {
			return ValidatedPolicy{}, fmt.Errorf("%w: unknown backend %q", ErrInvalidPolicy, id)
		}
		total += uint32(weight)
	}
	if total != WeightScale {
		return ValidatedPolicy{}, fmt.Errorf("%w: weights must total %d, got %d", ErrInvalidPolicy, WeightScale, total)
	}
	weights := make(map[string]uint16, len(candidate.Weights))
	for id, weight := range candidate.Weights {
		weights[id] = weight
	}
	return ValidatedPolicy{policy: Policy{Version: candidate.Version, ExpiresAt: candidate.ExpiresAt, Weights: weights}}, nil
}

type Store struct {
	mu      sync.RWMutex
	current ValidatedPolicy
	set     bool
}

func (s *Store) Publish(policy ValidatedPolicy) {
	s.mu.Lock()
	s.current, s.set = policy, true
	s.mu.Unlock()
}

func (s *Store) Current(now time.Time) (ValidatedPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.set || !s.current.ExpiresAt().After(now) {
		return ValidatedPolicy{}, false
	}
	return s.current, true
}
