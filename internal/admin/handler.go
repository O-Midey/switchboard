package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
)

type Handler struct {
	targets  []*backend.Target
	strategy string
	started  time.Time
}

func NewHandler(targets []*backend.Target, strategy string) http.Handler {
	h := &Handler{targets: targets, strategy: strategy, started: time.Now()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /-/healthz", h.live)
	mux.HandleFunc("GET /-/readyz", h.ready)
	mux.HandleFunc("GET /api/v1/state", h.state)
	mux.HandleFunc("GET /metrics", h.metrics)
	return mux
}

func (h *Handler) live(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ready(writer http.ResponseWriter, _ *http.Request) {
	for _, target := range h.targets {
		if target.Healthy() {
			writeJSON(writer, http.StatusOK, map[string]string{"status": "ready"})
			return
		}
	}
	writeJSON(writer, http.StatusServiceUnavailable, map[string]any{"code": "NO_HEALTHY_BACKENDS", "message": "No healthy upstream is available.", "statusCode": 503})
}

func (h *Handler) state(writer http.ResponseWriter, _ *http.Request) {
	targets := make([]backend.TargetSnapshot, 0, len(h.targets))
	for _, target := range h.targets {
		targets = append(targets, target.Snapshot())
	}
	writeJSON(writer, http.StatusOK, map[string]any{"generatedAt": time.Now().UTC(), "uptimeSeconds": int64(time.Since(h.started).Seconds()), "strategy": h.strategy, "policySource": "deterministic-baseline", "backends": targets})
}

func (h *Handler) metrics(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var output strings.Builder
	output.WriteString("# HELP switchboard_backend_healthy Whether the backend is healthy.\n# TYPE switchboard_backend_healthy gauge\n")
	output.WriteString("# HELP switchboard_backend_active_requests Current in-flight requests.\n# TYPE switchboard_backend_active_requests gauge\n")
	output.WriteString("# HELP switchboard_backend_window_requests Requests observed in the rolling window.\n# TYPE switchboard_backend_window_requests gauge\n")
	output.WriteString("# HELP switchboard_backend_window_error_rate Error ratio in the rolling window.\n# TYPE switchboard_backend_window_error_rate gauge\n")
	output.WriteString("# HELP switchboard_backend_window_latency_ms Average latency in the rolling window.\n# TYPE switchboard_backend_window_latency_ms gauge\n")
	for _, target := range h.targets {
		s := target.Snapshot()
		healthy := 0
		if s.Healthy {
			healthy = 1
		}
		label := fmt.Sprintf("backend=%q", s.ID)
		fmt.Fprintf(&output, "switchboard_backend_healthy{%s} %d\n", label, healthy)
		fmt.Fprintf(&output, "switchboard_backend_active_requests{%s} %d\n", label, s.Metrics.ActiveRequests)
		fmt.Fprintf(&output, "switchboard_backend_window_requests{%s} %d\n", label, s.Metrics.Requests)
		fmt.Fprintf(&output, "switchboard_backend_window_error_rate{%s} %f\n", label, s.Metrics.ErrorRate)
		fmt.Fprintf(&output, "switchboard_backend_window_latency_ms{%s} %f\n", label, s.Metrics.AvgLatencyMs)
	}
	_, _ = writer.Write([]byte(output.String()))
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
