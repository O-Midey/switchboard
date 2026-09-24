# ADR 0002: Use process-local rolling metrics first

Status: Accepted

## Context

Routing needs recent latency, error, and concurrency signals. A database or telemetry cluster would add network dependencies before cross-instance history is required.

## Decision

Keep fixed one-second buckets per backend in memory. Export snapshots through Prometheus exposition and the dashboard state endpoint. Metrics are operational signals, not an authoritative record.

## Consequences

Storage is bounded and hot-path updates are cheap. History resets on restart and instances do not share state. Cross-instance policy evaluation will require an external collector later, triggered by multi-instance deployment.

