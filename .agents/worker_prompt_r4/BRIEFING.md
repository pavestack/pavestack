# BRIEFING — 2026-06-20T17:43:30+02:00

## Mission
Refactor `promptMissing()` in `pave/internal/cli/create_service.go` to accept an `io.Reader` interface for interactive testing.

## 🔒 My Identity
- Archetype: teamwork_preview_worker
- Roles: implementer, qa, specialist
- Working directory: /Users/tengischinzorigt/Downloads/github/pavestack/.agents/worker_prompt_r4
- Original parent: ca24c60e-ce31-4041-9ae9-e6db2a7b9741
- Milestone: promptMissing-refactor

## 🔒 Key Constraints
- Refactor promptMissing to accept `io.Reader`.
- Write unit tests using `strings.NewReader` in `pave/internal/cli/create_service_test.go`.
- Do not start messages with "Great", "Certainly", "Okay", or "Sure". Be direct and technical.
- Write lessons learned to `.agents/memory/lessons_learned.md` at end.
- Use Conventional Commits.

## Current Parent
- Conversation ID: ca24c60e-ce31-4041-9ae9-e6db2a7b9741
- Updated: not yet

## Task Summary
- **What to build**: Modify `promptMissing` signature and internals, update call site, add unit tests.
- **Success criteria**: Tests pass, `make test`, `make lint`, `make fmt` all succeed.
- **Interface contracts**: `func promptMissing(in io.Reader, opts *createServiceOptions, databaseProvided bool) error`
- **Code layout**: Code under `pave/internal/cli/`.

## Key Decisions Made
- Refactored `promptMissing` signature to accept `io.Reader` and modified call site to use `cmd.InOrStdin()`.
- Wrote standard table-driven Go tests under `TestPromptMissing` in `create_service_test.go` to test all prompt flows.

## Artifact Index
- `pave/internal/cli/create_service.go` - Contains refactored promptMissing function.
- `pave/internal/cli/create_service_test.go` - Contains the new unit tests verifying promptMissing.

## Change Tracker
- **Files modified**:
  - `pave/internal/cli/create_service.go` - Refactored promptMissing signature and call site.
  - `pave/internal/cli/create_service_test.go` - Added TestPromptMissing.
- **Build status**: Passed.
- **Pending issues**: None.

## Quality Status
- **Build/test result**: Pass (Verified using go test and make test).
- **Lint status**: Pass (gofmt check and go vet checks all passed).
- **Tests added/modified**: `TestPromptMissing` added to verify interactive prompting with various input patterns.

## Loaded Skills
- None loaded.
