# pave — Pavestack Self-Service CLI

The `pave` CLI is the developer entry point to the Pavestack Internal Developer Platform. It scaffolds new services from the golden-path template and generates all required GitOps manifests.

## Constraint

The CLI is a **thin orchestrator**. It creates files and opens Pull Requests. It **never** runs `kubectl apply`, `helm install`, or any command that mutates live cluster state. Deployment is handled exclusively by Argo CD after PR merge.

## Installation

```bash
# From monorepo root
make pave
./bin/pave --version

# Or directly
cd pave && go build -o pave ./cmd/pave
```

## Usage

### `pave create-service`

Scaffold a new internal API service and generate GitOps manifests.

```bash
# Interactive — prompts for missing values
pave create-service

# Non-interactive — all flags provided
pave create-service \
  --name payments \
  --team team-payments \
  --database=false

# Skip automatic PR creation
pave create-service \
  --name payments \
  --team team-payments \
  --database=false \
  --no-pr
```

#### Flags

| Flag | Required | Default | Description |
|---|---|---|---|
| `--name` | Yes | (prompted) | DNS-safe service identifier (3–50 chars, lowercase) |
| `--team` | Yes | (prompted) | Owning team slug (3–64 chars, lowercase) |
| `--database` | No | `false` | Whether the service requires a managed database |
| `--no-pr` | No | `false` | Skip automatic GitHub PR creation |
| `--branch` | No | `pave/create-{name}-api` | Custom branch name for the PR |
| `--runtime` | No | `go` | Golden-path runtime. Only `go` is scaffoldable today. |
| `--exposure` | No | `internal` | `internal` (ClusterIP only) or `public` (adds an ALB — see cost estimate). |
| `--tier` | No | `tier-2` | `tier-1` (critical), `tier-2` (standard), `tier-3` (low-traffic). Drives replica count, resource sizing, and the printed cost estimate. |

Every run prints an estimated monthly cost range for the chosen tier/exposure/database combination before scaffolding — see `pave/internal/cost` for the (static, documented) pricing model.

### What it does

1. **Validates** the request against `schemas/service-request.schema.json`
2. **Copies** `/service-template-api` → `/services/{name}-api`
3. **Replaces** template references (module paths, service name, team owner, Helm chart name)
4. **Generates GitOps manifests** in `/platform-config/tenants/{name}/`:
   - `tenant.yaml` — tenant metadata for ApplicationSet
   - `base/kustomization.yaml` — namespace, RBAC, NetworkPolicy, ResourceQuota
   - `dev/values.yaml` and `dev/application.yaml` — dev environment
   - `prod/values.yaml` and `prod/application.yaml` — prod environment
5. **Opens a PR** via `gh pr create` (unless `--no-pr`)

### Post-scaffold workflow

```
pave create-service ──► git commit ──► git push ──► PR merge ──► Argo CD reconciles
```

## Architecture

```
cmd/pave/                CLI entrypoint
cmd/pave-api/             HTTP backend for the self-service portal (see below)
internal/cli/             Cobra command tree (root, create-service)
internal/scaffold/        Template copy, string replacement, Helm chart renaming
internal/gitops/          Tenant manifest generation, PR creation
internal/validate/        JSON Schema validation of service requests
internal/cost/             Tier sizing + static monthly cost estimation
internal/workspace/        Shared repo-root resolution (used by both cli and pave-api)
internal/apiserver/        pave-api HTTP handlers, job runner, catalog reader
schemas/                  JSON Schema definitions
```

## Environment variables

| Variable | Description |
|---|---|
| `PAVESTACK_ROOT` | Override monorepo root detection (auto-detected by default) |

## `pave-api` — backend for the self-service portal

`pave-api` is a thin HTTP shell around the exact same `scaffold`/`gitops`/`validate`
packages the CLI calls directly. The portal's "Create service" wizard and
"Request access" flow talk to it — see `pavestack-portal/README.md` for the
frontend side and the full route contract.

```bash
cd pave && go build -o ../bin/pave-api ./cmd/pave-api
PAVESTACK_ROOT=/path/to/pavestack ./bin/pave-api
# -> pave-api listening on :8787
```

| Variable | Default | Description |
|---|---|---|
| `PAVE_API_PORT` | `8787` | Listen port |
| `PAVE_API_REPO_ROOT` | auto-detected | Repo root to scaffold into — **point this at a scratch clone** unless you want it to really write into your checkout |
| `PAVE_API_DRY_RUN` | `true` | Set to `false` to actually shell out to `git`/`gh` and open a real PR on job completion. Scaffold + GitOps manifest writes always happen for real regardless of this flag — only the final PR step is gated. |
| `PAVE_API_CORS_ORIGIN` | `*` | CORS origin allowed to call the API (set to the portal's origin in production) |

Routes: `GET /healthz`, `GET|POST /api/v1/services`, `GET /api/v1/services/:name`,
`GET /api/v1/jobs/:id`, `GET /api/v1/cost-estimate`, `GET|POST /api/v1/access-requests`,
`PATCH /api/v1/access-requests/:id`. Access requests persist to
`<repoRoot>/.pave-api/access-requests.json` (gitignored) so approvals survive a restart.

## Development

```bash
go test ./...
go vet ./...
go build -ldflags="-X github.com/pavestack/pave/internal/cli.Version=1.0.0" -o pave ./cmd/pave
go build -o pave-api ./cmd/pave-api
```
