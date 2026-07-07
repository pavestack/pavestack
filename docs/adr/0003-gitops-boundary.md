# 3. Where Terraform's authority ends and Argo CD's begins

Date: 2026-07-01

## Status

Accepted

## Context

Pavestack's cluster is assembled by two tools with different execution
models: Terraform (`platform-infra/`), which plans against provider APIs
before applying, and Argo CD (`platform-config/`), which continuously
reconciles Kubernetes objects to match Git. Both tools are capable of
managing Kubernetes objects, so a line has to be drawn for every controller
and every custom resource: who owns it?

The deciding constraint is `terraform plan`. Terraform's `kubernetes_manifest`
resource requires a reachable Kubernetes API server *at plan time* to fetch
the CRD schema it validates against. In this repo's CI
(`.github/workflows/platform-infra.yml`), `plan` runs before any cluster
necessarily exists or is reachable from the runner, so any
`kubernetes_manifest` for a CRD instance breaks the plan step. This is
called out explicitly in the modules that hit it:

- `platform-infra/modules/secrets/main.tf` — "ClusterSecretStore and
  ExternalSecret resources are intentionally NOT created here:
  `kubernetes_manifest` requires a reachable API server at plan time...
  They are shipped instead as kustomize bases at
  `platform-config/templates/{cluster-secret-store,external-secret}`."
- `platform-infra/modules/ingress/main.tf` — the same reasoning for
  `ClusterIssuer` (cert-manager).

Beyond that plan-time limitation, Terraform's execution model (plan/apply,
run on a schedule or on merge) is also just a worse fit for objects that
change often and independently of infrastructure sizing — tenant
onboarding, per-environment policy modes, node pool shapes — where Argo CD's
continuous reconciliation and `selfHeal` give faster, safer convergence.

## Decision

Split ownership at the "does it need cloud IAM to exist" line:

**Terraform (`platform-infra/`) owns the bootstrap tier** — anything that
must exist, with the right IAM wiring, before Argo CD itself can run, plus
the Helm-installed controllers whose IAM roles Terraform must create
alongside them:
- the EKS cluster, node groups, VPC, ECR, GitHub OIDC (`modules/{eks,vpc,
  ecr,github-oidc}`)
- Argo CD itself (`modules/argocd-bootstrap`, a `helm_release`) — it has to
  exist before anything it would reconcile can be applied
- controllers whose pods need an IRSA role bound to specific AWS
  permissions: observability (`kube-prometheus-stack`), ingress
  (AWS Load Balancer Controller, external-dns, cert-manager), External
  Secrets Operator, Kyverno, Karpenter, Velero, OpenCost — each is a
  `helm_release` plus the matching `aws_iam_role`/policy in its own module

**Argo CD (`platform-config/`) owns everything it can reconcile without
needing to mint new cloud IAM** — tenant namespaces, policies, and CRD
instances whose controller is already running:
- tenant manifests under `tenants/<name>/` (namespace, RBAC, quotas, Helm
  values per environment)
- Kyverno `ClusterPolicy` objects and their per-environment audit/enforce
  overlays (`platform-config/policies/`, see ADR 6)
- CRD instances that depend on a controller's CRDs already being installed:
  `ClusterSecretStore` and `ExternalSecret` (ESO), `ClusterIssuer`
  (cert-manager), `NodePool`/`EC2NodeClass` (Karpenter) — all shipped as
  kustomize bases under `platform-config/templates/`

The rule of thumb: if provisioning it requires calling an AWS/Azure API
(create a role, a bucket, a load balancer), it's Terraform. If it's a
Kubernetes object whose controller is already running and reconciling from
Git, it's Argo CD.

## Consequences

- `terraform plan` stays green in CI without a live cluster, because no
  module tries to create a CRD instance via `kubernetes_manifest`.
- Every "bootstrap a new custom resource kind" controller (ESO, cert-manager,
  Karpenter) has a two-part rollout: Terraform installs the controller and
  its IAM role; a matching `platform-config/templates/<kind>` kustomize
  base ships the first CRD instance, applied after the controller's CRDs
  exist. This ordering must be respected manually when bootstrapping a new
  cluster (root Application first, then the templates that depend on
  already-installed CRDs).
- Ownership is per-object-kind, not per-controller: Kyverno the controller
  is Terraform-owned, but the `ClusterPolicy` objects it enforces are
  Argo-CD-owned. This split is intentional (policies change far more often
  than the controller) but means a contributor has to know which half of a
  feature to edit.
- If a future controller needs neither new IAM roles nor is expected to
  change frequently, it could go either way; default to Argo CD unless IAM
  wiring at install time forces Terraform.

See also: ADR 4 (Karpenter, whose `NodePool` CRDs are Argo-CD-owned per this
boundary), ADR 5 (ESO, whose `ClusterSecretStore` is Argo-CD-owned), ADR 6
(Kyverno policy promotion), ADR 8 (staging promotion model, which promotes
the Argo-CD-owned half of the system).
