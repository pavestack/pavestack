# Pavestack Gap-Closure Plan

**Source:** July 2026 "Deep Gap Analysis Report" (CNCF Platform Engineering Maturity Model benchmark)
**Status:** Approved plan — tracks each report gap to a concrete work item in this monorepo.

The report was written against the standalone `platform-infra` repository. Pavestack has since
become a monorepo (`pave` CLI, `platform-config` GitOps tree, `pavestack-portal`,
`service-template-api`), which already closes or substantially narrows several of the reported
gaps. This plan first reconciles the report against the current state, then sequences the
remaining work into five phases. Each phase is independently shippable and leaves CI green.

---

## 1. Reconciliation: report vs. current repo state

| # | Report gap | Current state | Verdict |
|---|------------|---------------|---------|
| 1 | Observability stack | `service-template-api` has OTel/structured-logging hooks, but no cluster backend (no Prometheus/Grafana/Loki/Tempo) | **Open** |
| 2 | Ingress / TLS / DNS | Nothing (no ALB controller, cert-manager, external-dns) | **Open** |
| 3 | Secrets management | Nothing (no ESO / Secrets Manager integration) | **Open** |
| 4 | FinOps / cost | Nothing (no Kubecost/OpenCost, budgets, Infracost) | **Open** |
| 5 | Developer self-service portal | `pave create-service` CLI + `pavestack-portal` (catalog, scorecards) + golden-path template exist | **Largely closed** — remaining: surface runtime/cost data in portal (Phase 4) |
| 6 | Policy engine | `platform-config/templates/` enforces default-deny NetworkPolicy, ResourceQuota, RBAC per tenant; no admission-control engine (Kyverno/Gatekeeper) | **Partially closed** — admission policies remain |
| 7 | Autoscaling / node optimization | Static `aws_eks_node_group` in `platform-infra/modules/eks/main.tf` | **Open** |
| 8 | Backup / DR | Nothing (no Velero, no RTO/RPO docs) | **Open** |
| 9 | OpenAPI / contract-first APIs | No `openapi.yaml` in `service-template-api` | **Open** |
| 10 | GitOps config repo | `platform-config/` exists: root Application, ApplicationSet-per-tenant, cluster projects for dev/prod | **Closed** |
| 11 | Staging environment | Only `envs/dev` and `envs/prod` | **Open** |
| 12 | VPC Flow Logs | Checkov skip `CKV2_AWS_11` still present in `modules/vpc/main.tf` | **Open** |
| 13 | CI cost estimation / drift / SBOM | `platform-infra.yml` has fmt/validate/TFLint/Checkov/Trivy/plan/apply; no Infracost, PR plan comment, scheduled drift detection, or SBOM | **Open** |
| 14 | Multi-tenancy | Namespace-per-tenant with quota/RBAC/netpol via `platform-config` templates + ApplicationSets | **Closed** (hardening continues via Gap 6) |
| 15 | Module versioning / terraform-docs | Relative-path module sources; no terraform-docs automation | **Open (low priority)** |
| 16–20 | AI-ready APIs, Crossplane, feature flags, data layer, app CI/CD golden path | App CI/CD golden path exists (`service-template-api.yml`, ECR push, GitOps PR flow → Gap 20 **closed**); rest deferred | **Deferred / partial** |

Also noted in the report: `github-oidc` used `Action = "*"`. It has since been narrowed to
service-scoped wildcards (`ec2:*`, `eks:*`, `iam:*`, …). Still broader than needed — kept as a
Phase 2 hardening item.

---

## 2. Phased plan

### Phase 1 — Observable & reachable platform (gaps 1, 2, 12)

Goal: any service scaffolded by `pave` gets metrics, logs, traces, and a TLS ingress with zero
extra developer effort.

