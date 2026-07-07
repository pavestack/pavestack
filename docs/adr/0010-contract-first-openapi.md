# 10. Contract-first OpenAPI for golden-path services

Date: 2026-07-01

## Status

Accepted

## Context

Every service scaffolded through the golden path (`service-template-api`,
generated via `pave create-service`) is an HTTP API. Two common approaches
exist for API documentation: generate an OpenAPI spec from code annotations
after the fact ("code-first"), or write the spec first and treat it as the
source of truth the implementation is checked against ("contract-first").
Code-first specs drift easily — the spec is whatever the annotations
produce, and nothing stops a handler from silently diverging from its
documented shape. Contract-first specs are what get linted, versioned, and
handed to API consumers, and can be checked in CI independently of the
implementation's internals.

Because every service starts from the same template (see ADR 2's monorepo
rationale — the CLI, the template, and the manifests it produces move
together), whatever contract-first mechanism `service-template-api` uses
gets multiplied across every generated service via `pave create-service`.

## Decision

Ship the OpenAPI spec as a committed file in the template
(`service-template-api/openapi.yaml`) and treat it as the contract:

- **Spec lives in the repo, not generated from code.** `openapi.yaml` is
  hand-maintained OpenAPI 3.1.
- **Served at runtime.** `service-template-api/openapi.go` embeds the YAML
  (`//go:embed openapi.yaml`) and exposes a `JSON()` function that converts
  it for serving back at `/openapi.json` — so the contract a consumer reads
  live from the running service is exactly the file reviewed in the PR,
  not a regenerated approximation.
- **Linted in CI.** `.github/workflows/service-template-api.yml` runs
  `npx @stoplight/spectral-cli lint openapi.yaml --ruleset .spectral.yaml`
  (the ruleset extends `spectral:oas`) after asserting the file exists
  (`test -f openapi.yaml`), so a spec that doesn't parse or violates OAS
  conventions fails CI before merge.
- **Carried through scaffolding.** `pave create-service`
  (`pave/internal/scaffold/scaffold.go`) copies `openapi.yaml` into the
  generated service via the same string-replacement mechanism used for
  `go.mod` and `catalog-info.yaml` (service name, title, and description
  substituted in place), and `verifyOpenAPISpec` fails scaffolding fast if
  the template's `openapi.yaml` is missing — the CLI treats "service has no
  contract" as a scaffolding error, not a follow-up task for the generated
  service's owner.

## Consequences

- Every generated service ships with a working, linted OpenAPI spec on day
  one — `pave create-service` cannot produce a service without one.
- The spec is only as accurate as the humans maintaining it; nothing in
  this decision checks that `openapi.yaml` matches the actual handler
  behavior at runtime (no contract-testing step is in place). Spectral
  catches malformed or non-conventional specs, not specs that lie about the
  implementation.
- `/openapi.json` is generated from the same embedded bytes Spectral lints
  in CI, so there's no separate "runtime spec" to drift from the
  repo-committed one.
- Because scaffolding is string-replacement-based (ADR 2's monorepo
  coupling between `pave`, the template, and generated services), renaming
  fields inside `openapi.yaml` beyond the templated name/title/description
  still requires editing the generated file directly — the CLI only
  substitutes what it already knows how to substitute.
- The developer portal's catalog build already takes advantage of the spec
  being a committed, well-known file: `generate-catalog.mjs`'s
  `loadApiSummary` reads each service's `openapi.yaml` directly at build
  time to summarize its API surface in the catalog, with no separate
  metadata format needed.

See also: ADR 2 (monorepo layout — why the CLI, template, and generated
services can share this mechanism atomically).
