<p align="center">
  <img src="brand/mark.svg" alt="Pavestack" width="72" height="72" />
</p>

<h1 align="center">Pavestack</h1>

<p align="center">
  <b>Provision standardized internal API services via GitOps — no platform-team approval queue, no standing cluster credentials.</b>
</p>

<p align="center">
  <a href="https://github.com/pavestack/pavestack/actions/workflows/monorepo-security.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/monorepo-security.yml/badge.svg" alt="Security Checks" /></a>
  <a href="https://github.com/pavestack/pavestack/actions/workflows/platform-infra.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/platform-infra.yml/badge.svg" alt="Infrastructure CI" /></a>
  <a href="https://github.com/pavestack/pavestack/actions/workflows/pave-cli.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/pave-cli.yml/badge.svg" alt="CLI CI" /></a>
  <a href="https://github.com/pavestack/pavestack/actions/workflows/scorecard.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/scorecard.yml/badge.svg" alt="OSSF Scorecard" /></a>
  <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License" />
</p>

Pavestack is an Internal Developer Platform (IDP) monorepo. A platform
team operates shared infrastructure and a golden-path service template;
product engineers self-service provision standardized internal API
services entirely through GitOps — `pave create-service` (CLI or the
portal's wizard, both calling the same backend logic) scaffolds the
service, writes GitOps manifests, and opens a pull request. CI builds,
scans, SBOMs, and signs the image; once the PR merges, Argo CD reconciles
the cluster to match Git via a canary rollout with automated
analysis-based rollback. No `kubectl apply` from a laptop, ever.

## Live surfaces

- **Landing page** (`landing/`) — the public/marketing site. Preview
  locally with `npx serve landing` (or point any static host at the
  directory — it's plain HTML/CSS/JS, no build step). Not deployed to a
  public URL as part of this exercise; see `landing/README.md` for
  one-command deploy instructions once you have a target.
- **Self-service portal** (`pavestack-portal/`) — the internal application
  developers use daily (catalog, create-service wizard, access requests,
  scorecards, observability, docs). Run locally with `make build-portal`
  then serve `pavestack-portal/out/`, or `npm run dev` inside
  `pavestack-portal/` for a live-reloading dev server. The create-service
  and access-request flows need `pave-api` running alongside it (see
  below) to do anything beyond showing an honest "backend unreachable"
  state.

## Repository layout

| Directory | Role |
|-----------|------|
| [`brand/`](brand/) | Canonical logo, favicon, and design tokens shared by the landing page and the portal |
| [`platform-infra/`](platform-infra/) | Terraform modules (VPC, EKS, ECR, GitHub OIDC, Argo CD bootstrap) |
| [`platform-config/`](platform-config/) | GitOps manifests reconciled by Argo CD — tenants, cluster-wide policies (Kyverno), observability (SLO alerts + dashboards), platform add-ons |
| [`service-template-api/`](service-template-api/) | Golden-path Go API scaffold — Argo Rollouts canary, OTel traces+metrics, trace-correlated logs |
| [`pave/`](pave/) | Self-service CLI (`pave create-service`) and `pave-api`, the HTTP backend the portal calls |
| [`landing/`](landing/) | Public marketing site |
| [`pavestack-portal/`](pavestack-portal/) | The self-service developer portal |
| [`services/`](services/) | Generated internal API services |

## Architecture

```
Developer
   │  pave create-service (CLI, or the portal wizard → pave-api)
   ▼
services/<name>-api            platform-config/tenants/<name>/
   │                                    │
   └──────────────┬─────────────────────┘
                   ▼
         git commit · push · PR
                   │
     CI: test → Trivy/Checkov/Gitleaks → build → SBOM (Syft) → sign (Cosign) → push to ECR
                   │
              PR merges
                   ▼
   Argo CD (ApplicationSets, one per tenant) reconciles
                   ▼
   EKS: Kyverno-policed, default-deny + DNS-allow NetworkPolicy,
   LimitRange/ResourceQuota, Argo Rollouts canary with
   Prometheus-analysis auto-rollback
                   ▼
   OTel traces + metrics + trace-correlated logs, SLO burn-rate
   alerts, Grafana dashboards-as-code
```

## Delivery model (GitOps)

1. **CI builds artifacts** — tests, security scans (Checkov, Trivy,
   Gitleaks, TFLint), SBOM generation (Syft) and keyless image signing
   (Cosign), container images pushed to ECR.
2. **CI opens PRs** — image tags and scaffold changes land in
   `platform-config`.
3. **Argo CD reconciles** — cluster state follows git via a canary
   rollout (Argo Rollouts); Argo CD Notifications alerts on drift/failed
   sync instead of silently self-healing. No `kubectl apply` from
   developer laptops.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | >= 1.23 | Build `pave`, `pave-api`, and `service-template-api` |
| Node.js | >= 20 | Build `pavestack-portal` |
| Terraform | >= 1.6 | Manage `platform-infra` |
| Docker | >= 24 | Build container images |
| `gh` (GitHub CLI) | >= 2 | Automated PR creation (optional) |
| AWS CLI | >= 2 | Cloud operations (optional for dev) |

## Quick start

### Platform infrastructure

```bash
# 1. Bootstrap remote state
cd platform-infra/bootstrap/remote-state && terraform init && terraform apply

# 2. Deploy dev environment
cd platform-infra/envs/dev
cp backend.hcl.example backend.hcl   # S3 backend with use_lockfile = true
terraform init -backend-config=backend.hcl
terraform apply
```

Register the root Application manually once per cluster:

```bash
kubectl apply -f platform-config/clusters/dev/root-application.yaml
```

This also bootstraps the cluster-wide add-ons (Kyverno + policies, Argo
Rollouts, kube-prometheus-stack, SLO alerts/dashboards) via
`platform-config/clusters/dev/platform-addons.yaml`.

### Scaffold a service

```bash
make pave
./bin/pave create-service --name payments --team team-payments --tier tier-2 --exposure internal --database=false
```

This prints an estimated monthly cost, copies `service-template-api` →
`services/payments-api`, and writes `platform-config/tenants/payments/`.

### Run the portal + its backend locally

```bash
cd pave && go build -o ../bin/pave-api ./cmd/pave-api && cd ..
PAVESTACK_ROOT="$(pwd)" ./bin/pave-api &          # http://localhost:8787

make build-portal
# static site in pavestack-portal/out/, or:
cd pavestack-portal && npm run dev                 # http://localhost:5173
```

### Run tests

```bash
make test
```

## Security defaults

- Distroless/non-root containers, SBOM + Cosign-signed images
- Default-deny tenant NetworkPolicies with an explicit DNS-egress allow
- Kyverno cluster policies: no privileged containers, mandatory resource
  limits, mandatory ownership labels, no `:latest` image tags
- GitHub OIDC for CI, IRSA for the EBS CSI driver (no long-lived AWS keys
  in repos)
- Argo CD Notifications alert on sync drift/failure instead of silently
  self-healing
- Terraform remote state in S3 with native lockfiles (`use_lockfile =
  true`, no DynamoDB)
- OpenTelemetry traces + metrics + trace-correlated structured logs in
  services
- KMS encryption for EKS secrets and S3 state
- Immutable ECR image tags with scan-on-push

## Progressive delivery & observability

`service-template-api` deploys as an Argo Rollouts canary; a Prometheus
`AnalysisTemplate` running during the rollout automatically aborts and
rolls back on a failing success-rate check. Multi-window multi-burn-rate
SLO alerts and a Grafana dashboard are checked into
`platform-config/observability/` as code, not clicked together in a UI.

## FinOps

Every AWS resource and Kubernetes namespace/workload is tagged for cost
attribution (`CostCenter`, `Team` — the latter matching the
`pavestack.io/team` k8s label). `platform-infra` PRs get an Infracost cost
estimate comment; `pave create-service` prints a cost range before
scaffolding. See `AGENTS.md` "cost-tagging convention".

## CI workflows

| Workflow | Component | Scanners |
|----------|-----------|----------|
| `.github/workflows/platform-infra.yml` | Terraform | Checkov, Trivy, TFLint, Infracost |
| `.github/workflows/platform-config.yml` | GitOps manifests | Checkov, Trivy |
| `.github/workflows/service-template-api.yml` | API template | Checkov, Trivy, SBOM (Syft), Cosign signing |
| `.github/workflows/pave-cli.yml` | pave CLI / pave-api | Checkov, Trivy |
| `.github/workflows/pavestack-portal.yml` | Portal | Checkov, Trivy |
| `.github/workflows/monorepo-security.yml` | Repository-wide guardrails | Checkov, Trivy, TFLint, Gitleaks |
| `.github/workflows/scorecard.yml` | Supply-chain posture | OSSF Scorecard |
| `.github/workflows/codeql.yml` | Static analysis | CodeQL (go/js/ts) |

## Environment outputs

Terraform environments expose: `cluster_name`, `cluster_endpoint`, `vpc_id`, `private_subnet_ids`, `public_subnet_ids`, `subnet_ids`, `ecr_repository_urls`.

## Documentation

- [`AGENTS.md`](AGENTS.md) — architectural rules for anyone working in this repo
- [`INTENT_SPEC.md`](INTENT_SPEC.md) — what Pavestack is, and its explicit anti-goals
- [`.agents/memory/decisions.md`](.agents/memory/decisions.md) — the architecture audit and every decision's "why"
- [`.agents/memory/lessons_learned.md`](.agents/memory/lessons_learned.md) — non-obvious things discovered along the way
- [`platform-config/GOLDEN_PATH_VERSIONING.md`](platform-config/GOLDEN_PATH_VERSIONING.md) — golden-path template versioning/deprecation policy

## Contributing

1. Fork the repository and create a feature branch
2. Make changes following existing patterns — all code must be production-aligned
3. Run `make test` and `make lint` before opening a PR
4. Ensure CI passes: security scans (Checkov, Trivy, Gitleaks, TFLint), SBOM/signing, and OSSF Scorecard must be green
5. Follow the GitOps constraint: CI builds artifacts and opens PRs; never run `kubectl apply` directly
