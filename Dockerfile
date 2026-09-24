FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/switchboard ./cmd/switchboard && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mock-backend ./cmd/mock-backend

FROM gcr.io/distroless/static-debian12:nonroot AS switchboard
COPY --from=build /out/switchboard /switchboard
ENTRYPOINT ["/switchboard"]

FROM gcr.io/distroless/static-debian12:nonroot AS mock-backend
COPY --from=build /out/mock-backend /mock-backend
ENTRYPOINT ["/mock-backend"]

