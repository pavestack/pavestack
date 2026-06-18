# platform-config

The GitOps source of truth for all Pavestack-managed Kubernetes workloads. This directory is reconciled by **Argo CD** — any change merged to `main` is automatically applied to the target cluster.

## Constraint

This directory contains **only deployment state** — Kubernetes manifests, Helm values, and Argo CD configuration. No application source code belongs here.

## Directory structure

```
platform-config/
├── clusters/
│   ├── dev/                          # Dev cluster Argo CD config
│   │   ├── root-application.yaml     # Root App-of-Apps (apply once manually)
│   │   ├── project-pavestack.yaml    # AppProject with namespace guardrails
│   │   └── applicationset-tenants.yaml  # Auto-discovers tenants via Git generator
│   └── prod/                         # Prod cluster (same structure as dev)
├── templates/                        # Reusable base manifests for tenant bootstrap
│   ├── namespace/                    # Namespace with pavestack.io labels
│   ├── rbac/                         # Developer Role + RoleBinding
│   ├── network-policy/               # Default-deny Ingress/Egress NetworkPolicy
│   └── resource-quota/               # CPU/Memory/Pod quotas per tenant
└── tenants/
    └── {service-name}/               # Per-service tenant directory
        ├── tenant.yaml               # Metadata: namespace, Helm path, owner, database
        ├── base/kustomization.yaml   # Kustomize overlay → templates (namespace, RBAC, etc.)
        ├── dev/
        │   ├── application.yaml      # Argo CD Application for dev
        │   └── values.yaml           # Helm values override for dev
        └── prod/
            ├── application.yaml      # Argo CD Application for prod
            └── values.yaml           # Helm values override for prod
```

## How it works

### Cluster bootstrap

1. Platform team applies the **root Application** once per cluster:
   ```bash
   kubectl apply -f platform-config/clusters/dev/root-application.yaml
   ```
2. The root Application discovers `project-pavestack.yaml` and `applicationset-tenants.yaml`
3. The ApplicationSet uses a **Git directory generator** to auto-discover all tenants under `tenants/*/base`
4. Each tenant's `base/kustomization.yaml` bootstraps the namespace with RBAC, NetworkPolicy, and ResourceQuota
5. A second ApplicationSet uses a **Git file generator** on `tenants/*/tenant.yaml` to deploy the Helm chart for each service

### Tenant onboarding

New tenants are added automatically by the `pave create-service` CLI:

```bash
pave create-service --name payments --team team-payments --database=false
```

This generates:
- `tenants/payments/tenant.yaml`
- `tenants/payments/base/kustomization.yaml`
- `tenants/payments/dev/application.yaml` + `values.yaml`
- `tenants/payments/prod/application.yaml` + `values.yaml`

After merge, Argo CD reconciles the new tenant.

### Image promotion

CI workflows update `tenants/{name}/dev/values.yaml` (and `prod/values.yaml`) with the new image tag and open a PR. After merge, Argo CD rolls out the new image.

## Templates

| Template | Purpose |
|---|---|
| `namespace/` | Creates a Namespace with `pavestack.io/tenant`, `pavestack.io/team`, `pavestack.io/managed-by` labels |
| `rbac/` | Creates a `developer` Role (read-only pods, services, deployments) and binds it to the team group |
| `network-policy/` | Default-deny all Ingress and Egress — services must explicitly allow traffic via their Helm chart NetworkPolicy |
| `resource-quota/` | Limits: 20 pods, 4 CPU / 8Gi memory requests, 8 CPU / 16Gi memory limits |

## Security

- **Default-deny NetworkPolicy** in every tenant namespace
- **Least-privilege RBAC** — developers get read-only access to their namespace
- **ResourceQuota** prevents noisy-neighbor resource consumption
- **Automated sync** via Argo CD with `prune: true` and `selfHeal: true`
