# AGENTS.md

Architectural rules for anyone (human or agent) working in this repo.
Read this before `INTENT_SPEC.md` and `.agents/memory/decisions.md` (the
latter has the "why" behind each rule below, with dates).

## Monorepo — stays one repo

`platform-infra`, `platform-config`, `service-template-api`, `pave`,
`pavestack-portal`, `landing`, `services`, `brand` all live in this one
repository. Do not propose splitting them into separate repos — the
GitOps flow (CI in this repo opens PRs into `platform-config` in this same
repo) and the shared `brand/` design tokens both depend on the monorepo
layout.

## The GitOps constraint (unchanged, still absolute)

CI builds artifacts and opens pull requests. Argo CD reconciles cluster
state from Git after merge. **Nobody and nothing runs `kubectl apply`,
`helm install`, or `argocd app set` against a service/tenant namespace as
part of the normal flow** — not `pave`, not `pave-api`, not a human. The
one standing exception is bootstrapping the root `Application` once per
cluster (documented in `platform-config/README.md`), because Argo CD has
to be told where to look before it can reconcile anything.

## Golden path template versioning

`service-template-api` is SemVer'd via its Helm `Chart.yaml`. See
`platform-config/GOLDEN_PATH_VERSIONING.md` for the deprecation/migration
policy — don't ship a breaking template change without bumping major and
adding a migration note there.

## Progressive delivery

`service-template-api` deploys as an Argo Rollouts `Rollout`
(`deploy/helm/service-template-api/templates/rollout.yaml`), not a plain
`apps/v1 Deployment` — this is the golden-path default going forward, not
opt-in. It uses Argo Rollouts' basic (replica-ratio) canary, no service
mesh assumed. `rollout.analysisEnabled: true` (default) attaches an
`AnalysisTemplate` that queries Prometheus for the success-rate metric
during the canary; a failing analysis makes Argo Rollouts **automatically
abort and roll back** — that's the platform's post-deploy auto-rollback
mechanism, not a hand-scripted health-check-then-`argocd app rollback`
job. Requires the `argo-rollouts` controller and `kube-prometheus-stack`,
both installed cluster-wide by
`platform-config/clusters/{dev,prod}/platform-addons.yaml`.

## Policy enforcement — Kyverno

Cluster-wide admission policies live in `platform-config/policies/kyverno/`
and install via the `kyverno`/`kyverno-policies` Applications in
`platform-addons.yaml`. Four policies, all `validationFailureAction:
Enforce`: no privileged containers, mandatory resource requests/limits, no
`:latest` image tags, mandatory `app.kubernetes.io/name` +
`pavestack.io/team` labels (the label policy matches `Rollout` as well as
`Deployment`/`StatefulSet`/`DaemonSet` — don't forget that when adding a
policy that targets workload kinds). `platform-config/templates/` still
carries the per-tenant defaults these policies backstop: `default-deny.yaml`
+ `allow-egress-dns.yaml` NetworkPolicies, and the `tenant-quota`
ResourceQuota + `tenant-limits` LimitRange.

## Observability — three pillars, one trace ID

`service-template-api/internal/telemetry` wires both a `TracerProvider`
and a `MeterProvider` off one OTLP endpoint; `otelhttp.NewHandler` (in
`internal/server`) uses the global providers to emit a span AND the
`http.server.request.duration` histogram for every request.
`internal/logging.FromContext`/`TraceContext` attach `trace_id`/`span_id`
to every log line from the active span. Any new golden-path service (or a
second template, if one is ever added) must preserve this: same
MeterProvider + TracerProvider wiring, same trace-id-in-logs pattern — an
AI agent (or a human) doing incident triage should always be able to jump
from a log line to its trace to its metric exemplar via one ID.
`platform-config/observability/` carries the multi-window multi-burn-rate
SLO alerts (`slo-burn-rate-alerts.yaml`) and the Grafana
dashboard-as-code ConfigMap (`slo-dashboard-configmap.yaml`) — edit the
dashboard by editing that JSON and letting Argo CD sync it, never in the
Grafana UI.

## FinOps — cost-tagging convention

