package router

import (
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
)

func TestRoundRobinSkipsUnhealthyAndRotates(t *testing.T) {
	a, b, c := target("a", true), target("b", false), target("c", true)
	r := &RoundRobin{}
	for i, want := range []string{"a", "c", "a"} {
		got, err := r.Select(Request{}, []*backend.Target{a, b, c})
		if err != nil || got.ID != want {
			t.Fatalf("selection %d = %v, %v; want %s", i, got, err, want)
		}
	}
}

func TestLeastConnectionsSelectsLowest(t *testing.T) {
	a, b := target("a", true), target("b", true)
	a.Metrics.Start()
	got, err := (LeastConnections{}).Select(Request{}, []*backend.Target{a, b})
	if err != nil || got.ID != "b" {
		t.Fatalf("got %v, %v; want b", got, err)
	}
}

func target(id string, healthy bool) *backend.Target {
	u, _ := url.Parse("http://" + id)
	t := backend.NewTarget(id, u, httputil.NewSingleHostReverseProxy(u), backend.NewMetrics(time.Minute))
	if healthy {
		t.RecordHealth(true, time.Now(), "", 1, 1)
	}
	return t
}
