<p align="center">
  <img src="assets/banner.svg" alt="PaveStack Banner" width="100%" />
</p>

<p align="center">
  <!-- Social Badges -->
  <a href="https://github.com/pavestack/pavestack/stargazers"><img src="https://img.shields.io/github/stars/pavestack/pavestack?style=social" alt="Stars" /></a>
  <a href="https://github.com/pavestack/pavestack/network/members"><img src="https://img.shields.io/github/forks/pavestack/pavestack?style=social" alt="Forks" /></a>
  <br>
  <!-- Status Badges -->
  <a href="https://github.com/pavestack/pavestack/actions/workflows/monorepo-security.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/monorepo-security.yml/badge.svg" alt="Security Checks" /></a>
  <a href="https://github.com/pavestack/pavestack/actions/workflows/platform-infra.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/platform-infra.yml/badge.svg" alt="Infrastructure CI" /></a>
  <a href="https://github.com/pavestack/pavestack/actions/workflows/pave-cli.yml"><img src="https://github.com/pavestack/pavestack/actions/workflows/pave-cli.yml/badge.svg" alt="CLI CI" /></a>
  <img src="https://img.shields.io/github/last-commit/pavestack/pavestack" alt="Last Commit" />
  <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License" />
</p>

<p align="center">
  <a href="https://git.io/typing-svg"><img src="https://readme-typing-svg.herokuapp.com?font=Fira+Code&pause=1000&color=0366d6&center=true&vCenter=true&width=435&lines=Internal+Developer+Platform;GitOps+Delivery;Golden+Path+Scaffolding" alt="Typing SVG" /></a>
</p>

<h3 align="center">Tech Stack</h3>
<p align="center">
  <a href="https://skillicons.dev">
    <img src="https://skillicons.dev/icons?i=go,nodejs,ts,terraform,docker,aws,kubernetes,githubactions&perline=8" alt="Tech Stack" />
  </a>
</p>


<br />

Pavestack is an Internal Developer Platform (IDP) MVP delivered as a monorepo. Platform teams operate shared infrastructure; product teams scaffold services via the golden path and deploy through GitOps.

## Repository layout

| Directory | Role |
|-----------|------|
| [`platform-infra/`](platform-infra/) | Terraform modules (VPC, EKS, ECR, GitHub OIDC, Argo CD bootstrap) |
| [`platform-config/`](platform-config/) | GitOps manifests reconciled by Argo CD |
| [`service-template-api/`](service-template-api/) | Golden-path Go API scaffold |
| [`pave/`](pave/) | Self-service CLI (`pave create-service`) |
| [`pavestack-portal/`](pavestack-portal/) | Read-only catalog and scorecards |
| [`services/`](services/) | Generated internal API services |

## Architecture

![Architecture Flowchart](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggTFJcbiAgQVtEZXZlbG9wZXJdIC0tPiBCW3BhdmUgQ0xJXVxuICBCIC0tPiBDW3NlcnZpY2VzL25hbWUtYXBpXVxuICBCIC0tPiBEW3BsYXRmb3JtLWNvbmZpZy90ZW5hbnRzXVxuICBFW0dpdEh1YiBBY3Rpb25zXSAtLT4gRltBbWF6b24gRUNSXVxuICBFIC0tPiBHW0dpdE9wcyBQUl1cbiAgRyAtLT4gRFxuICBIW0FyZ28gQ0RdIC0tPiBJW0FtYXpvbiBFS1NdXG4gIEQgLS0+IEhcbiAgSltwYXZlc3RhY2stcG9ydGFsXSAtLT4gS1tjYXRhbG9nLWluZm8ueWFtbF1cbiAgQyAtLT4gSyIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In19)

## Delivery model (GitOps)

1. **CI builds artifacts** — tests, security scans (Checkov, Trivy, Gitleaks, TFLint), container images pushed to ECR.
2. **CI opens PRs** — image tags and scaffold changes land in `platform-config`.
3. **Argo CD reconciles** — cluster state follows git; no `kubectl apply` from developer laptops.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | >= 1.23 | Build pave CLI and service-template-api |
| Node.js | >= 20 | Build pavestack-portal |
| Terraform | >= 1.6 | Manage platform-infra |
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

### Scaffold a service

```bash
make pave
./bin/pave create-service --name payments --team team-payments --database=false
```

This copies `service-template-api` → `services/payments-api` and writes `platform-config/tenants/payments/`.

### Run tests

```bash
make test
```

### Developer portal

```bash
make build-portal
# static site in pavestack-portal/out/
```

## Security defaults

- Distroless/non-root containers
- Default-deny tenant NetworkPolicies
- GitHub OIDC for CI (no long-lived AWS keys in repos)
- Terraform remote state in S3 with native lockfiles (`use_lockfile = true`, no DynamoDB)
- Structured logging and OpenTelemetry hooks in services
- KMS encryption for EKS secrets and S3 state
- Immutable ECR image tags with scan-on-push

## CI workflows

| Workflow | Component | Scanners |
|----------|-----------|----------|
| `.github/workflows/platform-infra.yml` | Terraform | Checkov, Trivy, TFLint, Gitleaks |
| `.github/workflows/platform-config.yml` | GitOps manifests | Checkov, Trivy, Gitleaks |
| `.github/workflows/service-template-api.yml` | API template | Checkov, Trivy, Gitleaks |
| `.github/workflows/pave-cli.yml` | pave CLI | Checkov, Trivy, Gitleaks |
| `.github/workflows/pavestack-portal.yml` | Portal | Checkov, Trivy, Gitleaks |
| `.github/workflows/monorepo-security.yml` | Repository-wide guardrails | Checkov, Trivy, TFLint, Gitleaks |

## Environment outputs

Terraform environments expose: `cluster_name`, `cluster_endpoint`, `vpc_id`, `private_subnet_ids`, `public_subnet_ids`, `subnet_ids`, `ecr_repository_urls`.

## Contributing

1. Fork the repository and create a feature branch
2. Make changes following existing patterns — all code must be production-aligned
3. Run `make test` and `make lint` before opening a PR
4. Ensure CI passes: security scans (Checkov, Trivy, Gitleaks, TFLint) must be green
5. Follow the GitOps constraint: CI builds artifacts and opens PRs; never run `kubectl apply` directly
