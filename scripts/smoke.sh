#!/bin/sh
set -eu

curl --fail --silent http://localhost:9090/-/readyz >/dev/null
curl --fail --silent http://localhost:8080/smoke >/dev/null
curl --fail --silent http://localhost:9090/api/v1/state >/dev/null
curl --fail --silent http://localhost:9090/metrics | grep --quiet switchboard_backend_healthy
echo "Switchboard smoke test passed."

