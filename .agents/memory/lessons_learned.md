# Lessons learned

Non-obvious things discovered while doing this pass. Read before making
similar changes again.

## `default-deny` NetworkPolicy without an egress allow silently breaks DNS

`platform-config/templates/network-policy/default-deny.yaml` denies both
ingress and egress with an empty `podSelector`. That's correct
zero-trust intent, but with zero explicit allow rules it also blocks a
pod's own DNS resolution (egress UDP/TCP 53 to CoreDNS is still egress).
This wasn't caught by any existing test or CI check because nothing in
this repo actually deploys a pod against a real cluster — it would only
have surfaced the first time someone deployed a real generated service
and it couldn't resolve `service-template-api.svc.cluster.local` or any
external hostname. Lesson: a "default-deny" NetworkPolicy template always
needs a paired "but here's what every workload needs regardless" allow
list (DNS at minimum), reviewed explicitly, not assumed to be someone
else's problem later.

## GitHub Actions only executes workflows under the *root* `.github/workflows/`

`platform-infra/.github/workflows/terraform.yml` existed, had a
soft-fail Checkov step, and looked like a real gap in the CI gating
checklist — except it could never actually run; GitHub only discovers
workflow files at the repository root's `.github/workflows/`, not in a
subdirectory that happens to be named the same way. It was a stale
leftover, presumably from before `platform-infra` was folded into this
monorepo (its own note in `platform-infra/README.md`'s history suggests
platform-infra existed standalone once). Lesson: before "fixing" a CI
gating gap, check whether the workflow file in question is even reachable
by GitHub Actions' `paths:`/directory discovery rules — a `grep` for the
scanner name isn't enough, you have to know where GitHub actually looks.

## Middleware ordering matters for span-context propagation, and it's easy to get backwards

`otelhttp.NewHandler(inner, name)` injects the active span into a *new*
request object's context and calls `inner.ServeHTTP(w, newReq)` — it does
not mutate the original `*http.Request` a caller holds a reference to. A
logging middleware that wraps *outside* `otelhttp.NewHandler` and then
reads `r.Context()` after calling `next.ServeHTTP(rec, r)` is reading the
wrong `r` — the span was attached to a different request object one layer
in. The fix is `otelhttp.NewHandler(loggingMiddleware(mux), name)`
(otelhttp outermost), not the other way around. Caught this by tracing
through what object each layer actually holds, not by running against a
live collector — there wasn't one in this sandbox to catch it at runtime,
so this class of bug has to be caught by reading the middleware chain
carefully, not by "it compiled."

## `strings.Replacer` (Go) does one left-to-right pass, not sequential substitution

`pave/internal/scaffold/scaffold.go`'s `walkReplace` uses
`strings.NewReplacer(pairs...)`, which — unlike calling `.Replace()`
repeatedly — scans the original string once and, at each position, uses
the *first matching pattern in argument order* (not the longest, not by
scanning all and picking best). This means: (1) more-specific/longer
patterns must be listed before shorter substrings they contain (e.g.
`"github.com/pavestack/service-template-api"` before
`"service-template-api"`), or the shorter pattern will fire mid-match and
corrupt the output; (2) text you insert via a *replacement* is never
itself re-scanned for further matches, since the whole thing runs in one
pass over the *original* string — so injecting a `pavestack.io/team:
{{request.Team}}` block was safe from accidentally colliding with the
separate `"team-platform"` → `request.Team` pair even when
`request.Team` itself contains a string that looks like it could match
another pair. Worth remembering before adding another replacement pair to
that list — order relative to existing longer/overlapping patterns
matters, but replacement-output re-matching is never a risk.

## A test fixture schema that's more permissive than the real schema hides real validation bugs

`pave/internal/testutil.SetupWorkspace` embeds its own copy of
`service-request.schema.json` for tests to use (rather than reading the
real file), and that copy had `additionalProperties: false` — but so did
the real schema. Adding `runtime`/`exposure`/`tier` fields to
`ServiceRequest` without also adding them to *both* the real schema and
this test fixture immediately broke `TestCreateServiceCmd` with an
`additionalProperties 'runtime', 'exposure', 'tier' not allowed` error —
which is actually the *correct* thing to have happen (it caught a real
oversight before it could ship), but it's a reminder that test fixtures
which "mirror" production config need updating in lockstep with the
config they mirror, and grep-ing for the fixture's content before
assuming a schema change is "just" a production-file edit saves a
debugging cycle.

## Cost SHAs and GitHub Action pins are not something an LLM should assert from memory or a page summary

Asked `WebFetch` to read a GitHub release page and report the commit SHA a
tag points to; it returned a plausible, correctly-shaped 40-hex-character
string each time — but there's no way to verify that against the actual
upstream commit from inside this sandbox (both `git ls-remote` over HTTPS
and the GitHub REST API were blocked by the outbound egress policy).
Treating an LLM's summary of a rendered webpage as a verified SHA for a
supply-chain-security pin would be worse than not pinning tightly at all.
The honest choice was: pin by version tag, say why in a comment, and let
Dependabot (which *does* have real access to verify) open the actual
SHA-pinning PR. Don't manufacture false confidence in a security control.

## Terraform/Helm/Kustomize CLIs are fetchable via `go run`/direct download even when their package registries are blocked

`go run sigs.k8s.io/kustomize/kustomize/v5@v5.4.3` and
`go run helm.sh/helm/v3/cmd/helm@v3.16.3` both worked in this sandbox (the
Go module proxy is reachable even though `registry.terraform.io` and
`github.com` git/API access are not), and the Terraform *binary itself* is
downloadable directly from `releases.hashicorp.com` (a plain zip, not the
provider registry) even though `terraform validate`'s provider-plugin
download from `registry.terraform.io` is blocked. This made it possible to
actually render and verify the Kustomize/Helm output (catching real
templating bugs, like a Helm `{{ }}` vs. Argo-Rollouts-`{{ }}` escaping
issue in `analysistemplate.yaml`) even without full `terraform validate`.
Worth trying these fetch paths before concluding "no way to verify this
without a full toolchain install."
