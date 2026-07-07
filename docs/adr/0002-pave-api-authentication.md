# 2. pave-api authentication via GitHub OAuth, not a new identity provider

Date: 2026-07-02

## Status

Accepted

## Context

`pave-api` had no authentication at all. Every `/api/v1/*` endpoint,
including `POST /api/v1/services` (real scaffold + GitOps writes) and
`PATCH /api/v1/access-requests/{id}` (approve/deny), accepted requests
from anyone who could reach the process. The access-request approval
endpoint in particular only required the caller to supply an `approver`
string in the request body - nothing verified that string was the
identity of the actual caller, so any client could self-approve its own
access request.

`pave-api` is not deployed anywhere today (no `platform-config/tenants/pave-api`,
no live namespace) - it's a local/dev backend the portal points at. But
"not deployed yet" is not a reason to leave the control-plane API with no
identity model; it's exactly the point at which to add one, before a real
deployment makes retrofitting harder.

## Decision

Authenticate portal (browser) callers with a standard OAuth 2.0
Authorization Code flow against GitHub, and authorize sensitive actions
by GitHub org/team membership - not a new identity provider or a
database of users.

This fits the platform as it already exists: every other identity concept
in this repo is already a GitHub concept. Service owners are GitHub team
slugs (`catalog-info.yaml`'s `owner` field). `pave` itself shells out to
`gh` to open pull requests. `service-template-api`'s CI already trusts
GitHub's OIDC issuer for keyless Cosign signing. Introducing a second,
independent identity store (Okta/Auth0, or a homegrown user table) would
be a new source of truth for "who is allowed to do what" that has to be
kept in sync with GitHub org/team membership by hand - the same category
of problem `AGENTS.md`'s "Portal data model" section already calls out
for service metadata ("don't invent a second source of truth").

Session mechanics: after the OAuth exchange, `pave-api` issues its own
signed session cookie (HMAC-SHA256 over a JSON payload - see
`pave/internal/auth/session.go`) rather than a JWT. `pave-api` is both the
sole issuer and sole verifier of this token, so JWT's algorithm-negotiation
surface (a genuine source of real-world auth bugs - `alg: none`, RS256/HS256
confusion) buys nothing here; a minimal signed-cookie format has less
attack surface for the same guarantee.

Alternatives considered and rejected:

- **mTLS via a service mesh.** No service mesh exists anywhere in
  `platform-infra`/`platform-config` today. Standing one up solely to
  authenticate one API would be a large, disproportionate amount of new
  infrastructure for this problem.
- **Full SSO/IdP (Okta, Auth0, etc.).** Heavier than an internal platform
  tool for people who already have GitHub org membership needs, and reintroduces
  the "second source of truth for identity" problem above.
- **A static shared API key.** Weak (one leaked key compromises everyone),
  and doesn't solve the actual problem: the access-request approval
  endpoint needs to know *which* caller is approving, not just that *some*
  authorized caller is present.
- **GitHub Actions OIDC token verification for CI/automation callers.**
  Genuinely the right mechanism if/when something in CI needs to call
  `pave-api` directly (mirroring the cosign-signing trust already in this
  repo) - but nothing does today (checked: no workflow references
  `pave-api`/`PAVE_API_*`). Building JWKS-fetching/JWT-verification for a
  caller that doesn't exist yet would be speculative complexity; add it
  when a concrete caller shows up, following the same GitHub-OIDC pattern
  already established in `.github/workflows/service-template-api.yml`.

Authorization boundary chosen for this round: GET endpoints
(`/api/v1/services`, `/api/v1/cost-estimate`, `/api/v1/access-requests`
listing) stay public, since they only ever return data already committed
to this repo (`catalog-info.yaml`/`scorecard.yaml`) - gating reads
wouldn't add real confidentiality. Every mutating endpoint
(`POST /api/v1/services`, `POST /api/v1/access-requests`,
`PATCH /api/v1/access-requests/{id}`) requires a verified session.
Deciding an access request additionally requires GitHub team membership
in `PAVE_API_APPROVER_TEAM` (default `platform`) - and the approver
recorded is always the verified session identity, never a client-supplied
field, which is what closes the self-approval gap.

Fail-closed default: `config.Load()` refuses to start `pave-api` unless
either GitHub OAuth is fully configured (`PAVE_API_SESSION_SECRET`,
`PAVE_API_GITHUB_CLIENT_ID`, `PAVE_API_GITHUB_CLIENT_SECRET`) or
`PAVE_API_DISABLE_AUTH=true` is set explicitly. This mirrors the existing
`PAVE_API_DRY_RUN=true`-by-default philosophy in `AGENTS.md` ("never
silently auto-grant access") - an unconfigured OAuth app must not
silently mean "no auth," it must mean "won't start" unless a human
explicitly opts out for local dev/CI.

## Consequences

- Running `pave-api` locally now requires either a real (even a
  throwaway, personal) GitHub OAuth App, or `PAVE_API_DISABLE_AUTH=true` -
  see `pave/.env.example`. This is a small new step for local dev that
  didn't exist before.
- Team membership is cached in the session at login time (`sessionTTL` =
  12h). A team change on GitHub takes effect on the caller's next login,
  not instantly - an accepted tradeoff for a session with a bounded
  lifetime, not a persistent grant.
- If/when `pave-api` gets a real deployment (`platform-config/tenants/pave-api`),
  the GitHub OAuth App's callback URL and `PAVE_API_BASE_URL`/
  `PAVE_API_PORTAL_URL` need real values - today's defaults
  (`http://localhost:8787`/`http://localhost:5173`) are dev-only.
- CORS can no longer default to `"*"` now that cookies are in play
  (credentialed CORS requests reject a wildcard origin) - see the
  `apiserver.New` CORS-origin handling.
