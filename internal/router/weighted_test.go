package router

import (
	"net/http"
	"testing"

	"github.com/O-Midey/switchboard/internal/backend"
)

func TestWeightedSelectionIsStableForRequestKey(t *testing.T) {
	a, b := target("a", true), target("b", true)
	r, err := NewWeighted(map[string]uint16{"a": 7_000, "b": 3_000})
	if err != nil {
		t.Fatal(err)
	}
	request := Request{Method: http.MethodGet, Path: "/search", Header: http.Header{"X-Request-Id": []string{"request-42"}}}
	first, err := r.Select(request, []*backend.Target{a, b})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		selected, selectErr := r.Select(request, []*backend.Target{a, b})
		if selectErr != nil || selected.ID != first.ID {
			t.Fatalf("selection %d = %v, %v; want %s", i, selected, selectErr, first.ID)
		}
	}
}

func TestWeightedSelectionFallsBackAcrossHealthyTargets(t *testing.T) {
	a, b := target("a", false), target("b", true)
	r, err := NewWeighted(map[string]uint16{"a": 10_000, "b": 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Select(Request{Path: "/health"}, []*backend.Target{a, b}); err == nil {
		t.Fatal("expected no positive weight among healthy targets")
	}
}
