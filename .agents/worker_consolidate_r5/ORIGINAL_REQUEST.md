## 2026-06-20T15:48:11Z
You are a teamwork_preview_worker.
Your working directory path is `/Users/tengischinzorigt/Downloads/github/pavestack/.agents/worker_consolidate_r5`.
Workspace directory of the project: `/Users/tengischinzorigt/Downloads/github/pavestack`

Objective: Consolidate the duplicated test workspace setup logic (`setupTestWorkspace` in `cli/create_service_test.go` and `setupRepoRoot` in `scaffold/scaffold_test.go`) into a shared `internal/testutil.SetupWorkspace(t *testing.T, fsys ...afero.Fs)` function.

Detailed Steps:
- Create a new package/directory `pave/internal/testutil` and create `pave/internal/testutil/testutil.go`.
- Implement a shared function `SetupWorkspace(t *testing.T, fsys ...afero.Fs) (afero.Fs, string)` in the `testutil` package.
  - If a filesystem is provided as a variadic parameter, use it and populate the mock repository structure (schema files in `pave/schemas/service-request.schema.json`, template files in `service-template-api/`, and directories like `platform-config`, `pave`) under a virtual root path like `/workspace`.
  - If no filesystem is provided, create a physical temporary directory using `t.TempDir()`, wrap it in `afero.NewOsFs()`, populate it on the physical disk, and return that filesystem and path.
- Refactor `pave/internal/cli/create_service_test.go` to remove `setupTestWorkspace` and call `testutil.SetupWorkspace(t)` to get the physical temp directory.
- Refactor `pave/internal/scaffold/scaffold_test.go` to remove `setupRepoRoot` and call `testutil.SetupWorkspace(t, afero.NewMemMapFs())` to set up and run tests completely in-memory.
- Verify everything by running `go test -v ./internal/...` inside `pave/` and `make test`, `make lint`, `make fmt` at the workspace root.

Important Rules:
- All commits must follow Conventional Commits (e.g. `refactor(testutil): consolidate test workspace setup`).
- Verify the changes using `make test`.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Handoff Requirement:
Write a handoff report in `/Users/tengischinzorigt/Downloads/github/pavestack/.agents/worker_consolidate_r5/handoff.md` summarizing the refactored test utilities, files modified, and test execution results. Send a message to parent conversation ID (ca24c60e-ce31-4041-9ae9-e6db2a7b9741) once complete.
