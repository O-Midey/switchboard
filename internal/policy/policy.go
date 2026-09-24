// Package policy defines the future control-plane seam. It is deliberately not
// wired into request routing in milestone one.
package policy

import (
	"context"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
)

const WeightScale uint32 = 10_000

type Input struct {
	GeneratedAt time.Time
	Backends    []backend.TargetSnapshot
}

type Policy struct {
	Version   string
	ExpiresAt time.Time
	Weights   map[string]uint16
}

// ValidatedPolicy is the only policy representation that a future router may
// consume. Its fields are private so callers cannot bypass validation.
type ValidatedPolicy struct{ policy Policy }

func (p ValidatedPolicy) Version() string            { return p.policy.Version }
func (p ValidatedPolicy) ExpiresAt() time.Time       { return p.policy.ExpiresAt }
func (p ValidatedPolicy) Weights() map[string]uint16 { return cloneWeights(p.policy.Weights) }

func cloneWeights(weights map[string]uint16) map[string]uint16 {
	copyOfWeights := make(map[string]uint16, len(weights))
	for id, weight := range weights {
		copyOfWeights[id] = weight
	}
	return copyOfWeights
}

// Provider may eventually be backed by Jev. Its output is untrusted until a
// separate validator checks membership, bounds, freshness, and fallback rules.
type Provider interface {
	Propose(context.Context, Input) (Policy, error)
}
