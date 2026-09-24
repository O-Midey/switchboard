package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"unknown":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown field to fail")
	}
}

func TestValidateRejectsDuplicateBackends(t *testing.T) {
	cfg := validConfig()
	cfg.Backends = append(cfg.Backends, cfg.Backends[0])
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate ID to fail")
	}
}

func validConfig() Config {
	return Config{ProxyAddress: ":8080", AdminAddress: ":9090", RoutingStrategy: "round-robin", ShutdownTimeout: Duration(5_000_000_000), MetricsWindow: Duration(30_000_000_000), Backends: []Backend{{ID: "a", URL: "http://localhost:8081", HealthPath: "/health", HealthInterval: Duration(2_000_000_000), HealthTimeout: Duration(1_000_000_000), UnhealthyAfter: 2, HealthyAfter: 1}}}
}
