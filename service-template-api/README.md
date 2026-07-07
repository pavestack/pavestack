# service-template-api

Golden-path scaffold for internal API services in the Pavestack platform.

## Features

- **Go API server** with `/health`, `/ready`, and `/openapi.json` endpoints
- **Contract-first API**: `openapi.yaml` is the source of truth and is served live at `/openapi.json`
- **OpenTelemetry** instrumentation hooks (traces via OTLP/HTTP)
- **Structured logging** via `go.uber.org/zap` (JSON output, ISO 8601 timestamps)
- **Multi-stage Dockerfile** — `distroless/static-debian12:nonroot` final image (~2 MB)
- **Helm chart** with sane CPU/memory limits, liveness/readiness probes, non-root security context
- **Default-deny NetworkPolicy** per tenant namespace

## Project layout

```
cmd/server/         Entrypoint — signal handling, graceful shutdown
internal/config/    Environment-based configuration (SERVICE_NAME, LOG_LEVEL, etc.)
internal/logging/   Structured zap logger factory
internal/server/    HTTP handler with health, ready, and openapi.json routes
internal/telemetry/ OpenTelemetry TracerProvider bootstrap
deploy/helm/        Helm chart with Deployment, Service, ServiceAccount, NetworkPolicy
openapi.yaml         OpenAPI 3.1 contract — the source of truth for the HTTP surface
openapi.go           go:embed of openapi.yaml, exposed as JSON for /openapi.json
```

## Contract-first API

`openapi.yaml` at the repo root is the source of truth for this service's HTTP
surface. Workflow:

1. Edit `openapi.yaml` first when adding, changing, or removing an endpoint.
2. Update the handler in `internal/server` to match.
3. The spec is embedded via `go:embed` (see `openapi.go`) and converted to
   JSON once at startup; the running service serves it live at
   `GET /openapi.json`, so consumers always see what's actually deployed.
4. CI lints `openapi.yaml` with [Spectral](https://github.com/stoplightio/spectral)
   (`.spectral.yaml`, extending `spectral:oas`) on every push/PR.

When this template is scaffolded via `pave create-service`, `openapi.yaml`'s
title, description, and server description are automatically rewritten with
the new service name, the same way `go.mod` and `catalog-info.yaml` are.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `SERVICE_NAME` | `service-template-api` | OpenTelemetry service name and health payload |
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `LOG_LEVEL` | `info` | One of: `debug`, `info`, `warn`, `error` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (empty) | OTLP/HTTP endpoint for traces |
| `READY` | `true` | Readiness probe initial state |

## Build and run locally

```bash
go build -o server ./cmd/server
./server
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Docker

```bash
docker build -t service-template-api .
docker run --rm -p 8080:8080 service-template-api
```

## Tests

```bash
go test ./...
```

## Helm chart

```bash
helm template my-release deploy/helm/service-template-api -f deploy/helm/service-template-api/values.yaml
```

Default resource limits:

| | Requests | Limits |
|---|---|---|
| CPU | 100m | 500m |
| Memory | 128Mi | 512Mi |

## CI/CD flow

1. Push to `main` triggers `.github/workflows/service-template-api.yml`
2. Go tests, an OpenAPI spec lint (Spectral), and security scans (Trivy, Checkov, Gitleaks) run
3. Docker image is built and pushed to ECR with `git rev-parse --short HEAD` as tag
4. Trivy scans the container image
5. CI opens a PR to update `platform-config/tenants/service-template-api/*/values.yaml` with the new image tag
6. After merge, Argo CD reconciles the new image to EKS

## Metadata

- [catalog-info.yaml](catalog-info.yaml) — Backstage-compatible service catalog descriptor
- [scorecard.yaml](scorecard.yaml) — Platform compliance scorecard
