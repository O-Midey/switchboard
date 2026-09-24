# Switchboard roadmap

This roadmap tracks the path from the current local system to a globally deployable production system. It is intentionally ordered by risk: first make the deterministic proxy trustworthy, then observe Jev safely, then allow bounded influence, and only then distribute the system across instances and regions.

“Global production level” is an outcome proved by tests, operations, and real workload evidence. It does not mean adding every distributed-systems tool in advance.

## Status legend

- ✅ Complete and verified
- 🚧 In progress
- ⬜ Not started
- ⏸ Gated by evidence or an earlier milestone

An item becomes complete only when its code, tests, documentation, and operational evidence are merged. Code existing locally is not enough.

## Progress summary

| Milestone | Outcome | Status |
|---|---|---|
| M0 | Deterministic local foundation | ✅ Complete |
| M1 | Hardened HTTP data plane | ⬜ Not started |
| M2 | Production observability and operations | ⬜ Not started |
| M3 | Jev shadow control plane | ⬜ Not started |
| M4 | Secure single-region pilot | ⏸ Gated by M1–M3 |
| M5 | Bounded Jev actuation | ⏸ Gated by pilot evidence |
| M6 | Multi-instance control plane | ⏸ Gated by scale evidence |
| M7 | Global multi-region operation | ⏸ Gated by regional demand |
| M8 | Stable v1.0 open-source release | ⏸ Gated by production evidence |

## Production targets

These are provisional engineering targets. M1 load testing and the first real workload must validate or revise them.

| Property | Target |
|---|---|
| Routing safety | Never select an unhealthy or unconfigured destination |
| Model isolation | No model call in the request path; invalid or stale policy cannot affect traffic |
| Availability | At least 99.95% monthly routing availability per production region, measured separately from upstream application failures |
| Proxy overhead | p99 added latency at or below 5 ms at the declared supported workload and no more than 70% sustained resource saturation |
| Recovery | Automatic deterministic fallback on policy expiry or control-plane failure; regional RTO at or below 15 minutes |
| Policy integrity | Every active policy is validated, versioned, attributable, bounded, auditable, and reversible |
| Security | TLS in transit, authenticated administration, least-privilege identities, secret rotation, and no sensitive data in logs or model context |
| Delivery | Reproducible builds, signed releases, staged rollout, automatic rollback signals, and a tested rollback procedure |

Throughput is deliberately not promised yet. We will publish a supported requests-per-second and concurrency envelope only after repeatable benchmarks on declared hardware.

## Non-negotiable invariants

1. The request data plane remains deterministic.
2. Jev proposes policy; trusted code validates and applies it.
3. No policy may name an unconfigured destination.
4. An unhealthy destination is never selected, regardless of policy weight.
5. Expired, malformed, unavailable, or unsafe policy falls back to the deterministic baseline.
6. The admin surface never shares the public trust boundary without authentication and authorization.
7. Retries are not enabled until method safety, body replay, retry budgets, and idempotency are defined.
8. A global failover must not silently violate security, residency, or policy constraints.

---

## M0 — Deterministic local foundation

**Outcome:** A reliable local baseline exists before AI affects traffic.

- [x] **M0-01** Concurrent Go reverse proxy
- [x] **M0-02** Three configurable mock backends
- [x] **M0-03** Active health checks with recovery/failure thresholds
- [x] **M0-04** Rolling latency, error, request, and active-request metrics
- [x] **M0-05** Round-robin and least-connections routing adapters
- [x] **M0-06** Strict configuration validation and environment overrides
- [x] **M0-07** Structured logs and graceful shutdown
- [x] **M0-08** Liveness, readiness, state, and Prometheus endpoints
- [x] **M0-09** Docker Compose environment and control-room dashboard
- [x] **M0-10** Race-tested unit/integration coverage and CI
- [x] **M0-11** Jev provider seam without live integration
- [x] **M0-12** Validated, expiring policy representation
- [x] **M0-13** Deterministic weighted router for future shadow evaluation

**Evidence:** CI passes Go race tests, dashboard checks, builds, and container builds. Local failure testing proves unhealthy backend exclusion and recovery.

---

## M1 — Harden the HTTP data plane

**Outcome:** The proxy has explicit, tested behavior for real HTTP traffic and overload.

