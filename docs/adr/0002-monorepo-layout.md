# 2. Monorepo layout for the platform and its consumers

Date: 2026-07-01

## Status

Accepted

## Context

Pavestack is an Internal Developer Platform: it spans Terraform modules that
provision the cluster and its controllers (`platform-infra/`), the GitOps
tree Argo CD reconciles onto that cluster (`platform-config/`), a golden-path
service scaffold (`service-template-api/`), the CLI that generates services
from that scaffold and writes tenant manifests (`pave/`), and a read-only
developer portal built from the same repo's contents
(`pavestack-portal/`). Generated services live in `services/`.

These pieces are tightly coupled by construction, not by accident:

- `pave create-service` reads `service-template-api/` and writes directly
  into `platform-config/tenants/<name>/`. A change to the template's Helm
  chart shape (new value, renamed key) and the tenant manifests it produces
  must move together or the CLI generates broken output.
- `pavestack-portal`'s catalog is generated at build time
  (`scripts/generate-catalog.mjs`) by reading tenant and service metadata
  committed elsewhere in the repo — there is no API call, so the portal's
  build depends on those files being co-located and at a known relative
  path.
- `platform-infra` and `platform-config` are two halves of one deployment
  model (see ADR 3): a Terraform module change (e.g. a new Kyverno policy
  CRD) and the Argo CD-managed manifest that depends on it (a new
  `ClusterPolicy`) are easiest to land, review, and roll back as one
  changeset when the module set is still small and evolving quickly.

The alternative — one repo per component (`pave`, `platform-infra`,
`platform-config`, `service-template-api`, `pavestack-portal`) — is the
conventional choice for teams at larger scale, and is where this platform is
likely headed once each component has independent release cadences and
external consumers (see ADR 9 for the trigger condition already defined for
`platform-infra`).

## Decision

Ship all five components in a single repository, with directory-level
ownership boundaries enforced by convention (CODEOWNERS-style, not physical
repo boundaries) rather than by splitting into multiple repositories.

Cross-component changes (CLI scaffold + tenant manifest shape, Terraform
module + the Argo CD manifests that consume its outputs, portal catalog
generator + the metadata files it reads) land as a single atomic commit and
a single PR, reviewed and CI-gated together. `.github/workflows/` has one
workflow per component (`pave-cli.yml`, `platform-infra.yml`,
`service-template-api.yml`, etc.) plus a repo-wide security workflow, so CI
cost scales with what actually changed via path filters, not with the
number of repos.

Release tagging already anticipates a split: `release-please-config.json`
defines `.` as one package and `platform-infra` as a separately-versioned
component, so `platform-infra` can be extracted later without inventing a
new versioning scheme at that time.

## Consequences

- A developer changing the golden path (scaffold + template + tenant
  manifest shape) makes one PR instead of coordinating three. This is the
  main return on this decision.
- CI must scope jobs with path filters per component, or every PR pays the
  cost of every component's test suite. This is already the case
  (`.github/workflows/*.yml` are per-component).
- There's a single `main` branch and a single set of branch-protection
  rules across components with different risk profiles (Terraform apply vs.
  a portal static-site build). Environment-level GitHub protections (see
  ADR 8) carry the actual risk-gating, not repo boundaries.
- If `pave`, `service-template-api`, or the portal ever need external
  consumers (used by other organizations, published as a standalone CLI),
  splitting them out is a deliberate, later decision — not a default. ADR 9
  names "first external consumer" as the trigger for `platform-infra`; the
  same bar applies to any other component before it leaves the monorepo.

See also: ADR 3 (Terraform/Argo CD split within `platform-infra` and
`platform-config`), ADR 9 (module versioning and the extraction trigger).
