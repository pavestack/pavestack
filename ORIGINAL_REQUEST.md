# Original User Request

## Initial Request — 2026-06-20T17:22:57+02:00

<USER_REQUEST>
Implement all 5 architectural refactoring candidates for the pavestack repository: collapsing the GitOps manifest renderer, deepening the portal UI, introducing a Filesystem seam in the scaffold module, deepening the CLI interactive prompting, and consolidating the duplicated test-workspace setup.

Working directory: /Users/tengischinzorigt/Downloads/github/pavestack
Integrity mode: benchmark

## Requirements

### R1. GitOps Manifest Renderer Refactoring
Split the GitOps manifest generation in `pave/internal/gitops/gitops.go` into two distinct modules: a `TenantManifest` renderer that uses `text/template` instead of `fmt.Sprintf` with positional arguments, and a `VersionControl` module that wraps git/gh operations.

### R2. Portal UI Monolith Extraction
Deepen the portal UI by extracting the various sub-components from `pavestack-portal/src/main.tsx` (e.g., ScoreRing, StatCard, etc.) and icons into a new `app/` directory. `main.tsx` should primarily be a mount call, and `app/index.ts` should only export the `<App />` component.

### R3. Filesystem Seam for Scaffold Module
Refactor `pave/internal/scaffold/scaffold.go` to accept a filesystem interface (like `io/fs.FS` or `afero.Fs`) rather than calling the `os` package directly. Refactor the tests to use an in-memory filesystem.

### R4. CLI Interactive Prompting Seam
Update the `promptMissing()` function in `pave/internal/cli/create_service.go` to accept an `io.Reader` (e.g., via cobra's `cmd.InOrStdin()`) to make interactive prompting testable.

### R5. Consolidate Test Workspace Setup
Extract the duplicated test workspace setup logic (`setupTestWorkspace` and `setupRepoRoot`) found in the CLI and scaffold tests into a shared `internal/testutil.SetupWorkspace(t *testing.T)` function.

## Acceptance Criteria

### Verification Methods
- [ ] **Programmatic**: The provided `Makefile` targets (`make fmt`, `make lint`, `make test`) must pass for the Go code.
- [ ] **Agent-as-judge**: An agent must review the diffs to ensure the architectural goals are met without cheating (e.g., ensuring `text/template` is actually used instead of `fmt.Sprintf`, UI files are physically split, `io/fs` is used instead of `os`).

### 1. GitOps Refactoring
- [ ] `pave/internal/gitops/gitops.go` no longer uses `fmt.Sprintf` with positional arguments for template generation.
- [ ] `text/template` is used for GitOps manifest generation.
- [ ] Git/gh operations are isolated into a `VersionControl` module.

### 2. Portal UI Monolith Extraction
- [ ] Components from `pavestack-portal/src/main.tsx` are split into separate files inside a new `app/` directory.
- [ ] `main.tsx` is significantly smaller and primarily acts as a mount call.
- [ ] Portal UI tests continue to pass.

### 3. Filesystem Seam
- [ ] `pave/internal/scaffold/scaffold.go` uses an injected filesystem interface instead of direct `os` package calls.
- [ ] Scaffold tests use an in-memory filesystem.

### 4. CLI Interactive Prompting Seam
- [ ] `promptMissing()` in `pave/internal/cli/create_service.go` accepts an `io.Reader` (or uses `cmd.InOrStdin()`).

### 5. Consolidate Test Workspace Setup
- [ ] Duplicated setup logic in `create_service_test.go` and `scaffold_test.go` is removed.
- [ ] A shared `SetupWorkspace` function is used across the test files.
</USER_REQUEST>
