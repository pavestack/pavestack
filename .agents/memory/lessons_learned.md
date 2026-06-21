## App Lifecycle Extraction (Phase 2)
- Extracted application lifecycle logic (logging, telemetry, server starting/stopping, and OS/context signals) into a reusable `App` struct in `service-template-api/internal/app/app.go`.
- Simplified `cmd/server/main.go` to instantiate and run the `App` module.
- Testing HTTP servers dynamically: By exposing `Listener net.Listener` on the `App` struct, tests can bind to an OS-allocated free port (`127.0.0.1:0`) and assign the listener to the `App` before running. This prevents port conflicts in tests, guarantees the server binds to a free port, and allows the test to immediately know the target address without complex synchronization or race-prone polling.
- Test cleanup: Using a cancellable context passed to `App.Run` allows clean and graceful teardown of the server, matching the exact SIGINT/SIGTERM lifecycle in production.

## JSON Catalog Generator Refactoring (Phase 2)
- Refactored the catalog generator to use `js-yaml` rather than custom RegExp parsing. This eliminates parsing bugs and handles nested block structures or quoted strings out-of-the-box.
- Using standard YAML loading allows accessing metadata, spec, annotations, and scorecards via straightforward object properties, making the code much simpler and more robust.

## Teamwork Quota Exhaustion Pattern
- When the teamwork multi-agent system hits quota (429) repeatedly, check whether the work was already completed before concluding failure. The cron monitoring tasks continue firing after completion and will keep hitting quota — this looks alarming but is benign. Check `.agents/orchestrator/progress.md` and `.agents/orchestrator/handoff.md` to confirm actual completion status.
- Kill the stale teamwork subagent after confirming completion to stop the quota noise.

## Branch Protection
- This repo has branch protection on `main` requiring PRs and status checks. Direct push bypasses protection. For future work, use a PR-based workflow.
