package proxy

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
)

func NewTarget(id, rawURL string, window time.Duration, log *slog.Logger) (*backend.Target, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(parsed)
	reverseProxy.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}).DialContext, ForceAttemptHTTP2: true, MaxIdleConns: 256, MaxIdleConnsPerHost: 64, IdleConnTimeout: 90 * time.Second, TLSHandshakeTimeout: 3 * time.Second, ResponseHeaderTimeout: 30 * time.Second}
	reverseProxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, proxyErr error) {
		log.Error("upstream request failed", "backend", id, "error", proxyErr)
		writeError(writer, http.StatusBadGateway, "UPSTREAM_ERROR", "The selected upstream could not complete the request.")
	}
	reverseProxy.ModifyResponse = func(response *http.Response) error {
		response.Header.Set("X-Switchboard-Backend", id)
		return nil
	}
	return backend.NewTarget(id, parsed, reverseProxy, backend.NewMetrics(window)), nil
}