- [ ] **M1-01** Define the public HTTP compatibility contract: streaming, cancellation, upgrades, trailers, flushing, `HEAD`, and large bodies
- [ ] **M1-02** Define trusted-proxy and forwarded-header handling; prevent client spoofing
- [ ] **M1-03** Generate or propagate request IDs and correlate proxy, health, and policy events
- [ ] **M1-04** Add configurable request-header, body, and concurrency limits without breaking streaming
- [ ] **M1-05** Add route/upstream deadlines and cancellation propagation
- [ ] **M1-06** Add admission control and load shedding with structured `429`/`503` responses
- [ ] **M1-07** Design circuit breaking and recovery without synchronized retry storms
- [ ] **M1-08** Centralize the Go error contract and boundary response shaping
- [ ] **M1-09** Add transactional configuration reload with validation and last-known-good rollback
- [ ] **M1-10** Add fuzz tests for configuration, headers, policy inputs, and proxy edge cases
- [ ] **M1-11** Add real-socket integration tests for disconnects, slow upstreams, partial responses, and shutdown draining
- [ ] **M1-12** Build reproducible benchmark and soak suites with CPU, memory, allocation, latency, and throughput reports

### M1 exit gate

- The supported HTTP feature set is documented and tested.
- Race, fuzz, soak, and failure-injection suites are green.
- Resource usage is bounded under overload.
- The first supported capacity envelope and proxy-overhead result are published with hardware and workload details.
- Rollback to the M0 behavior is documented and tested.

---

## M2 — Production observability and operations

**Outcome:** Operators can detect, explain, and recover from failures before users report them.

- [ ] **M2-01** Replace window-only telemetry with stable counters, histograms, and gauges suitable for Prometheus
- [ ] **M2-02** Add OpenTelemetry traces with request, backend, policy-version, and outcome correlation
- [ ] **M2-03** Define metric cardinality budgets and redact sensitive values before export
- [ ] **M2-04** Add golden-signal dashboards for traffic, errors, saturation, latency, and backend health
- [ ] **M2-05** Define SLIs, provisional SLOs, burn-rate alerts, and error-budget policy
- [ ] **M2-06** Add alerts for no healthy targets, policy expiry, fallback activation, saturation, and abnormal decision churn
- [ ] **M2-07** Write runbooks for upstream failure, overload, bad configuration, bad policy, control-plane outage, and rollback
- [ ] **M2-08** Add build/version/config metadata to diagnostics without exposing secrets
- [ ] **M2-09** Define log, metric, trace, and audit-event retention and ownership
- [ ] **M2-10** Run and document the first game day

### M2 exit gate

- Every claimed failure mode has a signal, alert, owner, and runbook.
- A clean operator can diagnose a staged incident using only published telemetry and documentation.
- A game day proves detection, rollback, and recovery timings.

---

## M3 — Jev shadow control plane

**Outcome:** Jev makes real policy proposals from live snapshots, but cannot affect traffic.

- [ ] **M3-01** Define the minimal, typed Jev input/output contract and version it
- [ ] **M3-02** Normalize telemetry into a bounded snapshot with freshness and provenance
- [ ] **M3-03** Implement the Jev provider adapter behind the existing `Provider` interface
- [ ] **M3-04** Apply strict timeouts, cancellation, call budgets, and safe failure normalization
- [ ] **M3-05** Run policy generation asynchronously and outside the request path
- [ ] **M3-06** Validate every proposal for membership, normalized weights, TTL, version, and rate-of-change limits
- [ ] **M3-07** Add hysteresis/cooldowns to prevent routing oscillation
- [ ] **M3-08** Compare baseline and proposed decisions in shadow mode without changing live routing
- [ ] **M3-09** Record compact decision/audit events without request bodies, secrets, or sensitive headers
- [ ] **M3-10** Build a replayable evaluation corpus from synthetic and consented/redacted traffic states
- [ ] **M3-11** Define offline evaluation metrics: invalid-policy rate, stability, latency, fallback rate, and improvement over baselines
- [ ] **M3-12** Add a global kill switch and provider-independent deterministic fallback

### M3 exit gate

- Jev never appears in request-path profiles or dependency graphs.
- Invalid and stale proposal tests achieve 100% rejection.
- Shadow evaluation runs for the agreed observation window with no unexplained policy oscillation.
- Evidence shows a measurable improvement over round-robin/least-connections for at least one declared workload.
- Provider outage and credential revocation drills leave live routing unaffected.

---

## M4 — Secure single-region pilot

**Outcome:** Switchboard can run safely for a small, explicitly allowed workload in one region.

