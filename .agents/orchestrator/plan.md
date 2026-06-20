# Plan — Pavestack Architectural Refactoring

We will implement the 5 refactoring candidates in a systematic manner. Each candidate corresponds to a milestone. We will run the Explorer -> Worker -> Reviewer -> Challenger -> Auditor cycle for each milestone, or spawn workers and reviewers to carry out the changes.

## Refactoring Milestones
1. **Milestone 1: GitOps Refactoring (R1)**
   - Target: `pave/internal/gitops/gitops.go`
   - Split into `TenantManifest` renderer (using `text/template`) and `VersionControl` module (wrapping git/gh CLI actions).
   - Ensure no `fmt.Sprintf` with positional arguments is used for template generation.
2. **Milestone 2: Portal UI Monolith Extraction (R2)**
   - Target: `pavestack-portal/src/main.tsx` and sub-components.
   - Extract `ScoreRing`, `StatCard`, etc. and icons into `pavestack-portal/src/app/` directory.
   - `main.tsx` should only contain the React mount call.
   - `app/index.ts` should export the `<App />` component.
3. **Milestone 3: Filesystem Seam for Scaffold (R3)**
   - Target: `pave/internal/scaffold/scaffold.go` and its tests.
   - Inject filesystem interface (`io/fs.FS` or `afero.Fs`).
   - Refactor tests to use an in-memory filesystem (e.g. `afero.NewMemMapFs`).
4. **Milestone 4: CLI Interactive Prompting Seam (R4)**
   - Target: `pave/internal/cli/create_service.go` (`promptMissing` function).
   - Modify `promptMissing` to accept an `io.Reader` (e.g. `cmd.InOrStdin()`) for testability.
5. **Milestone 5: Consolidate Test Workspace Setup (R5)**
   - Targets: `create_service_test.go` and `scaffold_test.go` in pave package.
   - Extract duplicated `setupTestWorkspace` / `setupRepoRoot` into a shared `internal/testutil.SetupWorkspace(t *testing.T)` function.

## Verification & Guardrails
- Validate that all code passes locally with `make fmt`, `make lint`, `make test`.
- All Git operations must be committed with Conventional Commits (feat, fix, chore, docs).
- No direct Kubernetes changes; everything through ArgoCD GitOps templates if applicable.
- Forensic Auditor checks to prevent cheating or dummy implementations.
- Write lessons learned to `.agents/memory/lessons_learned.md`.
