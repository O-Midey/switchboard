package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Target struct {
	ID      string
	URL     *url.URL
	Proxy   *httputil.ReverseProxy
	Metrics *Metrics

	mu            sync.RWMutex
	healthy       bool
	lastCheck     time.Time
	lastError     string
	successStreak int
	failureStreak int
}

type TargetSnapshot struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Healthy   bool      `json:"healthy"`
	LastCheck time.Time `json:"lastCheck"`
	LastError string    `json:"lastError,omitempty"`
	Metrics   Snapshot  `json:"metrics"`
}

func NewTarget(id string, targetURL *url.URL, proxy *httputil.ReverseProxy, metrics *Metrics) *Target {
	return &Target{ID: id, URL: targetURL, Proxy: proxy, Metrics: metrics}
}

func (t *Target) RecordHealth(success bool, checkedAt time.Time, detail string, healthyAfter, unhealthyAfter int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastCheck = checkedAt
	if success {
		t.successStreak++
		t.failureStreak = 0
		t.lastError = ""
		if t.successStreak >= healthyAfter {
			t.healthy = true
		}
		return
	}
	t.failureStreak++
	t.successStreak = 0
	t.lastError = detail
	if t.failureStreak >= unhealthyAfter {
		t.healthy = false
	}
}

func (t *Target) Healthy() bool { t.mu.RLock(); defer t.mu.RUnlock(); return t.healthy }

func (t *Target) Snapshot() TargetSnapshot {
	t.mu.RLock()
	snapshot := TargetSnapshot{ID: t.ID, URL: t.URL.String(), Healthy: t.healthy, LastCheck: t.lastCheck, LastError: t.lastError}
	t.mu.RUnlock()
	snapshot.Metrics = t.Metrics.Snapshot()
	return snapshot
}
