# Project: Pavestack Architectural Refactoring

## Architecture
- `pave`: The self-service CLI. Contains command-line commands and gitops, scaffold, cli libraries.
- `pavestack-portal`: Next.js developer portal displaying metrics, score rings, stat cards, etc.
- `platform-config`: Reconciled by ArgoCD, contains declarative manifests for tenant namespaces.
- `tests`: End-to-end testing suite validating the full CLI service generation and template compilation flow.

## Code Layout
- `pave/cmd/pave/main.go` - Entry point for the CLI.
- `pave/internal/cli/` - Cobra commands, prompt handling, interactive input.
- `pave/internal/gitops/` - GitOps template rendering and Git operations.
- `pave/internal/scaffold/` - Microservice skeleton scaffolding from service-template-api.
- `pavestack-portal/src/main.tsx` - App entry point and mounting.
- `pavestack-portal/src/app/` - Split components (extracted from main.tsx).
- `pave/internal/testutil/` - Shared test utilities including workspace setup.

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|---|---|---|---|
| 1 | GitOps Refactoring (R1) | Split gitops.go into TenantManifest rendering and VersionControl module. Use text/template. | none | DONE |
| 2 | Portal UI Monolith Extraction (R2) | Extract ScoreRing, StatCard, etc. into app/ directory under portal. App index.ts exports App component. | none | DONE |
| 3 | Filesystem Seam for Scaffold (R3) | Refactor scaffold.go to accept filesystem interface. Use in-memory filesystem in tests. | none | DONE |
| 4 | CLI Interactive Prompting Seam (R4) | Accept io.Reader in promptMissing() to make it testable. | none | DONE |
| 5 | Consolidate Test Workspace Setup (R5) | Consolidate duplicated test setup logic to internal/testutil.SetupWorkspace. | M3, M4 | DONE |
| 6 | Deepen Validation in Pave (R1) | Extract Validator struct with afero.Fs filesystem seam in validate.go. Add tests using in-memory fs. | none | DONE |
| 7 | Extract App module in service-template-api (R2) | Extract lifecycle logic from main.go to App module in internal/app/app.go. | none | DONE |
| 8 | Replace regex YAML parsing in pavestack-portal (R3) | Replace regex-based YAML parser in generate-catalog.mjs with js-yaml library. | none | DONE |

## Interface Contracts
### Gitops Renderer ↔ VersionControl
- `TenantManifest` handles YAML generation for a tenant given custom parameters, using `text/template`.
- `VersionControl` wraps shell commands for `git` and `gh` operations (pull requests, branching, staging, committing).

### Scaffold ↔ Filesystem Seam
- Scaffold operations (directory creation, file copying, string replacements) must interact via an injected filesystem interface (`io/fs.FS` or `afero.Fs`) rather than the native `os` package.

### Validator ↔ Filesystem Seam
- Validator operations (schema reading, directory checks) must interact via an injected `afero.Fs` interface instead of the direct `os` package.

### App Lifecycle Seam
- The `App` module exposes a `Run(ctx context.Context) error` method that coordinates server startup, telemetry initialization, logging, signal handling, and graceful shutdown.

