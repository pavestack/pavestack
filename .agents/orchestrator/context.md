# Context

## Repositories & Modules
- Workspace path: `/Users/tengischinzorigt/Downloads/github/pavestack`
- Core subsystems:
  - `pave/`: CLI tool (`go`)
  - `pavestack-portal/`: Developer Portal (`next.js`/`typescript`)
  - `platform-config/`: GitOps manifests
  - `tests/`: Integration/E2E tests (`go`)

## Architectural Constraints
- Monorepo boundaries (platform-infra vs service-template-api vs pave vs portal).
- GitOps: No direct kubectl apply.
- Conventional commits.
- Distroless/nonroot container configurations.
- Safety / Local checks (`make fmt`, `make lint`, `make test`).
