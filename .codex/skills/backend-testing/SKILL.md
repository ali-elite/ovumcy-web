---
name: backend-testing
description: Plan and run focused Go backend tests for ovumcy changes, interpret failures, and summarize readiness for commit and further frontend/browser testing.
---

## Purpose

Help the user validate ovumcy backend changes by:
- mapping code changes in Go packages to concrete test suites,
- proposing focused `go test` commands (and full `go test ./...` when appropriate),
- interpreting test output and suggesting next steps or fixes,
- preparing a concise status summary for commit and any follow-up frontend or browser-based testing skills.

## Inputs

- A short description of what was changed (for example from `feature-change` or a manual summary):
  - which Go packages, services, and API handlers were touched (if any),
  - whether any `internal/db` or migration code was changed,
  - whether the change is security/privacy-sensitive (auth, permissions, export, health-related logic),
  - whether the change also includes frontend-only edits under `web/` (these will be handled by separate frontend/browser testing skills).

## Workflow

1. Clarify scope of changes
   - If the described change only touches frontend assets under `web/` and no Go code:
     - state that backend testing is optional for this change;
     - suggest using frontend and browser-e2e testing skills instead, unless the user explicitly wants backend regression tests.
   - Otherwise, ask (or read from the previous skill) which Go packages and files were modified:
     - services under `internal/services/...`,
     - API handlers under `internal/api/...`,
     - any code in `internal/db`, `internal/security`, `internal/i18n`, or migrations.
   - Mark the change as security/privacy-critical if it touches auth, access control, export, or health-related domain logic.

2. Select appropriate test commands

   - Propose a small ordered list of `go test` commands, from most focused to broader, for example:
     - targeted packages for the affected services (e.g. `go test ./internal/services/auth/...`);
     - targeted packages for affected API handlers (e.g. `go test ./internal/api/auth/...`);
     - `go test ./...` if the change is non-trivial or spans multiple domains.
   - For each command, briefly explain what it validates.
   - Ask the user which commands are acceptable in the current environment and which ones to run first.

3. Run and interpret tests (interactive)

   - For each approved command:
     - run the command and capture output;
     - if tests pass:
       - confirm success and move to the next command.
     - if tests fail:
       - summarize failing tests (package, test name, and main error message);
       - classify failures as:
         - expected regressions (exposing a real bug in the new behavior), or
         - unintended breakages.
       - propose next steps:
         - which part of the recent change to inspect,
         - whether to adjust tests or code,
         - whether to temporarily block commit and return to the implementation skill (e.g. `feature-change`) for fixes.
   - Do not proceed to broader commands (`go test ./...`) if narrow tests are already failing, unless the user explicitly asks to continue.

4. Security- and migration-sensitive checks

   - If the change touches:
     - auth, sessions, or cookies;
     - permissions/roles (owner/partner);
     - export or health-related data;
     - migrations or `internal/db`:
       - explicitly call out that additional review is needed;
       - suggest adding or tightening unit tests around:
         - unauthorized/forbidden access,
         - data leakage (no PII in responses or logs),
         - migration invariants (no data loss, correct schema states).
   - When owner-only API routes are added or expanded, include explicit `401/403` regression checks in addition to browser role coverage so transport authorization contracts stay pinned at the backend layer.
   - For destructive settings routes that erase health data, include backend regressions for missing CSRF, missing current password, invalid current password, and the successful confirmed path in the same change.
   - When an HTML flow omits a previously stored optional field, add focused backend regressions that prove create uses the expected default and update preserves the existing stored value.
   - Ask whether the user wants to add such tests now or track them as `TODO(debt/...)` with a clear reason.

5. Handoff summary for commit and other testing skills

   - Produce a short summary including:
     - which `go test` commands were run,
     - which passed and which (if any) failed,
     - what fixes were applied or are still pending.
   - If all planned commands are green:
     - state that backend tests are green for the current change and it is ready for:
       - the `commit` skill to prepare a commit,
       - frontend-testing and browser-e2e skills if the change also affects UI.
   - If there are remaining failures:
     - clearly state that the change is not ready to commit and should go back to the implementation skill with references to failing tests.

## Constraints

- Always respect project `AGENTS.md`:
  - do not propose running migrations or destructive commands as part of testing;
  - keep test scope as focused as possible before suggesting `go test ./...`.
- Never modify production code or tests in this skill without an explicit, user-approved change plan.
- Do not run commands until the user has approved the test scope or directly asked for the command run.