- [ ] **M4-01** Publish minimal hardened OCI images with non-root execution, read-only filesystem support, health probes, and resource limits
- [ ] **M4-02** Add TLS termination guidance and verified upstream TLS support
- [ ] **M4-03** Protect admin endpoints with authenticated, authorized operator access
- [ ] **M4-04** Use short-lived workload identity and managed secrets; document rotation and revocation
- [ ] **M4-05** Add network policies separating public traffic, upstreams, admin access, and the Jev provider
- [ ] **M4-06** Add dependency, container, license, secret, and static security scanning to CI
- [ ] **M4-07** Produce an SBOM and sign release images and provenance
- [ ] **M4-08** Add staged deployment, readiness gates, canary rollout, and automatic rollback signals
- [ ] **M4-09** Define a pilot allowlist, traffic ceiling, owner, support hours, and incident escalation path
- [ ] **M4-10** Complete a threat model covering SSRF, header spoofing, admin compromise, policy poisoning, secret exposure, and denial of wallet

### M4 exit gate

- External security review has no unresolved critical/high findings.
- Deployment and rollback are repeatable from documentation.
- A deterministic-only pilot meets the provisional SLO and capacity target for the agreed observation window.
- Audit evidence identifies every configuration and policy change.

---

## M5 — Bounded Jev actuation

**Outcome:** Validated Jev policies may influence a small percentage of traffic under automatic safety controls.

- [ ] **M5-01** Add an atomic active-policy store with version monotonicity and last-known-good retention
- [ ] **M5-02** Gate policy activation behind configuration, environment, route, and traffic-percentage controls
- [ ] **M5-03** Start with an allowlisted canary cohort; never enable globally by default
- [ ] **M5-04** Enforce minimum/maximum weights and maximum change per policy interval
- [ ] **M5-05** Fall back automatically on expiry, validation failure, SLO burn, provider outage, or anomaly detection
- [ ] **M5-06** Expose the active policy, source, age, version, and fallback reason to authorized operators
- [ ] **M5-07** Compare live outcomes against a concurrent deterministic control group
- [ ] **M5-08** Prove manual disable and automatic rollback during a game day

### M5 exit gate

- Canary traffic meets or improves the deterministic control group without breaching SLOs.
- No request is routed by an unvalidated or expired policy.
- Automatic and manual rollback complete within the defined recovery target.
- Expansion requires explicit human approval backed by the evaluation report.

---

## M6 — Multi-instance control plane

**Outcome:** Multiple stateless data-plane instances consume one consistent, durable policy stream.

- [ ] **M6-01** Define the authoritative policy record, monotonic versioning, activation time, expiry, and audit history
- [ ] **M6-02** Select a durable policy store only after multi-instance deployment creates the need
- [ ] **M6-03** Distribute signed/versioned policy snapshots with bounded staleness and reconnect behavior
- [ ] **M6-04** Ensure each instance can route safely from local last-known-good state during control-plane partitions
- [ ] **M6-05** Add reconciliation for missed, duplicated, reordered, and conflicting policy updates
- [ ] **M6-06** Add fleet-level rollout, pause, rollback, and compatibility checks
- [ ] **M6-07** Test autoscaling, instance churn, rolling upgrades, and mixed-version compatibility
- [ ] **M6-08** Add service discovery only if measured backend-membership churn makes static rollout unsafe or too slow

### M6 adoption trigger

Do not build this milestone merely for “scale.” Start it when production requires more than one data-plane instance for capacity or availability, or when policy consistency across instances becomes an observed operational need.

### M6 exit gate

- Instance churn and control-plane partition tests preserve routing invariants.
- Policy convergence and maximum staleness are measured and meet the target.
- A fleet-wide rollback is demonstrated without restarting the entire fleet.

---

## M7 — Global multi-region operation

**Outcome:** Region-local data planes can survive regional failures without violating policy, security, or residency rules.

- [ ] **M7-01** Document supported regions, traffic origins, latency objectives, residency constraints, and provider availability
- [ ] **M7-02** Keep request handling region-local; do not add cross-region request hops without measured justification
- [ ] **M7-03** Deploy independent regional data planes with explicit capacity headroom
- [ ] **M7-04** Distribute policies globally with regional scope, signed provenance, and bounded staleness
- [ ] **M7-05** Add global traffic steering with health-aware regional failover
- [ ] **M7-06** Define whether policy state is global, regional, or hierarchical and record the consistency tradeoff in an ADR
- [ ] **M7-07** Make logs, metrics, audit events, backups, and support access comply with residency policy
- [ ] **M7-08** Test full regional evacuation, provider degradation, DNS/steering failure, and control-plane partition
- [ ] **M7-09** Validate slow-network behavior and dashboard accessibility for supported operator environments
- [ ] **M7-10** Publish regional capacity, RTO, and failure-mode evidence

### M7 adoption trigger

Begin global deployment only when users, latency, residency, or availability requirements justify a second region. Add regions one at a time; do not begin with active-active writes or a globally consistent database unless a proven invariant requires them.

### M7 exit gate

