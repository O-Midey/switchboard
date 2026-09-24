package router

import "github.com/O-Midey/switchboard/internal/backend"

type LeastConnections struct{}

func (LeastConnections) Name() string { return "least-connections" }

func (LeastConnections) Select(_ Request, targets []*backend.Target) (*backend.Target, error) {
	healthy := Healthy(targets)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyBackends
	}
	selected := healthy[0]
	minimum := selected.Metrics.Snapshot().ActiveRequests
	for _, candidate := range healthy[1:] {
		active := candidate.Metrics.Snapshot().ActiveRequests
		if active < minimum {
			selected, minimum = candidate, active
		}
	}
	return selected, nil
}

func New(strategy string) Router {
	if strategy == "least-connections" {
		return LeastConnections{}
	}
	return &RoundRobin{}
}
