# Architecture decision records

This directory records the architecturally significant decisions made about
Pavestack, in the format described by
[Michael Nygard's original ADR post](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)
(see [ADR 1](0001-record-architecture-decisions.md)). Each record is
immutable once accepted — a changed decision gets a new ADR that supersedes
the old one, rather than an edit to history.

## Index

| # | Title | Summary |
|---|---|---|
| [1](0001-record-architecture-decisions.md) | Record architecture decisions | Why this directory exists |
| [2](0002-monorepo-layout.md) | Monorepo layout for the platform and its consumers | One repo for `pave`, `platform-infra`, `platform-config`, `service-template-api`, `pavestack-portal` |
| [3](0003-gitops-boundary.md) | Where Terraform's authority ends and Argo CD's begins | Bootstrap-tier controllers via Terraform, everything reconcilable via Argo CD |
| [4](0004-karpenter-over-cluster-autoscaler.md) | Karpenter over Cluster Autoscaler, with a static system node group | Dynamic workload capacity plus a fixed system node group for bootstrapping |
| [5](0005-opencost-over-kubecost.md) | OpenCost over Kubecost for cost allocation | CNCF, lighter weight, reuses the existing Prometheus |
| [6](0006-eso-over-vault.md) | External Secrets Operator + AWS Secrets Manager over self-hosted Vault | IRSA and path-prefix scoping instead of running Vault |
| [7](0007-kyverno-over-gatekeeper.md) | Kyverno over Gatekeeper for policy enforcement | YAML policies over Rego; audit in dev, enforce in prod |
| [8](0008-staging-promotion-model.md) | Three-environment staging promotion model | dev/staging auto image promotion; prod via approved PR |
| [9](0009-module-versioning.md) | Relative-path Terraform modules until there's an external consumer | Version pinning deferred until a module has a consumer outside this repo |
| [10](0010-contract-first-openapi.md) | Contract-first OpenAPI for golden-path services | Spec-in-repo, served at `/openapi.json`, linted in CI, carried through scaffolding |

## Adding a new ADR

1. Copy the highest-numbered file as a starting point.
2. Number it sequentially (`NNNN-slug.md`, zero-padded to four digits).
3. Use `## Status` / `## Context` / `## Decision` / `## Consequences`
   headings; start `Status` as `Proposed` and move to `Accepted` once
   agreed.
4. Add a row to the index table above.
5. If the new ADR changes or replaces an existing one, mark the old ADR's
   status as `Superseded by ADR NNNN` rather than deleting or rewriting it.
