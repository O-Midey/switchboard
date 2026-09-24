package backend

import (
	"testing"
	"time"
)

func TestMetricsRollingWindow(t *testing.T) {
	current := time.Unix(100, 0)
	metrics := NewMetrics(2 * time.Second)
	metrics.now = func() time.Time { return current }
	metrics.Start()
	metrics.Finish(20*time.Millisecond, false)
	current = current.Add(time.Second)
	metrics.Start()
	metrics.Finish(40*time.Millisecond, true)
	snapshot := metrics.Snapshot()
	if snapshot.Requests != 2 || snapshot.Errors != 1 || snapshot.AvgLatencyMs != 30 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	current = current.Add(2 * time.Second)
	if got := metrics.Snapshot().Requests; got != 0 {
		t.Fatalf("expired requests = %d, want 0", got)
	}
}
