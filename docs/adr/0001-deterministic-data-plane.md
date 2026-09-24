# ADR 0001: Keep the request data plane deterministic

Status: Accepted

## Context

Jev is valuable for contextual policy decisions, but a model call in the request path would add variable latency, availability coupling, cost, and non-deterministic failure modes.

## Decision

The data plane uses concurrency-safe Go routers only. A later control plane may periodically propose versioned, expiring routing weights. Deterministic code validates backend membership, ranges, freshness, and fallback before atomically publishing a policy.

## Consequences

The proxy can continue when Jev is unavailable and every routed request is explainable by a concrete policy version. Adaptation is slightly delayed by the policy interval; that is an intentional reliability tradeoff.

