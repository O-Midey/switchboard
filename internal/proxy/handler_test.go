package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
	"github.com/O-Midey/switchboard/internal/router"
)

func TestHandlerProxiesToHealthyBackendAndRecordsMetrics(t *testing.T) {
	target, err := NewTarget("a", "http://upstream.test", time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	target.Proxy.Transport = roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusCreated, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(request.URL.Path)), Request: request}, nil
	})
	target.RecordHealth(true, time.Now(), "", 1, 1)
	handler := NewHandler(&router.RoundRobin{}, []*backend.Target{target}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://switchboard/orders/1", nil))
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d", response.Code)
	}
	if response.Header().Get("X-Switchboard-Backend") != "a" {
		t.Fatalf("backend header = %q", response.Header().Get("X-Switchboard-Backend"))
	}
	if got := target.Metrics.Snapshot().Requests; got != 1 {
		t.Fatalf("requests = %d", got)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestHandlerFailsClosedWithoutHealthyBackend(t *testing.T) {
	target, err := NewTarget("a", "http://127.0.0.1:1", time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(&router.RoundRobin{}, []*backend.Target{target}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://switchboard/", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", response.Code)
	}
}
