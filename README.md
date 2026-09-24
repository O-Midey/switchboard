# Switchboard

Switchboard is an adaptive reverse proxy with a deterministic data plane and an intentionally separate policy control plane. Milestone one establishes the trustworthy baseline: concurrent proxying, active health checks, rolling backend telemetry, round-robin and least-connections routing, operational endpoints, graceful shutdown, and a control-room dashboard.

Jev is **not** connected to request routing yet. The future model may propose bounded, expiring weights or policies; ordinary Go code validates and executes them. If the model or control plane fails, the proxy keeps serving with its last valid policy or a deterministic baseline.

## Quick start

Requirements: Docker with Compose. Then:

```bash
make compose-up
curl http://localhost:8080/hello
open http://localhost:3000
make smoke
```

The proxy listens on `:8080`. The admin interface is bound to `127.0.0.1:9090` by Compose and exposes:

- `GET /-/healthz` — process liveness
- `GET /-/readyz` — at least one healthy upstream
- `GET /api/v1/state` — dashboard state contract
- `GET /metrics` — Prometheus text exposition

Stop the stack with `make compose-down`.

## Local development

Run each mock backend in a separate terminal, then start Switchboard with the local configuration:

```bash
go run ./cmd/mock-backend -id atlas -port 8081 -latency 20ms -error-rate 0.01
go run ./cmd/mock-backend -id boreal -port 8082 -latency 75ms -error-rate 0.03
go run ./cmd/mock-backend -id cirrus -port 8083 -latency 180ms
go run ./cmd/switchboard -config configs/switchboard.local.json
```

Install and run the dashboard with `make dashboard-install dashboard-dev`. Use `make check` for the same lint, race-test, build, and dashboard checks expected by CI.

## Configuration

Configuration is strict JSON: unknown fields and unsafe values fail startup. See [`configs/switchboard.local.json`](configs/switchboard.local.json). Listener addresses and routing strategy may be overridden with environment variables documented in [`.env.example`](.env.example). Backend configuration is immutable for this milestone; restart to apply it.

## Repository map

```text
cmd/switchboard       process assembly and lifecycle
cmd/mock-backend      configurable upstream simulator
internal/backend      target health state and rolling metrics
internal/router       deterministic routing interface and adapters
internal/proxy        reverse-proxy data plane
internal/health       active upstream probing
internal/admin        readiness, state, and Prometheus endpoints
internal/policy       future Jev policy-provider contract (not wired)
web                   Next.js control-room dashboard
docs                  architecture, operations, and decisions
```

## Current guarantees and limits

- Routing never selects an unhealthy backend.
- Shared routing, health, and metric state is concurrency-safe and race-tested.
- Shutdown stops new work and drains HTTP servers within the configured deadline.
- Telemetry is in-memory and process-local. It resets on restart and is not a billing/audit source.
- This milestone has no retry layer. Automatic proxy retries can duplicate unsafe requests and require a request replay/idempotency design first.
- TLS, external authentication, rate limiting, dynamic discovery, and multi-instance policy distribution belong at later deployment milestones.

Read [the architecture](docs/architecture.md), [operations guide](docs/operations.md), and [ADRs](docs/adr/) before extending the routing path.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for quality gates and [SECURITY.md](SECURITY.md) for private vulnerability reporting guidance. The project is currently pre-release; interfaces may change before `v1.0.0`.

