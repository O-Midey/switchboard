# Switchboard

Switchboard is an experiment in making infrastructure more responsive.

I’m building it with **Jev**, Go, Docker, and a small live dashboard. The idea is simple: when several services can handle a request, Switchboard watches what is happening and helps choose the best place to send it.

This is the first step toward software that can make fast, contextual decisions instead of following the same static rule every time.

## Where it is now

The foundation is working: Switchboard can send traffic to three local services, notice when one is unhealthy, track what is happening, and show the system in the dashboard.

Jev is intentionally waiting behind that foundation. The next phase will let Jev suggest routing decisions, while the core system keeps the final say and continues working if the model is unavailable.

The control-plane groundwork is now in place too: candidate policies must name the configured backends, use normalized weights, carry a version, and expire. A deterministic weighted router can evaluate those policies in shadow mode while live traffic continues using the proven baseline.

## Try it locally

Requirements: Docker with Compose. This starts the proxy, three sample services, and the dashboard:

```bash
make compose-up
```

Then open [http://localhost:3000](http://localhost:3000) and send a request through [http://localhost:8080](http://localhost:8080).

Stop everything with:

```bash
make compose-down
```

The project is still early. The [production roadmap](ROADMAP.md) tracks what is complete, what comes next, and the evidence required before Switchboard can be called production-ready.

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

Read the [production roadmap](ROADMAP.md), [architecture](docs/architecture.md), [operations guide](docs/operations.md), and [ADRs](docs/adr/) before extending the routing path.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for quality gates and [SECURITY.md](SECURITY.md) for private vulnerability reporting guidance. The project is currently pre-release; interfaces may change before `v1.0.0`.
