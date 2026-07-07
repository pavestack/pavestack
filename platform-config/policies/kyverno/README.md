# Kyverno cluster policies

Cluster-scoped admission policies, installed by
`platform-config/clusters/{dev,prod}/platform-addons.yaml` alongside the
Kyverno controller itself. These apply to every namespace/tenant
automatically — no per-tenant opt-in.

| Policy | Enforces |
|---|---|
| `disallow-privileged.yaml` | No `privileged: true`, no `allowPrivilegeEscalation: true` |
| `require-resource-limits.yaml` | Every container declares `resources.requests`/`limits` for cpu+memory (defense-in-depth alongside the `tenant-limits` LimitRange, which only supplies a default) |
| `require-labels.yaml` | Every Deployment/StatefulSet/DaemonSet carries `app.kubernetes.io/name` and `pavestack.io/team` |
| `disallow-latest-tag.yaml` | Every container image has an explicit, non-`:latest` tag |

All four run with `validationFailureAction: Enforce` — a manifest that
violates one is rejected at admission, not merely reported. `background:
true` means Kyverno's background scanner also flags any pre-existing
resource that violates a policy (visible via `kyverno-reports` /
PolicyReport objects), so drift introduced outside these four rules (e.g. a
manual `kubectl edit`) is still surfaced even though the platform's GitOps
constraint means that path shouldn't be used in practice.

## Why Kyverno over OPA Gatekeeper

Both are valid. Kyverno was chosen because its policies are native
Kubernetes YAML (no separate Rego language to learn/review), it ships a
background scanner for drift out of the box, and `service-template-api`'s
Helm chart and `pave create-service`'s tenant-manifest rendering already
produce plain YAML/labels that map directly onto Kyverno's pattern-matching
style. See `.agents/memory/decisions.md` for the full comparison.
