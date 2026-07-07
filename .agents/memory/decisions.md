# Architecture decisions — 2026 platform engineering upgrade

Date: 2026-07-02. Scope: full audit + upgrade of `platform-infra`,
`platform-config`, `service-template-api`, `pave`, plus a new landing page
(`landing/`) and a substantially rebuilt self-service portal
(`pavestack-portal/`). This log records *why*, not just *what* — the diff
already shows what changed.

## Starting state (verified by direct code reading before any changes)

Before this pass: no OPA/Kyverno policies anywhere; no SBOM/image
signing; no gitleaks; Checkov was soft-fail (non-gating) in 3 of 7 CI
workflows; no Argo Rollouts/progressive delivery (plain `Deployment`
everywhere); Argo CD had no Notifications config (drift auto-healed
silently, never alerted); no OTel metrics pipeline (traces only, and not
correlated to logs); no LimitRange (only a namespace-wide ResourceQuota);
no default-allow-egress-DNS NetworkPolicy (default-deny had zero explicit
allows, which — combined with no LimitRange forcing per-pod sizing —
meant a tenant's own DNS resolution was actually broken by the platform's
own security posture); no cost-tagging beyond 4 generic keys, no FinOps CI
step, no cost estimate in `pave create-service`; `pave` only had
`--name`/`--team`/`--database`, no tier/runtime/exposure, no reusable
backend a portal could call; the portal was a single-page read-only
catalog viewer, dark-mode-only, no router; the landing page didn't exist;
the existing `logo.svg`/`assets/banner.svg` were a generic blue→purple
isometric-gradient mark — exactly what the brief said to avoid.
`platform-infra/.github/workflows/terraform.yml` was a dead file (GitHub
Actions only executes workflows under the *root* `.github/workflows/`,
so this nested one had never actually run) with a soft-fail Checkov step
that consequently didn't matter — deleted rather than fixed, since fixing
a workflow that can never execute would have been theater.

## Part 1 — architecture

