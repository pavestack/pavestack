# Disaster recovery runbook

This runbook covers backup and restore of Pavestack EKS clusters using
[Velero](https://velero.io/), provisioned by `platform-infra/modules/backup`.

## Status

**The procedures below have not yet been exercised against a live cluster.**
The first quarterly drill (see [Quarterly drill](#quarterly-drill)) is what
validates this runbook. Until that drill has run at least once, treat every
command here as best-effort and re-verify flags/output against the installed
Velero CLI version before relying on it during a real incident.

## RTO / RPO targets

| Environment | RTO (time to restore service) | RPO (max acceptable data loss) |
| ----------- | ------------------------------ | ------------------------------- |
| dev         | 8 hours                         | 24 hours                        |
| prod        | 2 hours                         | 24 hours                        |

RPO is derived from the default `daily-cluster` Velero schedule
(`var.backup_schedule`, default `0 3 * * *` — once per day). It can be
tightened per environment by overriding `backup_schedule` in `envs/<env>` to
run more frequently; RPO is always approximately one schedule interval.

## What Velero covers

- Kubernetes API objects (Deployments, Services, ConfigMaps, Secrets, CRDs,
  etc.) in the namespaces included by the backup (`includedNamespaces: ["*"]`
  by default).
- EBS volume snapshots for PersistentVolumes backed by the `aws` volume
  snapshot location, taken at backup time (`snapshotVolumes: true`).

## What Velero does not cover

- **Terraform state.** Cluster and networking infrastructure (VPC, EKS
  control plane, IAM, the backup module itself) is defined in
  `platform-infra` and recreated via `terraform apply`, not via Velero. State
  lives in the versioned, KMS-encrypted S3 bucket provisioned by
  `platform-infra/bootstrap/remote-state` — that bucket is its own backup
  (versioning) and is out of scope for Velero.
- **ECR image repositories.** Container images are not backed up by Velero.
  Recovery relies on CI re-pushing images from source, or on ECR's own
  durability; there is currently no cross-region ECR replication.
- **RDS or other managed data stores.** Pavestack does not yet provision RDS
  or any managed database. If/when one is added, it will need its own
  backup/restore story (e.g. automated snapshots) — this runbook does not
  cover it.

## GitOps caveat

Argo CD (via `platform-config`) is the source of truth for most workload
manifests and will recreate that declarative state on its own once a cluster
rejoins the `platform-config` tree — you generally do not need Velero to
restore Deployments, Services, or other GitOps-managed objects that exist in
`platform-config`.

Velero remains necessary for:

- PersistentVolume contents (via EBS snapshots) — GitOps has no opinion on
  data inside a volume.
- Secrets and other runtime state created out-of-band (by controllers,
  operators, or manual `kubectl` actions) rather than committed to
  `platform-config`.
- Any other out-of-band cluster state that GitOps reconciliation would not
  recreate.

In practice: let Argo CD resync first, then use Velero to restore the data
and out-of-band state Argo CD cannot.

## Restore procedures

### Full-cluster restore

1. Recreate the underlying infrastructure for the target environment:

   ```
   cd platform-infra/envs/<env>
   terraform apply
   ```

   This recreates the VPC, EKS cluster, IAM roles, ECR repos, and the backup
   module (S3 bucket, KMS key, Velero IRSA role) — the Velero backup bucket
   itself is expected to already exist and be intact, since it is not
   deleted by a cluster teardown.

2. Let Argo CD bootstrap and resync `platform-config` against the new
   cluster so GitOps-managed workloads come back first.

3. Restore the most recent Velero backup:

   ```
   velero restore create <env>-full-restore-$(date +%Y%m%d) \
     --from-backup <backup-name>
   ```

   List available backups first with `velero backup get` if `<backup-name>`
   is not already known (the daily schedule produces backups named
   `daily-cluster-<timestamp>`).

4. Monitor restore progress:

   ```
   velero restore describe <env>-full-restore-<date> --details
   velero restore logs <env>-full-restore-<date>
   ```

### Namespace-level restore

To restore a single namespace (e.g. after an accidental deletion) without
touching the rest of the cluster:

```
velero restore create <namespace>-restore-$(date +%Y%m%d) \
  --from-backup <backup-name> \
  --include-namespaces <namespace>
```

### Post-restore verification

1. Confirm Argo CD reports all Applications `Synced`/`Healthy`:

   ```
   argocd app list
   argocd app get <app-name>
   ```

2. Confirm restored PersistentVolumeClaims are bound and pods are running:

   ```
   kubectl get pvc -A
   kubectl get pods -A --field-selector=status.phase!=Running
   ```

3. Run environment smoke checks against `service-template-api`-derived
   tenant services (health endpoint checks, basic ingress reachability) to
   confirm the restore is functionally complete, not just object-complete.

## Quarterly drill

Every quarter, exercise the restore path without touching production data by
restoring into a scratch namespace using Velero's namespace mapping:

```
velero restore create dr-drill-$(date +%Y%m%d) \
  --from-backup <backup-name> \
  --namespace-mappings <source-namespace>:dr-drill-<source-namespace>
```

Verify the restored objects and any restored PVC data in the scratch
namespace, record the wall-clock time taken against the RTO targets above,
then tear down the scratch namespace. Record drill results (date, backup
used, duration, any gaps found) and update this runbook if the actual
procedure diverges from what's written here.
