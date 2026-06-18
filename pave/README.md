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
cmd/pave/               CLI entrypoint
internal/cli/            Cobra command tree (root, create-service)
internal/scaffold/       Template copy, string replacement, Helm chart renaming
internal/gitops/         Tenant manifest generation, PR creation
internal/validate/       JSON Schema validation of service requests
schemas/                 JSON Schema definitions
```

## Environment variables

| Variable | Description |
|---|---|
| `PAVESTACK_ROOT` | Override monorepo root detection (auto-detected by default) |

## Development

```bash
go test ./...
go vet ./...
go build -ldflags="-X github.com/pavestack/pave/internal/cli.Version=1.0.0" -o pave ./cmd/pave
```
