package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const DefaultPath = "configs/switchboard.json"

type Duration time.Duration

func (d *Duration) UnmarshalJSON(value []byte) error {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	*d = Duration(parsed)
	return nil
}

func (d Duration) Value() time.Duration { return time.Duration(d) }

type Backend struct {
	ID             string   `json:"id"`
	URL            string   `json:"url"`
	HealthPath     string   `json:"healthPath"`
	HealthInterval Duration `json:"healthInterval"`
	HealthTimeout  Duration `json:"healthTimeout"`
	UnhealthyAfter int      `json:"unhealthyAfter"`
	HealthyAfter   int      `json:"healthyAfter"`
}

type Config struct {
	ProxyAddress    string    `json:"proxyAddress"`
	AdminAddress    string    `json:"adminAddress"`
	RoutingStrategy string    `json:"routingStrategy"`
	ShutdownTimeout Duration  `json:"shutdownTimeout"`
	MetricsWindow   Duration  `json:"metricsWindow"`
	Backends        []Backend `json:"backends"`
}

func Load(path string) (Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if value := os.Getenv("SWITCHBOARD_PROXY_ADDRESS"); value != "" {
		cfg.ProxyAddress = value
	}
	if value := os.Getenv("SWITCHBOARD_ADMIN_ADDRESS"); value != "" {
		cfg.AdminAddress = value
	}
	if value := os.Getenv("SWITCHBOARD_ROUTING_STRATEGY"); value != "" {
		cfg.RoutingStrategy = value
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var problems []error
	if err := validateAddress("proxyAddress", c.ProxyAddress); err != nil {
		problems = append(problems, err)
	}
	if err := validateAddress("adminAddress", c.AdminAddress); err != nil {
		problems = append(problems, err)
	}
	if c.ProxyAddress == c.AdminAddress {
		problems = append(problems, errors.New("proxyAddress and adminAddress must differ"))
	}
	if c.RoutingStrategy != "round-robin" && c.RoutingStrategy != "least-connections" {
		problems = append(problems, fmt.Errorf("routingStrategy must be round-robin or least-connections, got %q", c.RoutingStrategy))
	}
	if c.ShutdownTimeout.Value() <= 0 {
		problems = append(problems, errors.New("shutdownTimeout must be positive"))
	}
	if c.MetricsWindow.Value() < time.Second {
		problems = append(problems, errors.New("metricsWindow must be at least 1s"))
	}
	if len(c.Backends) == 0 {
		problems = append(problems, errors.New("at least one backend is required"))
	}
	ids := make(map[string]struct{}, len(c.Backends))
	for i, backend := range c.Backends {
		prefix := fmt.Sprintf("backends[%d]", i)
		if backend.ID == "" {
			problems = append(problems, fmt.Errorf("%s.id is required", prefix))
		}
		if _, exists := ids[backend.ID]; exists {
			problems = append(problems, fmt.Errorf("duplicate backend id %q", backend.ID))
		}
		ids[backend.ID] = struct{}{}
		parsed, err := url.ParseRequestURI(backend.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			problems = append(problems, fmt.Errorf("%s.url must be an absolute URL", prefix))
		}
		if backend.HealthPath == "" || !strings.HasPrefix(backend.HealthPath, "/") {
			problems = append(problems, fmt.Errorf("%s.healthPath must start with /", prefix))
		}
		if backend.HealthInterval.Value() <= 0 || backend.HealthTimeout.Value() <= 0 {
			problems = append(problems, fmt.Errorf("%s health durations must be positive", prefix))
		}
		if backend.HealthTimeout.Value() >= backend.HealthInterval.Value() {
			problems = append(problems, fmt.Errorf("%s healthTimeout must be less than healthInterval", prefix))
		}
		if backend.UnhealthyAfter < 1 || backend.HealthyAfter < 1 {
			problems = append(problems, fmt.Errorf("%s health thresholds must be at least 1", prefix))
		}
	}
	return errors.Join(problems...)
}

func validateAddress(name, address string) error {
	if address == "" {
		return fmt.Errorf("%s is required", name)
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		return fmt.Errorf("%s is invalid: %w", name, err)
	}
	return nil
}
