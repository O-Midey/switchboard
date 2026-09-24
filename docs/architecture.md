# Architecture

## Scope and assumptions

Milestone one targets a single stateless proxy process, three static HTTP upstreams, and local/operator use. It optimizes for a trustworthy routing baseline rather than premature distribution. Expected initial load is up to thousands of requests per second per instance, subject to benchmark validation; request bodies are streamed, while metrics remain bounded by backend count and rolling-window seconds.

Non-goals are TLS termination, persistence, retries, dynamic service discovery, authentication, multi-region coordination, and AI-selected routing.

## System shape

```text
                              later control plane (not connected)
                         snapshots -> Jev provider -> validator
                                                | expiring policy
                                                v
client -> proxy handler -> deterministic Router -> healthy upstream
               |                    ^                    |
               |                    |                    v
               +-> rolling metrics -+<- health checker <-+
                         |
                         +-> admin state / Prometheus -> dashboard
```

The `Router` interface is the data-plane seam. It receives a request description and the configured targets, filters on authoritative health, then returns one target. The policy `Provider` is the future control-plane seam. It can only propose versioned, expiring weights; a validator and deterministic weighted router will mediate any future use.

## Correctness invariants

1. An unhealthy target is never selected.
2. No model call, dashboard call, or telemetry export occurs in the routing hot path.
3. Target active-request accounting is balanced around every proxy attempt.
4. Configuration is fully validated before listeners start.
5. Readiness fails closed when every target is unhealthy; liveness remains independent.
6. Policy input and output are untrusted, bounded data. A future model cannot name arbitrary destinations.

## Failure behavior

| Failure | Current behavior | Follow-up trigger |
|---|---|---|
| One upstream fails | Health threshold excludes it; other healthy targets continue | Tune thresholds from observed failure/recovery data |
| All upstreams fail | Proxy returns structured `503`; readiness fails | Add controlled degraded responses only with product semantics |
| Upstream times out before headers | Transport emits structured `502` | Introduce per-route deadlines when workload classes exist |
| Admin/dashboard fails | Proxy continues unaffected | Split deployment only if resource contention is measured |
| Process receives SIGTERM | Both listeners drain within the configured deadline | Add orchestration pre-stop timing in deployment manifests |
| Future policy provider fails | Use last valid unexpired policy, then deterministic baseline | Implement with the first Jev milestone |

## Security and trust

The admin listener is a privileged operational surface and should remain on a private network or loopback. Upstream URLs come only from startup configuration; requests cannot choose a destination, preventing an SSRF routing primitive. The proxy forwards ordinary HTTP headers according to Go's reverse-proxy behavior; production deployment must define trusted proxy hops and an explicit forwarded-header policy before accepting Internet traffic.

Containers run as non-root with `no-new-privileges`. Secrets are not required in milestone one. Future provider credentials must stay in the control plane and must never enter request logs, dashboard payloads, or model context.

## Evolution gates

- Add weighted policy routing only after deterministic baselines have benchmark and failure-injection results.
- Add external metrics storage only when retention or cross-instance aggregation is required.
- Add service discovery only when backend membership changes faster than safe configuration rollout.
- Add a queue only if policy evaluation must be buffered/replayed; it is not part of request handling.
- Split control and data plane deployments only when independent scaling, permissions, or failure isolation justify the operational cost.

