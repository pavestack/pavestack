## 2026-06-20T15:43:30Z
You are a teamwork_preview_worker.
Your working directory path is `/Users/tengischinzorigt/Downloads/github/pavestack/.agents/worker_prompt_r4`.
Workspace directory of the project: `/Users/tengischinzorigt/Downloads/github/pavestack`

Objective: Refactor `promptMissing()` in `pave/internal/cli/create_service.go` to accept an `io.Reader` interface to make interactive prompting testable.

Detailed Steps:
- Modify `promptMissing` signature in `pave/internal/cli/create_service.go` to:
  `func promptMissing(in io.Reader, opts *createServiceOptions, databaseProvided bool) error`
- Inside `promptMissing()`, replace `bufio.NewReader(os.Stdin)` with `bufio.NewReader(in)`.
- Update the call site inside `newCreateServiceCmd()` to pass `cmd.InOrStdin()` when invoking `promptMissing`.
- Write unit tests inside `pave/internal/cli/create_service_test.go` (e.g., `TestPromptMissing`) that exercise `promptMissing` with mocked user input using `strings.NewReader("...\n")` to verify that service name, team, and database flag prompt correctly.
- Verify everything by running `go test -v ./internal/cli` inside `pave/` and `make test`, `make lint`, `make fmt` at the workspace root.

Important Rules:
- All commits must follow Conventional Commits (e.g. `refactor(cli): inject io.Reader into promptMissing`).
- Verify the changes using `make test`.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Handoff Requirement:
Write a handoff report in `/Users/tengischinzorigt/Downloads/github/pavestack/.agents/worker_prompt_r4/handoff.md` summarizing the files changed, new test functions, and test execution results. Send a message to parent conversation ID (ca24c60e-ce31-4041-9ae9-e6db2a7b9741) once complete.
