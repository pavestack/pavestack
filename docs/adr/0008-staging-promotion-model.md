# 8. Three-environment staging promotion model

Date: 2026-07-01

## Status

Accepted

## Context

Pavestack promotes two independent things toward production: application
images (via `platform-config`/Argo CD) and infrastructure (via
`platform-infra`/Terraform). A two-environment model (dev, prod) means
every change — a new image or a new Terraform module — goes straight from a
developer's dev-shaped validation to production traffic, with no
prod-shaped rehearsal in between. That's a real gap for infrastructure
changes in particular: a change that plans cleanly against dev's smaller
topology can still behave differently at prod's module set, AZ count, or
policy baseline.

A three-environment model (dev, staging, prod) closes that gap, but adds
operational cost: another Terraform state, another Argo CD `ApplicationSet`
target, another environment's worth of infrastructure to run. That cost is
only worth it if staging is cheap to keep in sync with dev (so it doesn't
become a second thing to manually promote) while still being
prod-*shaped* enough to catch topology-specific issues before prod does.

## Decision

Run three environments — dev, staging, prod — with different promotion
mechanisms for images versus infrastructure, detailed in
[`docs/promotion.md`](../promotion.md):

**Images**: dev and staging are promoted together, automatically. The
`service-template-api` CI workflow builds an image on merge to `main` and
opens one GitOps PR that bumps `image.tag` in both
`platform-config/tenants/service-template-api/dev/values.yaml` and
`.../staging/values.yaml` in the same commit. Once merged, each cluster's
Argo CD `ApplicationSet` rolls the tag out with `selfHeal: true`. Staging
therefore always tracks the same image tag as dev — its purpose is to
validate a build against prod-shaped topology, not to serve as a second
manually-gated step. Promoting a soaked build to prod is a separate,
manually-created PR touching only `.../prod/values.yaml`.

**Infrastructure**: staging runs the same module set, AZ count, and policy
baseline as prod, but dev-grade instance sizing — cheap to run, but
representative of prod's *shape*. `.github/workflows/platform-infra.yml`
plans across a `[dev, staging, prod]` matrix on every PR and posts each
environment's plan as a PR comment; the apply matrix covers all three
environments, but each matrix entry is gated by its own GitHub
**environment** protection rules (`environment:
${{ matrix.environment }}`), so dev, staging, and prod can each carry
independent required reviewers and wait timers. There is no automation that
promotes a Terraform plan straight from dev to prod: each environment's
apply is independently triggered and gated, applied in order (dev, then
staging, then prod).

## Consequences

- Staging catches prod-shaped issues (module interactions, AZ-count-
  dependent behavior, policy-baseline violations) before they reach
  production, without doubling the manual-approval burden — only prod
  requires a human-approved PR for images.
- Because staging always tracks dev's image tag, there is no "promote dev
  to staging" step to forget or automate separately; the only manual gate
  in the image pipeline is staging → prod.
- The Terraform side has no equivalent auto-promotion: every environment's
  apply, including staging's, is independently gated by GitHub environment
  protection rules. This is intentional (infrastructure changes carry
  different risk than image bumps) but means the two halves of "promotion"
  in this platform work differently from each other — a contributor has to
  know which kind of change they're making to know what promotion looks
  like.
- Adding a fourth environment (or removing staging) means updating the
  matrix in `platform-infra.yml`, the per-cluster `ApplicationSet`s, and the
  GitOps PR workflow that bumps two `values.yaml` files at once — three
  places, not one.

See also: ADR 3 (Terraform/Argo CD ownership split that this promotion
model operates across), ADR 7 (Kyverno's own audit-to-enforce promotion,
which follows the same dev-first-then-prod philosophy but per-policy rather
than per-environment).
