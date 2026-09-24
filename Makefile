.PHONY: all build test test-race lint fmt check compose-up compose-down smoke dashboard-install dashboard-dev

all: check

build:
	go build ./cmd/... ./internal/...

test:
	go test ./cmd/... ./internal/...

test-race:
	go test -race ./cmd/... ./internal/...

fmt:
	gofmt -w $$(find cmd internal -name '*.go')

lint:
	test -z "$$(gofmt -l cmd internal)"
	go vet ./cmd/... ./internal/...
	cd web && npm run lint && npm run typecheck

check: lint test-race build
	cd web && npm run build

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

smoke:
	./scripts/smoke.sh

dashboard-install:
	cd web && npm install

dashboard-dev:
	cd web && npm run dev
