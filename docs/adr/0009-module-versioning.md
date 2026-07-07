# 9. Relative-path Terraform modules until there's an external consumer

Date: 2026-07-01

## Status

Accepted

## Context

`platform-infra/envs/{dev,prod}/main.tf` instantiate every module by
relative path — `source = "../../modules/vpc"`,
`source = "../../modules/eks"`, and so on — rather than by a pinned Git
ref, a Terraform Registry address, or a private module registry with
semantic version tags. The conventional argument against relative-path
modules is that they give up independent versioning: an environment always
gets whatever is on disk at that commit, and there's no way to pin `envs/prod`
to an older module version while `envs/dev` tries a newer one.

That argument matters most when a module has consumers outside the repo
that apply on their own schedule against their own state. Today, this
repo's modules have exactly one kind of consumer: the `envs/{dev,prod}`
(and `envs/azure/*`) directories in this same repo, which — per ADR 2 — are
already committed to landing infrastructure and environment changes
atomically in one PR, reviewed and merged together. Pinning module versions
and cutting registry releases would add ceremony (a release step between
"module change" and "env picks it up") without adding safety, since the env
and the module already move in lockstep by construction.

Release tooling for a versioned split already exists in embryonic form:
`release-please-config.json` defines `platform-infra` as its own
release-please **component**, separate from the repo-root package,
producing its own version/tag on merge — the plumbing a Terraform Registry
or Git-ref-pinned consumption model would need is already half-built.

## Decision

Keep relative-path module sources (`../../modules/<name>`) for as long as
`platform-infra`'s only consumers are the environments in this repo.
Do not introduce module version pinning, a private registry, or Git-ref
sourcing preemptively.

Define the trigger for revisiting this explicitly: **the first external
consumer** of a `platform-infra` module — another repository, another
team's environment, or a customer applying these modules against their own
AWS account outside this repo's `envs/`. At that point, cut releases from
the existing `platform-infra` release-please component, publish tagged
module sources (Git ref or registry), and require environments to pin a
version rather than float on `main`.

## Consequences

- No release ceremony between a module change and an environment picking it
  up — `terraform plan`/`apply` in `envs/dev` and `envs/prod` always see
  the latest module code in the same commit, matching how ADR 2 expects
  cross-cutting changes to land.
- There is currently no way to test a module change against `envs/dev`
  before it also reaches `envs/prod`'s module source — environment-level
  promotion (ADR 8) is the only safety net; a bad module change reaches
  every environment's `plan` simultaneously, though `apply` is still
  independently gated per environment.
- Because `platform-infra` already has its own release-please component,
  turning on module version pinning later is a matter of publishing tagged
  sources and updating `source =` lines in `envs/*/main.tf` — not
  standing up new release infrastructure from scratch.
- Whoever notices the first external consumer is responsible for raising
  this ADR for revision; there's no automated check that flags "someone
  outside this repo is now depending on a module."

See also: ADR 2 (monorepo layout and the atomic-change rationale this
decision depends on), ADR 8 (per-environment apply gating as the safety net
in place of module version pinning).
