# INTENT_SPEC.md

What Pavestack is trying to be, and — as important — what it deliberately
is not. Read alongside `AGENTS.md` (the rules) and
`.agents/memory/decisions.md` (the history of how we got here).

## What Pavestack is

An Internal Developer Platform monorepo: a platform team operates shared
infrastructure (`platform-infra`) and a golden path (`service-template-api`
+ `pave`), reconciled entirely through GitOps (`platform-config` + Argo
CD). Product engineers self-service provision standardized internal API
services — through the CLI, or through the portal's wizard calling the
same logic via `pave-api` — with zero platform-team approval step for a
standard request, zero standing cluster credentials on a developer laptop,
and full policy/security/cost enforcement applied automatically at
admission time rather than reviewed by a human.

Two developer-facing surfaces, one product:
- `landing/` — the public marketing site (open-source pitch, architecture
  explainer, "why GitOps").
- `pavestack-portal/` — the internal self-service application developers
  use daily (catalog, create-service, scorecards, observability, docs).

## Anti-goals

Explicit boundaries — things Pavestack should *not* grow into, because
each one either duplicates an existing tool worse than the original, or
reintroduces the manual bottleneck self-service is meant to remove.

- **The portal does not replace the Argo CD UI.** It summarizes sync
  status/health per service for a quick glance; it does not offer manual
  sync, rollback, resource diffing, or any cluster-mutating action. If a
  developer needs to actually operate Argo CD, they use Argo CD.
- **The portal does not replace Grafana.** The Observability view shows a
  golden-signals snapshot for orientation; deep investigation happens in
  the real Grafana instance the dashboard-as-code ConfigMap provisions
  (`platform-config/observability/`), which the portal should link to,
  not reimplement.
- **`pave`/`pave-api` do not grant cluster access.** Namespace/permission
  access always goes through the "Request access" approval workflow — no
  code path silently grants `write`/`admin` scope, ever, regardless of
  who's asking.
- **Golden-path scaffolding is not multi-runtime yet.** `runtime` is a
  field in the service-request schema so the UI/API contract doesn't need
  to change later, but only `go` is actually scaffoldable today. Don't
  present `node`/`python` as available in any UI without a real second
  template behind them.
- **The platform does not gate merges on template currency.** A service
  running an older golden-path major version gets a visible advisory (see
  `platform-config/GOLDEN_PATH_VERSIONING.md`), never a blocked PR.
  Deprecation is a conversation with a deadline, not a merge block.
- **No second source of truth for service metadata.** Both the portal and
  `pave-api` read `catalog-info.yaml`/`scorecard.yaml` directly — nobody
  should add a database, a separate YAML index, or a hand-maintained
  spreadsheet that could drift from the files that are actually reconciled
  into the cluster.
- **Not a multi-cloud / multi-cluster-vendor abstraction layer.** This is
  an AWS + EKS + Argo CD platform, on purpose. Don't add a
  provider-agnostic abstraction "in case" of GCP/Azure later — build that
  if and when it's actually needed.
- **The landing page and the portal are not the same design register.**
  See AGENTS.md "design system" — don't collapse the expressive/functional
  split for consistency's sake; the split is the point.

## Non-goals for this round specifically

Documented so a future pass doesn't assume something was overlooked
rather than deliberately deferred (see `.agents/memory/decisions.md` for
the reasoning behind each):

- Per-tenant SLO `PrometheusRule` generation (today there's one reference
  instance for `service-template-api`; templating one per tenant from
  `pave create-service` is a real follow-up, not built speculatively here).
- A live Argo CD API / Prometheus integration behind the portal's
  environment-sync and observability views — both are clearly labeled
  simulated/illustrative pending that wiring, not faked as real.
- Rasterized favicon (`.ico`) generation — the SVG favicon is used
  directly; broad enough modern-browser support that a raster fallback
  wasn't worth the build-step complexity here.
