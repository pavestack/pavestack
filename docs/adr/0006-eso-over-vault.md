# 6. External Secrets Operator + AWS Secrets Manager over self-hosted Vault

Date: 2026-07-01

## Status

Accepted

## Context

Tenant workloads need application secrets (database credentials, API keys)
delivered into Kubernetes without committing them to Git or to
`platform-config`. The two common approaches are a self-hosted HashiCorp
Vault cluster (with its own unseal/HA/backup operational burden) or a
managed cloud secret store synced into the cluster by a controller.

Running Vault means operating another stateful service with its own
availability, unsealing, and backup story — real weight for a platform
whose whole premise is to reduce operational burden on tenant teams. AWS
Secrets Manager is already managed infrastructure the platform depends on
elsewhere (this is an AWS-hosted EKS platform), and the External Secrets
Operator (ESO) is a lightweight controller that syncs secrets from a cloud
provider (or Vault, or others) into native Kubernetes `Secret` objects — it
does not itself hold secret material.

Multi-tenant isolation still has to be solved regardless of backing store:
one IAM identity with blanket `secretsmanager:GetSecretValue` access would
let any tenant's `ExternalSecret` read any other tenant's secrets.

## Decision

Deploy External Secrets Operator (`platform-infra/modules/secrets`) against
AWS Secrets Manager as the backing store, using IRSA (IAM Roles for Service
Accounts) for the controller's AWS credentials rather than static keys, and
scope its IAM policy to a path prefix rather than granting account-wide
Secrets Manager access:

```
Resource = "arn:aws:secretsmanager:${var.region}:${account_id}:secret:${var.secret_path_prefix}/*"
```

The module's comment is explicit about the shape of this scoping: "ListSecrets
(which cannot be resource-scoped) is [omitted]... least-privilege Secrets
Manager read access... scoped to `${var.secret_path_prefix}/*`." Isolation
between tenants is therefore enforced by naming convention
(`secret_path_prefix` per environment/tenant) plus IAM resource scoping, not
by network boundaries or separate Vault namespaces.

Per ADR 3, the `ClusterSecretStore` and `ExternalSecret` CRD instances that
point tenant namespaces at their slice of Secrets Manager are not created by
Terraform (`kubernetes_manifest` cannot plan against a CRD schema before the
cluster and ESO exist) — they ship as Argo CD-reconciled kustomize bases at
`platform-config/templates/{cluster-secret-store,external-secret}`.

## Consequences

- No Vault cluster to run, unseal, or back up; secret durability rides on
  AWS Secrets Manager's own availability and backup guarantees.
- The controller's AWS credentials are scoped via IRSA to one path prefix,
  so a compromised ESO controller (or a misconfigured `ExternalSecret` in
  one tenant's namespace) cannot read another tenant's secrets by IAM
  policy alone — but `ListSecrets` cannot be resource-scoped, so
  correctness for enumeration-style access still depends on the controller
  behaving as intended, not solely on IAM.
- Tenants gain a dependency on AWS Secrets Manager pricing (per-secret,
  per-API-call) rather than a self-hosted store with no per-secret cost.
- Adding a new tenant means provisioning both an AWS Secrets Manager entry
  under the right path prefix and the matching `ExternalSecret`/
  `ClusterSecretStore` manifest — two systems to keep in sync, mitigated by
  the path-prefix convention being the same shape across environments.
- If Pavestack ever needs a secrets backend outside AWS (e.g. the Azure
  variant under `platform-infra/modules/azure/`), ESO's pluggable provider
  model means the controller doesn't need to change — only the backing
  `ClusterSecretStore` provider config does.

See also: ADR 3 (Terraform/Argo CD ownership boundary for the
`ClusterSecretStore` CRD instance).
