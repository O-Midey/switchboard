package router

import (
	"errors"
	"net/http"

	"github.com/O-Midey/switchboard/internal/backend"
)

var ErrNoHealthyBackends = errors.New("no healthy backends available")

type Request struct {
	Method string
	Path   string
	Header http.Header
}

// Router is the data-plane seam. Implementations must be concurrency-safe and
// choose only from the supplied healthy targets.
type Router interface {
	Select(Request, []*backend.Target) (*backend.Target, error)
	Name() string
}

func Healthy(targets []*backend.Target) []*backend.Target {
	healthy := make([]*backend.Target, 0, len(targets))
	for _, target := range targets {
		if target.Healthy() {
			healthy = append(healthy, target)
		}
	}
	return healthy
}