Every AWS resource gets `Project`, `Repository`, `Environment`,
`ManagedBy`, `CostCenter`, `Team` via one `local.tags` map per environment
root module, merged in by every child module. `Team`'s value must match
the `pavestack.io/team` Kubernetes label a tenant's namespace/workloads
carry — one team taxonomy, not two. `platform-infra`'s CI posts an
Infracost cost estimate on PRs (gated on `INFRACOST_API_KEY` being set).
`pave create-service` (and `GET /api/v1/cost-estimate` on `pave-api`)
print a static, documented cost-range estimate by tier before scaffolding
— see `pave/internal/cost`.

## Platform adoption metric

Every service scaffolded via `pave create-service` (CLI or through
`pave-api`) gets a `pavestack.io/created-via: pave-cli` annotation stamped
into its `catalog-info.yaml` by `pave/internal/scaffold`. A service without
that annotation is "manual" (hand-authored `catalog-info.yaml`, not
scaffolded). The portal's Overview stat tile and any future adoption
dashboard must compute this ratio from that annotation — don't invent a
second telemetry/analytics pipeline to answer "pave vs. manual".

## `pave-api` — safety defaults

`pave/cmd/pave-api` reuses the CLI's own `scaffold`/`gitops`/`validate`
packages to give the portal a real backend, not a mock. It defaults to
`PAVE_API_DRY_RUN=true` (skip the real `git push`/`gh pr create` step) —
an operator must explicitly set `PAVE_API_DRY_RUN=false` to let a running
instance open real PRs. Scaffold + GitOps manifest writes happen for real
either way; only the final PR step is gated. Never change that default.

## Design system — one brand, two registers

`brand/` (`mark.svg`, `mark-mono.svg`, `favicon.svg`, `tokens.css`) is the
canonical source; `landing/assets/brand/` and
`pavestack-portal/public/brand/` + `src/styles/tokens.css` are copies of
it, not forks. Don't introduce a second color palette anywhere in either
frontend — everything resolves through the `tokens.css` custom properties
(dark default, `[data-theme="light"]` override).

The two frontends intentionally use the brand differently:
- **`landing/`** is the expressive/theatrical surface — Space Grotesk at
  display scale, illustrated diagrams, animation is fine.
- **`pavestack-portal/`** is the dense/functional surface — heading scale
  caps at `text-xl` equivalent, Space Grotesk appears only in the nav
  wordmark, everywhere else is Inter (UI text) / JetBrains Mono (code,
  tabular numeric columns via `font-variant-numeric: tabular-nums`).

Don't blur this line by adding hero-scale typography inside the portal
app shell, or by making the landing page as restrained as the portal.

## Portal data model — one source of truth

The portal (and `pave-api`'s `/api/v1/services`) both read
`catalog-info.yaml` + `scorecard.yaml` directly from
`service-template-api/` and `services/*/` — that's the only source of
truth for service metadata. `pavestack-portal/scripts/generate-catalog.mjs`
(JS reader) and `pave/internal/apiserver/catalog.go` (Go reader)
intentionally duplicate the *parsing logic* rather than sharing a library
across runtimes, but must never diverge on *what fields they read* or
*what they compute* (e.g. both derive `createdVia` from the
`pavestack.io/created-via` annotation, both treat missing environment
data as absent rather than fabricating it). If you add a new field to one
reader, add it to the other.

## Action pinning

Every action in `.github/workflows/*.yml` is pinned to a commit SHA with a
version comment (e.g. `uses: owner/repo@<sha> # v7`), verified against the
real upstream tag before landing. Actions added in this round
(`gitleaks/gitleaks-action`, `anchore/sbom-action`,
`sigstore/cosign-installer`, `ossf/scorecard-action`,
`actions/upload-artifact`, `infracost/actions/setup`) are pinned by
version tag only, each with an inline comment explaining why: this
session's sandboxed environment couldn't reach `registry.terraform.io` /
`git ls-remote` over the outbound proxy to verify an exact SHA, and
shipping a fabricated SHA is worse than a mutable-but-correct tag pinned
to a real release. Dependabot (`.github/dependabot.yml`,
`github-actions` ecosystem, already scoped to `/`) will open the first
SHA-pinning PR for each of these — accept it rather than hand-picking a
SHA later without verifying it against the actual tag commit.

## Anti-goals

See `INTENT_SPEC.md`.