### DevSecOps CI gating
Gitleaks (`.gitleaks.toml` + a job in `monorepo-security.yml`, repo-wide,
gating, runs on every PR regardless of path filters) was the one item from
the brief's explicit scanner list (`Trivy, Checkov, TFLint, Gitleaks,
OSSF Scorecard`) that was completely absent. Checkov's three soft-fail
instances (`platform-config.yml`, `pavestack-portal.yml`, the now-deleted
`platform-infra/.github/workflows/terraform.yml`) were switched to the
same `config_file: .checkov.yaml` gating pattern already used everywhere
else — there was no principled reason those three were softer than the
rest; it read as an oversight, not a deliberate risk acceptance. Added
`scorecard.yml` (OSSF Scorecard, uploads SARIF to code scanning). Added
SBOM generation (`anchore/sbom-action`, SPDX) + keyless Cosign signing +
attestation to `service-template-api.yml`'s image-build job, using the
job's existing GitHub OIDC token (`id-token: write` was already granted
for AWS auth) — no new secret, no signing key to manage or leak.

**Action-pinning tradeoff, explicit**: every action already in the repo is
SHA-pinned with a version comment. This sandbox's egress proxy blocks
`registry.terraform.io` and `git ls-remote` over `https://github.com/...`
(confirmed via `/root/.ccr/README.md`'s "do not retry organization policy
denials" guidance), so the four newly-added actions could not be
SHA-verified from inside this session. An earlier attempt used `WebFetch`
to have a small model read a GitHub release page and report a SHA — that
SHA is *not* independently verifiable and treating it as ground truth
would have been worse than not pinning at all (a wrong SHA is a supply-chain
correctness bug, silently). Landed with version tags instead, each with a
comment explaining why, and Dependabot (already configured for the
`github-actions` ecosystem at `/`) will open the real SHA-pinning PR. This
is a deliberate, documented exception, not an oversight.

### Kyverno over OPA/Gatekeeper
Both satisfy the brief. Kyverno was chosen because (a) its policies are
native Kubernetes YAML, no Rego to introduce as a second policy language
for a team that otherwise only writes YAML/Helm/Go, (b) it ships a
background scanner (`background: true`) that flags pre-existing drift for
free, (c) `disallow-latest-tag`/`require-resource-limits`/
`disallow-privileged` all pattern-match directly against `Pod` (Kyverno
intercepts the actual Pod object regardless of parent controller — no
autogen needed), which meant they worked against the new `Rollout`-based
golden path with zero extra config. Only `require-labels` had to
explicitly add `Rollout` to its `match.resources.kinds` (it targets the
controller object, not Pod) — caught this by reasoning through Kyverno's
match semantics before shipping, not after.

### LimitRange + egress-DNS NetworkPolicy — closing a real, working bug
`default-deny.yaml` denies all ingress *and* egress with zero explicit
allows. Without an egress allow for DNS, every tenant pod's own hostname
resolution (including in-cluster service discovery) would fail — this
wasn't a theoretical gap, it would have broken the golden path's own
generated services the first time someone actually deployed one.
`allow-egress-dns.yaml` (UDP+TCP/53, cluster-wide namespace selector) is
the one platform-wide free allow; everything else stays default-deny by
design (a service's own chart must declare its own additional egress, same
pattern `service-template-api`'s `networkpolicy.yaml` already used for
ingress). `tenant-limits` LimitRange backstops `require-resource-limits`
Kyverno policy with actual default injection (defense in depth: Kyverno
makes omission a hard failure, LimitRange means a chart that's missing
sizing still gets something sane rather than being rejected outright at
the platform layer with no recovery).

### Progressive delivery
`service-template-api`'s Helm chart deploys a `Rollout` (basic/replica-ratio
canary — no service mesh or ALB weighted-target-group integration assumed,
since neither exists in this stack yet) instead of a `Deployment`. The
canary runs a background `AnalysisTemplate` against Prometheus
(`http_server_request_duration_seconds_count`, which is exactly the metric
`internal/telemetry`'s new OTel metrics pipeline produces via
`otelhttp.NewHandler`'s semantic-convention histogram) starting at step 1;
a failed analysis makes Argo Rollouts **automatically abort and revert to
the last-known-good ReplicaSet** — this is the auto-rollback mechanism the
brief asked for, and it's the *native* Argo Rollouts mechanism rather than
a bespoke script watching health checks and shelling out to
`argocd app rollback`, which would have been reinventing what Rollouts
already does correctly. Chart version bumped `0.1.0` → `0.2.0` (pre-1.0,
so a breaking infra change in a minor bump is acceptable per
`platform-config/GOLDEN_PATH_VERSIONING.md`'s own policy) with a comment
explaining the requirement on the `argo-rollouts` controller now being
installed cluster-wide.

### Drift alerting (not just silent self-heal)
Every Application/ApplicationSet already runs `syncPolicy.automated.selfHeal:
true`, so drift *was* being corrected — just silently, with no record that
someone bypassed the GitOps PR flow via `kubectl edit`. Added a global
`argocd-notifications-cm` subscription (`platform-config/observability/
argocd-notifications-drift.yaml`) covering every Application via the
`subscriptions:` block rather than per-Application annotations — the
tenant ApplicationSets generate Applications dynamically and shouldn't need
hand-annotation per tenant for this to work. The actual webhook URL is
deliberately *not* in the repo (documented as a real secret to wire
through the normal secrets pipeline); until it's set, the triggers are
configured but undeliverable, which is a safe default, not a broken one.

### SLO burn-rate alerts + dashboards as code
Static "error rate > X%" thresholds either miss slow multi-day budget
burns or over-alert on brief blips. Implemented the standard Google SRE
multi-window multi-burn-rate pattern (`platform-config/observability/
slo-burn-rate-alerts.yaml`): a fast-burn page (14.4x over 1h, confirmed by
5m), a slow-burn ticket (6x over 6h, confirmed by 30m), and a long-window
budget-trend ticket (3d). This required a real Prometheus to exist for the
`PrometheusRule` CRD to mean anything, so `kube-prometheus-stack` was
added to `platform-addons.yaml` as a genuinely-installed cluster addon
(sized conservatively — single replica, 7d/15d retention dev/prod — not a
speculative "someone will need this eventually" over-build). The Grafana
dashboard (`slo-dashboard-configmap.yaml`) is provisioned via the sidecar
ConfigMap-label mechanism specifically so "dashboards as code" is real —
edit the JSON, Argo CD syncs it, a manual Grafana UI edit gets reverted on
next sync (documented inline so nobody's surprised by that).

### OTel metrics + trace/log correlation
Added a `MeterProvider`/`otlpmetrichttp` pipeline alongside the existing
`TracerProvider` in `internal/telemetry`, sharing one `Resource` and one
endpoint. `otelhttp.NewHandler` then emits both a span and the
`http.server.request.duration` histogram per request automatically — this
metric name is what the new `AnalysisTemplate` and `PrometheusRule` both
query, so the three pieces (canary analysis, SLO alerts, actual metrics
pipeline) are load-bearing on each other, not independently-plausible YAML
that would silently do nothing together. `internal/logging.FromContext`
attaches `trace_id`/`span_id` (from `trace.SpanContextFromContext`) to log
lines; a new `loggingMiddleware` in `internal/server` logs one line per
request with those fields. Order mattered here: `otelhttp.NewHandler` has
to be the *outer* handler so the span is already in the request context by
the time the inner logging middleware reads it — got this backwards on the
first pass (logging outside otelhttp, reading the pre-span request object)
and caught it by reasoning through what object each middleware layer
actually holds a reference to, not by testing against a live collector
(none exists in this sandbox).

### FinOps
Cost tags (`CostCenter`, `Team`) added to the one shared `local.tags` map
per environment root module (already merged into every child module via
`var.tags` — no per-module change needed, which is exactly why that
existing structure was worth preserving rather than reworking). `Team`'s
value deliberately matches the `pavestack.io/team` k8s label so AWS Cost
Explorer's cost-allocation-tag report and in-cluster attribution don't
drift into two taxonomies. Found `bootstrap/remote-state/main.tf` had
defined `local.tags` but never actually applied it to any resource — a
pre-existing dead-code bug, fixed as part of this pass since it's directly
in scope. Infracost added as an independent CI job (no AWS credentials
needed — it prices from HCL directly) gated on `INFRACOST_API_KEY` being
set, posting a PR comment. `pave/internal/cost` provides a static,
documented tier→cost-range table (not a live pricing API call — a rough
order-of-magnitude estimate, explicitly labeled as such everywhere it's
surfaced) shared by the CLI's pre-scaffold printout and `pave-api`'s
`/cost-estimate` endpoint.

### `pave` CLI + `pave-api` backend
Added `--runtime`/`--exposure`/`--tier` flags (defaulting via
`ServiceRequest.ApplyDefaults()`, called identically by the CLI and by
`pave-api` so the two entry points can never disagree about what an unset
field means). Tier drives replica count and resource sizing in the
rendered tenant `values.yaml` (`RenderValuesTiered` — added as a new
method rather than changing `RenderValues`'s signature, preserving the
existing tested contract from the prior refactoring milestones in
`PROJECT.md`). Every scaffolded `catalog-info.yaml` gets
`pavestack.io/created-via: pave-cli` plus `tier`/`runtime`/`exposure`
annotations — this is the entire "platform adoption" metric mechanism: no
new database, just an annotation the existing catalog reader already
parses. `pave-api` (`pave/cmd/pave-api`) is a genuinely new HTTP service,
not a mock — it imports and calls the exact same `scaffold`/`gitops`/
`validate` packages the CLI calls, wrapped in an async job runner so the
portal can poll status instead of blocking an HTTP request for a full
scaffold+PR cycle. Defaults to `PAVE_API_DRY_RUN=true` so a demo/public
deployment can't silently open real PRs against a real repo; scaffold and
GitOps manifest writes happen for real regardless (that's the actual
value of the endpoint) — only the final `gh pr create` step is gated.
Extracted `pave/internal/workspace` (repo-root resolution) out of
`internal/cli` so both the CLI and `pave-api` share it instead of
duplicating the directory-walk logic — small, but exactly the kind of
seam the repo's own prior refactoring milestones (`PROJECT.md`) already
established as the house style.

## Part 2/3 — landing page and portal

Built by two agents working in parallel (in `landing/` and
`pavestack-portal/` respectively) against a shared, pre-written contract:
the `brand/` design tokens, the exact `pave-api` HTTP API shape, and the
`pavestack.io/created-via` adoption-annotation convention — written down
*before* dispatching either agent specifically so their two independent
outputs would agree on data shape without needing a later reconciliation
pass. Both were instructed, and independently verified via Playwright
screenshots (read back and inspected, not just "screenshot taken"), to be
honest about what's real vs. simulated data rather than presenting
placeholder numbers as live.

**Why `landing/` is a new top-level directory instead of
`pavestack-portal/web/landing`** (as one literal reading of the brief
suggested): `pavestack-portal/` already existed as a working, tested Vite
app *at its own root* (not under a `web/` subdirectory), with CI
(`.github/workflows/pavestack-portal.yml`), a `catalog.json` generation
step, and passing tests wired to that exact path. Restructuring a working,
tested app's directory layout to match a path implied by a brief written
before the codebase was inspected was judged higher-risk than the
structural deviation is costly — same outcome (one landing surface, one
portal surface, sharing one brand system), lower blast radius. Documented
here rather than silently diverging.

**Why the old `logo.svg`/`logo.png`/`assets/banner.svg` were replaced, not
edited**: they're a stacked-diamond isometric mark in blue→purple
gradients — the literal "generic AI/SaaS" look the brief calls out by
name to avoid. A new geometric mark (`brand/mark.svg`: teal hexagon
"paver tile" + amber forward chevron) was designed instead of
incrementally adjusting the old one, verified legible at 24px/64px/200px
in both themes via a rendered Playwright screenshot before propagating it
into either frontend. The old files are left in git history/on disk (not
deleted) but are no longer referenced from README, landing, or portal.

## Known limitations, stated rather than hidden

- Checkov's default Kubernetes ruleset likely doesn't recognize the
  `Rollout` (`argoproj.io/v1alpha1`) kind the way it recognizes
  `Deployment`, so some of its built-in pod-security checks may not
  evaluate against `rollout.yaml` the way they did against the old
  `deployment.yaml`. The four Kyverno policies added this round cover the
  same ground (privileged, resource limits, labels, latest-tag) and *do*
  match `Rollout`/`Pod` correctly, but this is a real, not fully
  reconciled, gap between the two scanning layers — worth a follow-up
  Checkov custom-check if Checkov coverage parity matters more than the
  admission-time enforcement Kyverno already provides.
- `terraform validate` could not be run in this session (this sandbox's
  egress policy blocks `registry.terraform.io`); `terraform fmt -check`
  passed, and the Terraform changes made here are structurally minimal
  (additive tags/variables to already-valid files) — but they have not
  been validated against a real provider schema. Run `terraform validate`
  in a normal CI environment before merging.
- Per-tenant `PrometheusRule`/`AnalysisTemplate` templating (one instance
  per generated service rather than one reference instance for
  `service-template-api`) is a real follow-up — not built speculatively
  here per the brief's own "document, don't over-engineer" instruction for
  the AI-incident-triage path.

## Part 4 — pave-api platform-maturity pass (auth, observability, tests, OpenAPI)

A follow-up audit (industry-practice gap analysis) found `pave-api` — the
platform's actual control-plane API — hadn't kept pace with
`service-template-api`'s standards from Parts 1-3: no authentication on
any endpoint (including access-request approval, which trusted a
client-supplied `approver` string with no identity check — a
self-approval hole), `log.Printf` instead of structured/correlated
logging, no metrics/tracing, CI never ran `go vet`/`gofmt`/the portal's
Vitest suite despite `make lint`/`make fmt`/`make test` already existing,
no OpenAPI contract, no Dockerfile. Addressed in five sequenced commits:

**CI wiring first** (cheapest, unblocks everything else's feedback loop):
wired `go vet`/`gofmt -l` into `pave-cli.yml`/`service-template-api.yml`
(previously only `go test` ran), added ESLint+Prettier to the portal
(pinned `eslint-plugin-react-hooks` to the 5.x `rules-of-hooks`+
`exhaustive-deps` ruleset rather than 6.x/7.x's React-Compiler-oriented
rules, which would have forced non-mechanical effect rewrites unrelated
to adding lint tooling — revisit if the portal ever adopts React
Compiler), wired the portal's Vitest suite into CI for the first time.

**Observability/resilience parity**: added `pave/internal/logging`
(zap) and `pave/internal/telemetry` (OTel), copied from
`service-template-api`'s packages rather than extracted into a shared
module — `pave` and `service-template-api` are separate `go.work`
modules with no existing shared module, and creating one now for two
call sites was judged bigger surgery than the problem warranted; revisit
if a third Go service needs the same pattern. Added
`pave/internal/app.App.Run` (graceful shutdown on SIGINT/SIGTERM,
mirroring `service-template-api/internal/app`), panic-recovery and
request-ID middleware (previously a single panicking handler would have
crashed the whole process), `exec.CommandContext` timeouts on the
`git`/`gh` shell-outs, and bounded `JobStore` eviction (was an unbounded
in-memory map).

**Authentication — the core of this pass**: GitHub OAuth 2.0
Authorization Code flow (`pave/internal/auth`) for the portal, GitHub
org/team membership for authorization, a `pave-api`-issued HMAC-signed
session cookie (deliberately not a JWT — `pave-api` is both sole issuer
and sole verifier, so JWT's algorithm-negotiation surface buys nothing
and only adds attack surface). Full reasoning, alternatives considered
(service mesh mTLS, full SSO/IdP, static API key, GitHub Actions OIDC for
CI callers), and the resulting authorization boundary are in
`docs/adr/0002-pave-api-authentication.md` — GitHub Actions OIDC
verification was deliberately deferred (no workflow calls `pave-api`
today; building JWKS verification for a caller that doesn't exist yet
would have been speculative). `config.Load()` fails closed: refuses to
start unless OAuth is fully configured or `PAVE_API_DISABLE_AUTH=true` is
explicit, mirroring `PAVE_API_DRY_RUN`'s existing safety-default
philosophy. Also added in the same pass since they're natural companions
to "an API that's actually reachable now has sessions/cookies in play":
per-route rate limiting (`golang.org/x/time/rate`, tighter budget on
`/auth/github/login`+`callback` than general mutating endpoints),
security response headers (CSP/X-Frame-Options/HSTS), a 1 MiB request-body
cap, and a hash-chained append-only audit log
(`access-requests.audit.ndjson`) alongside the existing mutable
`access-requests.json` — `newAuditLog` replays and re-verifies the whole
chain on every `pave-api` start, refusing to start if it's been altered
out-of-band, since a tamper-evidence feature that silently continues past
a broken chain would defeat its own purpose.

**Test coverage**: added coverage tooling to all three CI-running
workflows (Go `-coverprofile` + a floor check via `go tool cover -func`;
`@vitest/coverage-v8` with a threshold config) and tests for the four
previously-untested portal route components with real user interaction
(`ServiceDetail`, `Scorecards`, `RequestAccess`, `CreateServiceWizard`) —
portal statement coverage went from ~48% to ~76%. Left `Docs.tsx`/
`Observability.tsx`/`Sparkline.tsx`/`StepTracker.tsx`/`DataTable.tsx`
lightly covered rather than writing shallow tests purely to move a
number — they're static/simulated-data display, or (`DataTable`) a
virtualized list that's awkward to exercise meaningfully in jsdom.

**OpenAPI + Dockerfile**: `pave/api/openapi.yaml` (hand-authored, all 13
endpoints) is embedded and served live at `GET /api/v1/openapi.json`
(`pave/api/openapi.go`) rather than left as a doc-only file, so it can't
drift from what's actually running. `pave/API_VERSIONING.md` documents a
lighter-weight deprecation lifecycle than the golden-path template's,
since `pave-api` currently has exactly one caller (the portal, in this
same monorepo). `pave/Dockerfile` mirrors `service-template-api`'s
multi-stage/distroless/nonroot pattern; CI now builds and Trivy-scans the
image, but deliberately does not push to a registry or open a GitOps
promotion PR — `pave-api` isn't deployed anywhere (no
`platform-config/tenants/pave-api`), so there's nothing to promote into
yet. Giving it a live namespace was explicitly treated as a separate,
un-bundled decision — see the new "Non-goals for this round specifically"
entries in `INTENT_SPEC.md`.