1. **`platform-infra/modules/observability/`** — kube-prometheus-stack, Loki, Tempo, and
   Alertmanager via `helm_release` (same provider pattern as `argocd-bootstrap`).
   - Pre-provisioned Grafana dashboards (cluster, node, namespace, golden-signals per tenant).
   - Default alert rules (node pressure, pod crash-loop, Argo CD sync failure) routed to a
     configurable webhook (Slack-compatible).
   - Wire the OTel exporter endpoint into `service-template-api`'s Helm values so scaffolded
     services emit traces to Tempo out of the box.
2. **`platform-infra/modules/ingress/`** — AWS Load Balancer Controller (IRSA/pod identity),
   cert-manager (ACM or Let's Encrypt issuer), external-dns scoped to a platform Route53 zone.
   - Extend `service-template-api/deploy/helm` with an opt-in `ingress:` block so the golden
     path exposes services with TLS + DNS automatically.
3. **VPC Flow Logs** — remove the `CKV2_AWS_11` skip in `modules/vpc/main.tf`; add flow logs to
   CloudWatch Logs with KMS encryption and a retention variable (short in dev, longer in prod).
4. Register both new modules in `envs/dev` and `envs/prod` with environment-appropriate sizing;
   extend `test/infra_test.go` plan assertions to cover them.

**Acceptance:** `terraform plan` green in CI for both envs; a `pave`-scaffolded service shows
metrics/logs/traces in Grafana and is reachable over HTTPS with no manual steps.

### Phase 2 — Secure platform (gaps 3, 6-remainder, OIDC hardening)

1. **`platform-infra/modules/secrets/`** — External Secrets Operator + a `ClusterSecretStore`
   backed by AWS Secrets Manager (pod identity, least-privilege IAM path prefix per tenant,
   e.g. `pavestack/<tenant>/*`).
   - Add an `ExternalSecret` example to `platform-config/templates/` and an optional
     `secrets:` block to the service template chart.
2. **`platform-infra/modules/policy/`** — Kyverno via Helm plus baseline `ClusterPolicy` set in
   `platform-config/clusters/*`:
   - require resource requests/limits, disallow privileged pods and `:latest` tags,
     restrict images to the platform ECR registries, enforce non-root.
   - Audit mode in dev, enforce in prod, promoted after one clean week.
3. **Tighten `github-oidc` IAM** — split the single bootstrap statement into per-service
   statements with resource ARN scoping where AWS supports it (S3 state bucket, KMS keys,
   ECR repos by prefix, EKS cluster ARN); keep a documented, explicit exception list for
   actions that genuinely require `*`. Remove the `CKV_AWS_108` skip if it no longer fires.

**Acceptance:** no plaintext secret paths in the golden path; Kyverno rejects a
privileged/limit-less test pod in prod; Checkov skip count in `platform-infra` reduced to zero
or individually justified in-line.

### Phase 3 — Efficient & resilient platform (gaps 4, 7, 8, 13)

1. **Karpenter** — new `platform-infra/modules/karpenter/` (controller via Helm, node IAM role,
   interruption queue, default `NodePool` with Spot + on-demand and amd64/arm64). Shrink the
   managed node group to a minimal static "system" pool for Karpenter itself, rather than
   deleting it — safer migration, still demonstrates the pattern.
2. **`platform-infra/modules/finops/`** — OpenCost (lighter than Kubecost, CNCF), AWS Budgets
   with SNS alerts, and cost-allocation tags (`Tenant`, `CostCenter`) threaded through the
   existing `tags` variables and into `platform-config` tenant labels.
3. **`platform-infra/modules/backup/`** — Velero with S3 backup bucket (KMS, lifecycle rules),
   scheduled cluster backups, and a documented restore runbook in `docs/runbooks/dr.md` with
   explicit RTO/RPO targets per environment.
4. **CI upgrades in `.github/workflows/platform-infra.yml`:**
   - Infracost diff posted as a PR comment,
   - `terraform plan` output posted as a PR comment (updates in place),
   - scheduled (cron) plan job that opens/updates a drift issue on non-empty diff,
   - SBOM generation (syft) for `service-template-api` images in its workflow.

**Acceptance:** Karpenter provisions Spot/Graviton nodes for a test deployment; PRs show cost
and plan diffs; drift job runs nightly; a Velero restore has been exercised in dev and the
runbook reflects the real procedure.

### Phase 4 — Product-grade developer experience (gaps 9, 5-remainder, 11)

1. **Contract-first golden path** — add `openapi.yaml` (OpenAPI 3.1) to `service-template-api`,
   serve `/openapi.json` from the running service, validate the spec in
   `service-template-api.yml` CI (e.g. Redocly/Spectral lint), and have `pave create-service`
   template the spec with the service name. Render API docs in `pavestack-portal` per catalog
   entry.
2. **Portal depth** — extend `pavestack-portal` scorecards with the new signals: policy
   compliance (Kyverno reports), cost per tenant (OpenCost API export), deployment health
   (Argo CD sync status). Static-export-friendly: generated at build time by
   `scripts/generate-catalog.mjs` from committed report artifacts.
3. **Staging environment** — `platform-infra/envs/staging` (prod topology, dev sizing),
   `platform-config/clusters/staging`, and staging promotion step in the GitOps flow
   (dev → staging auto, staging → prod via PR approval). Extend the CI env matrix.

**Acceptance:** scaffolding a service yields a linted OpenAPI spec visible in the portal;
three environments exist with documented promotion flow.

### Phase 5 — Polish & maturity evidence (gaps 15, docs)

1. **ADRs** — backfill `docs/adr/` for the decisions already made and the ones this plan makes:
   monorepo layout, GitOps boundary, Karpenter adoption, OpenCost vs Kubecost, ESO vs Vault,
   Kyverno vs Gatekeeper, staging promotion model.
2. **terraform-docs** — pre-commit hook + CI check generating per-module `README.md`
   (the `vpc` module README becomes the pattern).
3. **Module versioning** — adopt release-please tags per module path (release-please is already
   configured at repo root); document a policy for pinning module refs once modules are
   consumed outside this monorepo. While envs and modules live in one repo and ship atomically,
   relative paths remain acceptable — record that as an ADR rather than adding registry
   overhead now.
4. **DORA/maturity story** — `docs/metrics.md` capturing deployment frequency and lead time
   from GitHub API, plus before/after compute cost once Karpenter has two weeks of data.

### Deferred (explicitly out of scope for now)

- **Crossplane** (gap 17): overlaps Terraform; revisit only if in-cluster self-service
  provisioning becomes a requirement.
- **Feature flags** (gap 18): no consuming service needs them yet; candidate for a future
  golden-path add-on (OpenFeature + flagd).
- **Data layer modules** (gap 19): add an RDS golden-path module when the first scaffolded
  service needs a database (`pave create-service --database=true` currently has no backend).
- **MCP/agent-ready APIs** (gap 16): follows naturally once Phase 4's `/openapi.json` ships.

---

## 3. Sequencing and dependencies

```
Phase 1 (observability, ingress, flow logs)
  └─► Phase 2 (secrets, policy)          — policies reference ingress/monitoring namespaces
        └─► Phase 3 (karpenter, finops, backup, CI) — OpenCost needs Prometheus (Phase 1)
              └─► Phase 4 (openapi, portal, staging) — scorecards consume Phase 2/3 signals
                    └─► Phase 5 (docs, versioning, metrics)
```

Ground rules for every phase:

- New cluster components are installed by Terraform only for bootstrap-tier tooling (same
  boundary as `argocd-bootstrap`); anything Argo CD can own goes in `platform-config` instead.
- Every new module ships with `variables.tf`/`outputs.tf`/`versions.tf`, plan-level coverage in
  `platform-infra/test/infra_test.go`, and registration in all environments.
- No new Checkov/TFLint skips without an in-line justification; each phase should reduce the
  existing skip count, not grow it.
- CI must stay green per phase — phases land as separate PRs (one PR per module where
  practical) so each is independently reviewable and revertible.
