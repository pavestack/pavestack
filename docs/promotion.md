# Promotion model

Pavestack promotes two independent things through `dev` → `staging` → `prod`:
application images (via `platform-config`/Argo CD) and infrastructure (via
`platform-infra`/Terraform). They use different mechanisms and different
approval gates.

## Application (image) promotion

### Dev and staging: automatic

The `service-template-api` CI workflow builds and pushes an image on every
merge to `main`, then opens a **GitOps PR** that bumps the image tag in both
`platform-config/tenants/service-template-api/dev/values.yaml` and
`platform-config/tenants/service-template-api/staging/values.yaml` in the same
commit. Once that PR is reviewed and merged, Argo CD's per-cluster
`ApplicationSet` (see `platform-config/clusters/{dev,staging}/applicationset-tenants.yaml`)
picks up the change via its default `values.yaml` file path and rolls the new
tag out automatically, with `selfHeal: true` keeping the cluster in sync.

This means staging always tracks the same image tag as dev — it exists to
validate a build against prod-shaped topology (same module set, same AZ
count, same policy set) before it goes anywhere near production traffic, not
as a separately-gated step.

### Staging to prod: manual

Prod is **not** touched by the automatic GitOps PR. Promoting a build that has
soaked in staging to prod is a separate, manually-created pull request that
edits only `platform-config/tenants/service-template-api/prod/values.yaml`
(typically bumping `image.tag` to the tag already running in staging). That
PR requires review/approval like any other change to `platform-config`; once
merged, the prod `ApplicationSet` reconciles it the same way.

This gives prod an explicit, human-approved gate that dev and staging don't
have, while keeping the underlying mechanism (Argo CD reconciling a
`values.yaml` file) identical across all three environments.

## Infrastructure promotion

Each environment's Terraform lives in its own `platform-infra/envs/<env>/`
directory with its own state file
(`platform-infra/<env>/terraform.tfstate` in the shared state bucket). There
is no shared apply — a change to a module or to an env's variables must be
applied per environment, in order:

1. **dev** — first target for any module or sizing change.
2. **staging** — same prod-shaped topology (module set, AZ count, policy
   baseline) as prod, but dev-grade instance sizing, so it validates the
   *shape* of a change cheaply before prod sees it.
3. **prod** — last.

`.github/workflows/platform-infra.yml` runs `fmt`/`validate`/`tflint`/
`checkov`/`trivy` and a `plan` across a `[dev, staging, prod]` matrix on every
PR, and posts the plan for each environment as a PR comment. The `apply` job
matrix covers all three environments as well, but each matrix entry is gated
by a GitHub **environment** (`environment: ${{ matrix.environment }}`), so
`dev`, `staging`, and `prod` can each carry their own required reviewers /
wait timers in GitHub's environment protection rules. On `push` to `main` all
three apply jobs are eligible to run (subject to their environment's
protection rules); `workflow_dispatch` lets an operator target one
environment at a time via the `environment` input
(`dev`, `staging`, or `prod`).

In practice this means: land the Terraform change, let the `dev` apply run
(or approve it), watch it, then approve `staging`, then approve `prod`. There
is no automation that promotes a plan straight from dev to prod — each
environment's apply is independently gated.

## Summary

| | dev | staging | prod |
|---|---|---|---|
| Image tag update | automatic (GitOps PR) | automatic (same GitOps PR) | manual PR |
| Infra apply | first, low-friction | prod-shaped validation | manual approval, last |
