# 7. Kyverno over Gatekeeper for policy enforcement

Date: 2026-07-01

## Status

Accepted

## Context

Pavestack needs cluster-wide admission policy (deny privileged pods, require
resource requests/limits, disallow `:latest` tags, restrict image
registries, require non-root containers) so tenant workloads meet a
baseline regardless of what individual teams write into their Helm charts.
The two dominant Kubernetes policy engines are OPA Gatekeeper, which
expresses policy in Rego (a general-purpose logic language with its own
learning curve), and Kyverno, which expresses policy as native Kubernetes
YAML resources (`ClusterPolicy`) using pattern matching, overlays, and
generators — no separate policy language to learn.

Pavestack's baseline policies (`platform-config/policies/baseline/`:
`disallow-privileged.yaml`, `disallow-latest-tag.yaml`,
`require-run-as-nonroot.yaml`, `require-requests-limits.yaml`,
`restrict-registries.yaml`) are straightforward validation rules — deny or
audit based on a pattern match against a Pod spec. This is exactly Kyverno's
sweet spot: rules are readable by anyone who can read a Kubernetes manifest,
without requiring Rego expertise to write or review them. Since the
policies live in `platform-config` and are reviewed like any other
Kubernetes manifest (see ADR 3), keeping them in the same YAML idiom as
everything else around them lowers the bar for platform and tenant teams
alike to read and propose changes.

Every policy also needs a rollout mode: a newly-added rule must not
instantly start rejecting existing workloads across every environment.

## Decision

Deploy Kyverno (`platform-infra/modules/policy`, `helm_release "kyverno"`
from `https://kyverno.github.io/kyverno`) as the admission controller, and
author baseline policies as native `ClusterPolicy` YAML in
`platform-config/policies/baseline/`.

Promote each policy from `Audit` to `Enforce` per environment via Kustomize
overlay patches, not by editing the baseline:
- `platform-config/policies/overlays/dev/kustomization.yaml` patches every
  baseline `ClusterPolicy`'s `spec.validationFailureAction` to `Audit` —
  violations are recorded as `PolicyReport`s, nothing is blocked.
- `platform-config/policies/overlays/prod/kustomization.yaml` patches the
  same policies to `Enforce` — non-conformant Pods are rejected at
  admission.
- The overlay comments state the promotion rule explicitly: "Only flip a
  policy here once it has run clean (zero violations) in the dev overlay
  for one full week."

Baseline rules exclude system namespaces (`kube-system`, `kyverno`,
`argocd`) from evaluation, since those workloads are platform-managed and
not subject to the same tenant-facing guardrails.

## Consequences

- Policy authors write and review Kubernetes-native YAML, not Rego — no
  extra language for platform or tenant engineers to learn, consistent with
  the rest of `platform-config` being kustomize/YAML end to end.
- The dev/prod split via Kustomize overlay means a policy's enforcement
  mode is a two-line JSON patch, not a separate policy file — but it also
  means the same `ClusterPolicy` name must exist identically in
  `baseline/` for every environment; a policy that should behave
  differently in *content* (not just enforcement mode) per environment
  doesn't fit this pattern cleanly.
- New policies default to `Audit` in every environment until proven clean
  for a week in dev, which trades faster rollout for a lower risk of an
  overly-strict new rule blocking legitimate deployments.
- Kyverno itself, like the rest of the bootstrap-tier controllers, is
  Terraform-installed; the `ClusterPolicy` objects it enforces are
  Argo-CD-owned and can be edited without a Terraform apply (ADR 3).
- If a future policy needs Rego's expressiveness (e.g. cross-resource
  aggregate checks Kyverno's pattern matching can't express), that's a gap
  this decision doesn't cover and would need to be revisited per-policy
  rather than by replacing Kyverno wholesale.

See also: ADR 3 (Terraform/Argo CD ownership boundary), ADR 4 (Karpenter,
whose controller namespace is excluded from baseline policy the same way).
