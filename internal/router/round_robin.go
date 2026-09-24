package router

import (
	"sync/atomic"

	"github.com/O-Midey/switchboard/internal/backend"
)

type RoundRobin struct{ next atomic.Uint64 }

func (r *RoundRobin) Name() string { return "round-robin" }

func (r *RoundRobin) Select(_ Request, targets []*backend.Target) (*backend.Target, error) {
	healthy := Healthy(targets)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyBackends
	}
	index := (r.next.Add(1) - 1) % uint64(len(healthy))
	return healthy[index], nil
}
