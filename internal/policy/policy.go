// Package policy defines the future control-plane seam. It is deliberately not
// wired into request routing in milestone one.
package policy

import (
	"context"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
)

type Input struct {
	GeneratedAt time.Time
	Backends    []backend.TargetSnapshot
}

type Policy struct {
	Version   string
	ExpiresAt time.Time
	Weights   map[string]uint16
}

// Provider may eventually be backed by Jev. Its output is untrusted until a
// separate validator checks membership, bounds, freshness, and fallback rules.
type Provider interface {
	Propose(context.Context, Input) (Policy, error)
}
