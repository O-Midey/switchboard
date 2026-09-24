package backend

import (
	"sync"
	"sync/atomic"
	"time"
)

type bucket struct {
	second   int64
	requests uint64
	errors   uint64
	latency  time.Duration
}

type Metrics struct {
	active  atomic.Int64
	mu      sync.Mutex
	buckets []bucket
	now     func() time.Time
}

type Snapshot struct {
	ActiveRequests int64   `json:"activeRequests"`
	Requests       uint64  `json:"requests"`
	Errors         uint64  `json:"errors"`
	ErrorRate      float64 `json:"errorRate"`
	AvgLatencyMs   float64 `json:"avgLatencyMs"`
}

func NewMetrics(window time.Duration) *Metrics {
	seconds := max(1, int(window/time.Second))
	return &Metrics{buckets: make([]bucket, seconds), now: time.Now}
}

func (m *Metrics) Start() { m.active.Add(1) }

func (m *Metrics) Finish(latency time.Duration, failed bool) {
	m.active.Add(-1)
	second := m.now().Unix()
	index := int(second % int64(len(m.buckets)))
	m.mu.Lock()
	b := &m.buckets[index]
	if b.second != second {
		*b = bucket{second: second}
	}
	b.requests++
	b.latency += latency
	if failed {
		b.errors++
	}
	m.mu.Unlock()
}

func (m *Metrics) Snapshot() Snapshot {
	cutoff := m.now().Unix() - int64(len(m.buckets)) + 1
	result := Snapshot{ActiveRequests: m.active.Load()}
	var latency time.Duration
	m.mu.Lock()
	for _, b := range m.buckets {
		if b.second < cutoff {
			continue
		}
		result.Requests += b.requests
		result.Errors += b.errors
		latency += b.latency
	}
	m.mu.Unlock()
	if result.Requests > 0 {
		result.ErrorRate = float64(result.Errors) / float64(result.Requests)
		result.AvgLatencyMs = float64(latency.Microseconds()) / 1000 / float64(result.Requests)
	}
	return result
}
