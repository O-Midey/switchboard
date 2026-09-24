package proxy

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
	"github.com/O-Midey/switchboard/internal/router"
)

type Handler struct {
	router  router.Router
	targets []*backend.Target
	log     *slog.Logger
}

func NewHandler(r router.Router, targets []*backend.Target, log *slog.Logger) *Handler {
	return &Handler{router: r, targets: targets, log: log}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	started := time.Now()
	target, err := h.router.Select(router.Request{Method: request.Method, Path: request.URL.Path, Header: request.Header.Clone()}, h.targets)
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "NO_HEALTHY_BACKENDS", "No healthy upstream is available.")
		h.log.Warn("request rejected", "method", request.Method, "path", request.URL.Path, "error", err)
		return
	}
	recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
	target.Metrics.Start()
	target.Proxy.ServeHTTP(recorder, request)
	latency := time.Since(started)
	failed := recorder.status >= http.StatusInternalServerError
	target.Metrics.Finish(latency, failed)
	h.log.Info("request completed", "method", request.Method, "path", request.URL.Path, "status", recorder.status, "backend", target.ID, "duration_ms", float64(latency.Microseconds())/1000, "strategy", h.router.Name())
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) { r.status = code; r.ResponseWriter.WriteHeader(code) }

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{"code": code, "message": message, "statusCode": status})
}