- At least two independent regions pass capacity, failover, and security tests; a third region validates repeatability before claiming a global operating model.
- Regional failure does not route to an unhealthy or policy-incompatible destination.
- RTO, policy-staleness, and routing SLOs are met during a documented regional game day.

---

## M8 — Stable v1.0 open-source release

**Outcome:** Operators can adopt, upgrade, secure, and support Switchboard without relying on repository authors.

- [ ] **M8-01** Freeze and document public configuration, policy, metrics, and admin contracts
- [ ] **M8-02** Publish compatibility and deprecation policy with semantic versioning
- [ ] **M8-03** Add upgrade, rollback, backup, and configuration-migration guides
- [ ] **M8-04** Publish reproducible signed binaries/images, checksums, SBOM, and release notes
- [ ] **M8-05** Add end-to-end examples for deterministic-only, shadow, and bounded-actuation modes
- [ ] **M8-06** Define supported versions, vulnerability response targets, disclosure contact, and maintenance ownership
- [ ] **M8-07** Complete independent security and reliability reviews
- [ ] **M8-08** Prove the documented production story with a clean-room installation and upgrade test

### M8 exit gate

- All declared v1 contracts have compatibility tests.
- Installation, upgrade, rollback, and incident procedures work from published artifacts.
- Production evidence supports every reliability and performance claim in the README.

---

## Explicitly deferred

These are not roadmap defaults. They require a measurable trigger and an ADR before adoption:

- Kubernetes or a service mesh
- A database, queue, cache, or stream platform
- Dynamic service discovery
- Cross-region writes or globally consistent state
- Request retries
- LLM calls in the request path
- Model-generated destinations, credentials, code, or executable actions
- Autonomous self-modification or self-deployment

## Open decisions and assumptions

These questions must be resolved with evidence before their dependent milestone begins:

| Question | Current assumption | Blocks |
|---|---|---|
| What traffic will Switchboard serve first? | HTTP APIs with streaming bodies; no claim yet for WebSockets/gRPC | M1 compatibility contract and benchmarks |
| What is the expected steady/peak load? | Unknown; benchmark ranges will be exploratory, not promises | Capacity targets, M4 pilot limits |
| Where will the first pilot run? | One region on a managed container platform or simple orchestrator | M4 deployment design |
| Is Kubernetes required? | No; adopt only when the chosen environment or fleet operations justify it | Packaging and rollout implementation |
| Is Switchboard multi-tenant? | No tenant control-plane model is assumed yet | Authentication, authorization, isolation, audit design |
| Which regions and residency rules apply? | Unknown; no cross-region data movement is assumed | M7 topology and providers |
| What Jev limits and guarantees apply? | The adapter must tolerate timeouts, rate limits, invalid output, and total outage | M3 budgets and evaluation design |
| Who owns on-call and incident response? | Repository owner during development; explicit rotation required before production | M2/M4 operations gates |
| What is the operating budget? | Keep request-path cost model-only and bound Jev calls by cadence and budget | M3 call frequency, telemetry retention, M7 footprint |

## Capacity and cost discipline

The dominant resources are expected to be proxy CPU/memory, network egress, telemetry volume, regional replicas, and Jev evaluations. Each milestone must measure these units directly.

- Jev evaluation is periodic or event-triggered, never per request.
- Every provider adapter must have request, token/compute, concurrency, and spend budgets.
- Telemetry retention and metric labels must be bounded before production ingestion.
- Regional expansion requires measured latency/availability demand and a capacity plan.
- A database, queue, cache, or discovery platform needs a documented trigger, owner, rollback path, and cost estimate.
- Load tests must include normal traffic, bursts, slow upstreams, large bodies, unhealthy targets, and recovery—not only a happy-path maximum RPS number.

## Tracking rules

1. Keep the stable milestone IDs in issue titles and pull requests, for example: `M1-03: propagate request IDs`.
2. Mark an item complete only in the same pull request that adds its evidence.
3. Link architecture-changing work to an ADR.
4. Record benchmark hardware, workload, raw results, and commands; do not publish unsupported performance claims.
5. Treat local tests, CI, deployment, smoke tests, canary results, and production SLOs as separate evidence.
6. Revisit targets after M1 benchmarks, the M4 pilot, and every new region.
7. If a phase fails its exit gate, roll back or remain in the current phase; do not relabel incomplete work as production-ready.

## Next development focus

The next work should start at **M1-01 through M1-03**:

1. Specify the supported HTTP behavior and trust model.
2. Implement trusted forwarded-header handling.
3. Add request identity and end-to-end correlation.

Those three items define the safe edge contract that the remaining reliability, observability, and Jev work will depend on.
