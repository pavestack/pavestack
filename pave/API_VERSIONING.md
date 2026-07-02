# pave-api HTTP contract versioning & deprecation policy

This governs the `pave-api` HTTP contract (`pave/api/openapi.yaml`,
served at runtime from `GET /api/v1/openapi.json`) - not the golden-path
service template's own versioning, which is covered separately by
`platform-config/GOLDEN_PATH_VERSIONING.md`.

## Versioning

- The contract is versioned by URL path prefix: `/api/v1/...`. There is
  currently only one version.
- `pave/api/openapi.yaml`'s `info.version` field tracks the contract's
  own SemVer independently of the URL version - it can bump on
  non-breaking changes (a new optional field, a new endpoint) without a
  new URL prefix.
- **Non-breaking** (bump `info.version` minor/patch, no new URL prefix):
  adding an optional request field, adding a new endpoint, adding a new
  field to a response object, loosening a validation rule. Existing
  callers (the portal, `pave` CLI users hitting the API directly) keep
  working unmodified.
- **Breaking** (requires a new URL prefix, `/api/v2/...`): removing or
  renaming a field, changing a field's type or meaning, removing an
  endpoint, tightening validation in a way that rejects previously-valid
  requests, or changing authentication/authorization requirements on an
  existing endpoint.

## Deprecation lifecycle

`pave-api` has exactly one caller today (`pavestack-portal`, built and
deployed from this same monorepo) plus whatever ad hoc scripts/tooling
call it directly. Because the caller list is small and this is an
internal platform tool (not a public API with unknown consumers), the
lifecycle is lighter-weight than the golden-path template's:

1. **Announce** - a new `/api/v2` prefix ships alongside `/api/v1`, both
   served by the same running `pave-api`. The portal is updated to call
   `/api/v2` in the same PR or a fast-following one.
2. **Sunset** - once nothing in this monorepo calls `/api/v1` anymore
   (verified by grepping `pavestack-portal/src` and `pave/internal/cli`
   for the old path), remove the `/api/v1` routes and update
   `openapi.yaml`. No fixed time window is promised, since there's no
   external SLA on this API today - if that changes (a real deployment
   with third-party callers), this document should be updated to add
   one.

## Why this exists

`pave-api` grew organically without a documented contract - `handleCreateService`/`handleDecideAccessRequest`/etc. changing shape silently would break the portal with no warning. `pave/api/openapi.yaml` makes the contract explicit and machine-readable (served live at `/api/v1/openapi.json`, so it can never drift out of sync with what's actually running the way a hand-maintained doc-only spec would). This document exists so a future breaking change has a documented path (new prefix) instead of an ad hoc "just change the response shape and hope the portal PR lands in the same deploy."
