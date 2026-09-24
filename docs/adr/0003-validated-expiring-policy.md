# ADR 0003: Validate and expire policy proposals before use

Status: Accepted

## Context

Jev will eventually propose routing weights from live observations. Model output is untrusted and may be incomplete, stale, malformed, or refer to a destination that the process does not own.

## Decision

Policy proposals must include a version, an expiry within a bounded TTL, exactly the configured backend IDs, and weights totaling 10,000 basis points. Only the private `ValidatedPolicy` representation can enter the future policy store or candidate weighted adapter. Expired policies are unavailable and must fall back to the deterministic baseline.

The weighted adapter is evaluated in shadow mode first. It does not replace the live router until shadow results, failure behavior, and rollback semantics have evidence behind them.

## Consequences

The control plane has a small, testable seam and cannot silently route to arbitrary destinations or retain stale model decisions. The first Jev integration will need a shadow comparison and telemetry path before it is allowed to affect requests.

