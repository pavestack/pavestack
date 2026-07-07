# Reports

This directory holds point-in-time exports from platform components, committed to the
repo so the portal catalog (`pavestack-portal/`) can surface fleet-wide signals at build
time without any runtime API calls — the portal is a static export, so it can only read
files that exist in the repo at build time.

Each file is refreshed by an operator or a scheduled automation job (e.g. a GitHub
Actions workflow or a cron job run against the cluster) that executes the export command
documented below and commits the resulting JSON. Consumers of these files (see
`pavestack-portal/scripts/generate-catalog.mjs`) tolerate a missing file, a missing
per-service entry, or invalid JSON — they surface "unknown" for that signal rather than
failing the catalog build.

Layout: `reports/<kind>/<env>.json`, one file per environment (`dev`, `prod`, ...).
Matching to a service is by name: the convention used across
`platform-config/tenants/<name>` is that namespace == Argo CD app name == catalog
`metadata.name`, so every report below is keyed by that same service name.

## reports/policy/\<env>.json — Kyverno PolicyReport summary

Aggregated pass/warn/fail counts per namespace, summarized from Kyverno's
`PolicyReport` / `ClusterPolicyReport` CRDs.

Produced with:

```sh
kubectl get polr -A -o json \
  | jq '{
      generatedAt: (now | todate),
      source: "kubectl get polr -A -o json",
      namespaces: (
        [.items[] | {namespace: (.scope.namespace // .metadata.namespace), summary}]
        | group_by(.namespace)
        | map({
            key: .[0].namespace,
            value: {
              pass: (map(.summary.pass) | add),
              warn: (map(.summary.warn) | add),
              fail: (map(.summary.fail + .summary.error) | add)
            }
          })
        | from_entries
      )
    }' > reports/policy/dev.json
```

Shape:

```json
{
  "generatedAt": "<ISO 8601 timestamp>",
  "source": "<command used to produce this file>",
  "namespaces": {
    "<service-name>": { "pass": 0, "warn": 0, "fail": 0 }
  }
}
```

## reports/cost/\<env>.json — OpenCost allocation export

Monthly cost allocation per namespace, from OpenCost's `/allocation` API
(https://www.opencost.io/docs/integrations/api), normalized to a 30-day window.

Produced with:

```sh
curl -s "http://opencost.opencost.svc:9003/allocation/compute?window=30d&aggregate=namespace" \
  | jq '{
      generatedAt: (now | todate),
      source: "OpenCost /allocation API (window=30d, aggregate=namespace)",
      currency: "USD",
      namespaces: (
        .data[0] | to_entries | map({key, value: {monthlyCost: .value.totalCost}}) | from_entries
      )
    }' > reports/cost/dev.json
```

Shape:

```json
{
  "generatedAt": "<ISO 8601 timestamp>",
  "source": "<command used to produce this file>",
  "currency": "USD",
  "namespaces": {
    "<service-name>": { "monthlyCost": 0 }
  }
}
```

## reports/deployments/\<env>.json — Argo CD app health/sync export

Per-application sync/health state, from the Argo CD CLI.

Produced with:

```sh
argocd app list -o json \
  | jq '{
      generatedAt: (now | todate),
      source: "argocd app list -o json",
      apps: (
        map({
          key: .metadata.name,
          value: {
            syncStatus: .status.sync.status,
            health: .status.health.status,
            lastSyncAt: .status.operationState.finishedAt
          }
        }) | from_entries
      )
    }' > reports/deployments/dev.json
```

Shape:

```json
{
  "generatedAt": "<ISO 8601 timestamp>",
  "source": "<command used to produce this file>",
  "apps": {
    "<service-name>": {
      "syncStatus": "Synced",
      "health": "Healthy",
      "lastSyncAt": "<ISO 8601 timestamp>"
    }
  }
}
```

## Consumption

`pavestack-portal/scripts/generate-catalog.mjs` reads `reports/<kind>/<env>.json` for
each known environment (`dev`, `prod`) and looks up each service by name. A missing
file, a missing per-service entry, or invalid JSON is treated as "unknown" for that
signal only — it never fails the catalog build. See
`pavestack-portal/scripts/report-helpers.mjs` for the merge logic and
`pavestack-portal/scripts/report-helpers.test.mjs` for tests covering both the
happy path and the absence fallbacks.
