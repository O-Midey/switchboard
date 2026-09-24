package health

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/O-Midey/switchboard/internal/backend"
	"github.com/O-Midey/switchboard/internal/config"
)

type Checker struct {
	client *http.Client
	log    *slog.Logger
}

func NewChecker(log *slog.Logger) *Checker { return &Checker{client: &http.Client{}, log: log} }

func (c *Checker) Start(ctx context.Context, target *backend.Target, cfg config.Backend) {
	c.check(ctx, target, cfg)
	ticker := time.NewTicker(cfg.HealthInterval.Value())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.check(ctx, target, cfg)
		}
	}
}

func (c *Checker) check(parent context.Context, target *backend.Target, cfg config.Backend) {
	ctx, cancel := context.WithTimeout(parent, cfg.HealthTimeout.Value())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL.JoinPath(cfg.HealthPath).String(), nil)
	if err != nil {
		target.RecordHealth(false, time.Now(), err.Error(), cfg.HealthyAfter, cfg.UnhealthyAfter)
		return
	}
	response, err := c.client.Do(request)
	if err != nil {
		target.RecordHealth(false, time.Now(), err.Error(), cfg.HealthyAfter, cfg.UnhealthyAfter)
		c.log.Debug("backend health check failed", "backend", target.ID, "error", err)
		return
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		detail := fmt.Sprintf("health check returned %d", response.StatusCode)
		target.RecordHealth(false, time.Now(), detail, cfg.HealthyAfter, cfg.UnhealthyAfter)
		c.log.Debug("backend health check failed", "backend", target.ID, "status", response.StatusCode)
		return
	}
	target.RecordHealth(true, time.Now(), "", cfg.HealthyAfter, cfg.UnhealthyAfter)
}
