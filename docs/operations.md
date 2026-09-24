# Operations

## Signals

Use `/-/healthz` for process liveness and `/-/readyz` for traffic readiness. Scrape `/metrics` from the private admin listener. The first alerts should cover no healthy backends, sustained upstream error rate, rising latency, and abnormal active-request growth.

Logs are newline-delimited JSON in the proxy process. Request logs include method, path, status, selected backend, routing strategy, and duration. They intentionally omit bodies, query values, and authorization headers.

## Shutdown and recovery

SIGINT or SIGTERM cancels health probes and asks both HTTP servers to drain. The process exits after both stop or the configured timeout elapses. Metrics are ephemeral and recover automatically from new traffic after restart.

## Local failure drill

1. Start the Compose stack and generate traffic through port 8080.
2. Stop one backend container.
3. Confirm its state becomes unhealthy after two failed checks and requests continue through the others.
4. Stop the remaining backends and confirm proxy `503` plus readiness failure.
5. Restore a backend and confirm it becomes selectable after one successful check.

Do not treat this drill as evidence for production SLOs. Benchmarking, soak tests, resource limits, TLS, hardened network policy, and deployment rollback proof remain required.

