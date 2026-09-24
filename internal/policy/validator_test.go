package policy

import (
	"errors"
	"testing"
	"time"
)

func TestValidatorRequiresNormalizedKnownWeights(t *testing.T) {
	now := time.Unix(100, 0)
	validator, err := NewValidator([]string{"atlas", "boreal"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := validator.Validate(Policy{Version: "v1", ExpiresAt: now.Add(30 * time.Second), Weights: map[string]uint16{"atlas": 7_000, "boreal": 3_000}}, now)
	if err != nil || valid.Version() != "v1" {
		t.Fatalf("valid policy = %+v, err = %v", valid, err)
	}
	_, err = validator.Validate(Policy{Version: "v2", ExpiresAt: now.Add(30 * time.Second), Weights: map[string]uint16{"atlas": 7_000, "elsewhere": 3_000}}, now)
	if !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("unknown backend error = %v", err)
	}
	_, err = validator.Validate(Policy{Version: "v3", ExpiresAt: now.Add(30 * time.Second), Weights: map[string]uint16{"atlas": 6_000, "boreal": 3_000}}, now)
	if !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("unnormalized error = %v", err)
	}
}

func TestValidatorRejectsExpiredAndLongLivedPolicies(t *testing.T) {
	now := time.Unix(100, 0)
	validator, err := NewValidator([]string{"atlas"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	base := Policy{Version: "v1", Weights: map[string]uint16{"atlas": uint16(WeightScale)}}
	_, err = validator.Validate(base, now)
	if !errors.Is(err, ErrPolicyExpired) {
		t.Fatalf("missing expiry error = %v", err)
	}
	base.ExpiresAt = now.Add(2 * time.Minute)
	_, err = validator.Validate(base, now)
	if !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("long TTL error = %v", err)
	}
}

func TestStoreDoesNotReturnExpiredPolicy(t *testing.T) {
	validator, err := NewValidator([]string{"atlas"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := validator.Validate(Policy{Version: "v1", ExpiresAt: time.Unix(110, 0), Weights: map[string]uint16{"atlas": uint16(WeightScale)}}, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	var store Store
	store.Publish(policy)
	if _, ok := store.Current(time.Unix(109, 0)); !ok {
		t.Fatal("expected current policy")
	}
	if _, ok := store.Current(time.Unix(110, 0)); ok {
		t.Fatal("expected expired policy to be unavailable")
	}
}
