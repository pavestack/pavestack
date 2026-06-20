## 2026-06-20T15:38:44Z
Objective: Refactor `pave/internal/scaffold/scaffold.go` to accept a filesystem interface (`afero.Fs` from `github.com/spf13/afero`) instead of calling the `os` package directly. Refactor the tests to use an in-memory filesystem.

Detailed Steps:
- Add `github.com/spf13/afero` to `pave/go.mod` (run `go get github.com/spf13/afero` in the `pave/` directory, and run `go mod tidy` if necessary).
- Update the signature of `CreateService` to:
  `func CreateService(fs afero.Fs, repoRoot string, request validate.ServiceRequest) (string, error)`
- Modify all file/directory operations in `pave/internal/scaffold/scaffold.go` (including `MkdirAll`, `WriteFile`, `ReadFile`, `Rename`, `OpenFile`, `Open`, and `WalkDir`) to use the injected `afero.Fs` filesystem. Note: use `afero.WalkDir` instead of `filepath.WalkDir` to walk the virtual filesystem.
- Update `pave/internal/cli/create_service.go` to pass `afero.NewOsFs()` when calling `CreateService`.
- Refactor the tests in `pave/internal/scaffold/scaffold_test.go` to initialize and use an in-memory filesystem via `afero.NewMemMapFs()`. The mock repo root and template directory structure must be set up in the virtual/in-memory filesystem, and the tests should not write anything to the actual physical disk.
- Verify everything by running `go test -v ./internal/scaffold` in `pave/` and `make test`, `make lint`, `make fmt` at the workspace root.

Important Rules:
- All commits must follow Conventional Commits (e.g. `refactor(scaffold): introduce afero filesystem seam`).
- Verify the changes using `make test`.
