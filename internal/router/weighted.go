package router

import (
	"errors"
	"hash/fnv"

	"github.com/O-Midey/switchboard/internal/backend"
)

var ErrNoPositiveWeights = errors.New("no healthy backend has a positive policy weight")

// Weighted is a deterministic candidate adapter for a validated policy. It
// is not wired into the live proxy until policy shadow results are trusted.
type Weighted struct{ weights map[string]uint16 }

func NewWeighted(weights map[string]uint16) (*Weighted, error) {
	copyOfWeights := make(map[string]uint16, len(weights))
	var total uint32
	for id, weight := range weights {
		copyOfWeights[id] = weight
		total += uint32(weight)
	}
	if total == 0 {
		return nil, ErrNoPositiveWeights
	}
	return &Weighted{weights: copyOfWeights}, nil
}

func (r *Weighted) Name() string { return "weighted-shadow" }

func (r *Weighted) Select(request Request, targets []*backend.Target) (*backend.Target, error) {
	healthy := Healthy(targets)
	var total uint32
	for _, target := range healthy {
		total += uint32(r.weights[target.ID])
	}
	if total == 0 {
		return nil, ErrNoPositiveWeights
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(requestKey(request)))
	point := uint32(hash.Sum32()) % total
	var cumulative uint32
	for _, target := range healthy {
		weight := uint32(r.weights[target.ID])
		if weight == 0 {
			continue
		}
		cumulative += weight
		if point < cumulative {
			return target, nil
		}
	}
	return nil, ErrNoPositiveWeights
}

func requestKey(request Request) string {
	if request.Header != nil {
		if key := request.Header.Get("X-Request-ID"); key != "" {
			return key
		}
	}
	return request.Method + " " + request.Path
}
