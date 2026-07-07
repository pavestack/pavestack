# Pavestack Portal

The primary developer-facing surface for the Pavestack IDP: service catalog, scorecards, service creation, access
requests, observability snapshots, and platform docs.

## Development

```bash
npm ci
npm run dev
```

Set `VITE_API_BASE_URL` if `pave-api` (the write-path backend, built separately at `pave/cmd/pave-api`) isn't running
on the default `http://localhost:8787/api/v1`. Views that call it (Create Service, Request Access, per-service cost
estimate) show an explicit "backend unreachable" state rather than a fake success when it isn't available.

## Static export

```bash
npm run build
```

Output is written to `out/` and can be hosted on S3/CloudFront or any static file host. Note that client-side routing
requires the host to fall back to `index.html` for unknown paths (Vite's dev/preview servers do this by default).

## Data model

`catalog.json` (read-only, generated at build time by `scripts/generate-catalog.mjs`) remains the source of truth for
name/owner/lifecycle/scorecard/tier/runtime/exposure/image-tag data. It scans:

- `service-template-api/`
- `services/*/`
- the matching `platform-config/tenants/<name>/` directory, for tier/runtime/exposure/database (from `tenant.yaml`)
  and image tags + configured resource requests/limits (from `dev/values.yaml` and `prod/values.yaml`)

It never mutates cluster or Git state. Argo CD sync status/health shown in the UI is illustrative sample data, clearly
labeled as such, pending a live Argo CD/Prometheus integration — it is not read from a live cluster.

Write actions (creating a service, requesting access) call the separate `pave-api` Go backend (`pave/cmd/pave-api`)
over a REST contract — see `src/lib/api.ts` for the exact request/response shapes this portal builds against.

## What's real vs. simulated

| Area                                                                          | Status                                                          |
| ----------------------------------------------------------------------------- | --------------------------------------------------------------- |
| Catalog metadata (name, owner, lifecycle, scorecard, tier, runtime, exposure) | Real — read from committed YAML                                 |
| Deployed image tag                                                            | Real — read from committed Helm values                          |
| Configured CPU/memory requests+limits                                         | Real — read from committed Helm values                          |
| Argo CD sync status / health                                                  | Illustrative sample data, labeled in the UI                     |
| Deployment history                                                            | Not wired up; UI says so explicitly                             |
| Live resource usage (CPU/mem/RPS graphs on the service detail page)           | Empty state; no fabricated numbers                              |
| Observability golden signals / sparklines                                     | Illustrative sample data, labeled in the UI                     |
| Cost estimate                                                                 | Live call to `pave-api`; shows an offline state if unreachable  |
| Create Service / Request Access                                               | Live calls to `pave-api`; shows an offline state if unreachable |
