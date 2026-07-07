# Baseline admission policies

Kyverno `ClusterPolicy` resources enforced cluster-wide by the [Kyverno](https://kyverno.io)
admission controller (installed by `platform-infra/modules/policy`). Argo CD applies the
`overlays/dev` or `overlays/prod` kustomization to each cluster — see
`clusters/{dev,prod}/application-policies.yaml`.

## Directory structure

```
policies/
├── baseline/        # Base ClusterPolicy definitions (kustomize base)
└── overlays/
    ├── dev/         # Patches spec.validationFailureAction to Audit
    └── prod/        # Patches spec.validationFailureAction to Enforce
```

## Policies

| Policy | Purpose |
|---|---|
| `require-requests-limits` | Every container must set CPU/memory requests and limits. |
| `disallow-privileged` | Denies privileged containers, `allowPrivilegeEscalation`, and `hostPID`/`hostIPC`/`hostNetwork`. |
| `disallow-latest-tag` | Requires an explicit image tag and forbids the mutable `:latest` tag. |
| `restrict-registries` | Images must come from an approved platform registry (ECR). The base policy ships with a `REPLACE_REGISTRY` placeholder — replace it with the account's ECR host (e.g. `123456789012.dkr.ecr.us-east-1.amazonaws.com/pavestack/*`) before applying; add more `anyPattern` entries for additional approved registries. |
| `require-run-as-nonroot` | Requires `runAsNonRoot: true` at the Pod or container `securityContext`. |

All policies exclude the platform's own system namespaces (`kube-system`, `kyverno`, `argocd`,
`observability`, `cert-manager`, `external-secrets`, `external-dns`) so only tenant workloads are
validated.

## Audit-in-dev → enforce-in-prod promotion

Every policy ships in the base with `validationFailureAction: Audit`. The overlays make the
per-environment behavior explicit:

- **dev** (`overlays/dev`): `Audit` — violations are recorded as `PolicyReport`/`ClusterPolicyReport`
  entries but nothing is blocked. This is where new or changed policies are proven out against real
  tenant workloads.
- **prod** (`overlays/prod`): `Enforce` — non-conformant Pods are rejected at admission.

Per the platform's rollout plan, a policy (or a change to one) is promoted from Audit to Enforce
only **after it has run clean in dev for one full week** (zero policy violations reported). To
promote:

1. Confirm a week of clean `PolicyReport`s for the policy in dev.
2. Flip the corresponding patch in `overlays/prod/kustomization.yaml` (it should already say
   `Enforce`; the check is that the base/dev behavior has been clean, not the prod overlay itself).
3. If this is a *new* policy, keep it `Audit` in prod for its own one-week soak before flipping.

Because Argo CD auto-syncs with `selfHeal: true`, updates to these files roll out automatically
once merged to `main`.
