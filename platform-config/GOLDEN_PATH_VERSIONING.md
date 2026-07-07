# Golden path versioning & deprecation policy

This governs `service-template-api` (today's only golden path) and any
future golden-path templates `pave create-service` can scaffold from.

## Versioning

- Each golden-path template is versioned independently with SemVer,
  tracked as the `version`/`appVersion` fields in its Helm
  `Chart.yaml` (`service-template-api/deploy/helm/service-template-api/Chart.yaml`
  today) and mirrored in `scorecard.yaml`'s evidence trail.
- **Patch** (`0.1.x`): security patches, dependency bumps, non-breaking
  fixes to the template itself. Existing scaffolded services are not
  required to take these.
- **Minor** (`0.x.0`): additive changes — a new optional Helm value, an
  additional CI scan, a new catalog-info.yaml annotation. Backward
  compatible; existing services keep working unmodified.
- **Major** (`x.0.0`): breaking changes to the contract a scaffolded
  service relies on (e.g. changing the health-check paths, the
  `internal/app.Run` seam, or the Helm values schema). Requires a
  documented migration guide (below) before the old major version is
  deprecated.

## Deprecation lifecycle

1. **Announce** — a new major version ships alongside a
   `MIGRATION-<old>-to-<new>.md` guide in the template directory. The
   old major version is marked `deprecated: true` (comment header) but
   keeps working — deprecation is never a breaking change on its own.
2. **Warn** — `pave create-service` and the `pave-api` `/cost-estimate`
   /`/services` responses surface a warning annotation
   (`pavestack.io/template-deprecated: "true"`) on any service still on
   a deprecated major version, visible in the portal scorecard as a
   non-blocking advisory (never an auto-block — see anti-goals in
   `INTENT_SPEC.md`: the platform does not gate merges on template
   currency).
3. **Sunset** — no earlier than 2 minor releases *and* 90 days after
   the warning stage, CI support (the reusable workflow steps) for the
   deprecated major version is removed. Services on it keep running —
   Argo CD does not stop reconciling — but no further template-level
   security patches are backported to it.

## Why this exists

Golden paths are shared infrastructure with many downstream
dependents (every scaffolded service). Silent breaking changes to
`service-template-api` would either break every generated service on
next sync, or force the platform team to review every tenant's
values before merging — reintroducing the manual approval bottleneck
self-service is meant to remove. A predictable SemVer + deprecation
window lets teams upgrade on their own schedule.
