---
name: commit
description: Prepare a clean, well-scoped commit for ovumcy, based on recent changes and tests, without introducing hidden technical debt.
---

## Purpose

Help the user turn the current working tree into a high-quality commit by:
- scoping which changes belong to this commit,
- checking diffs for tech debt, shortcuts, and privacy/security regressions,
- preparing clear commit messages (what + why),
- suggesting appropriate test commands to run before committing.

## Workflow

1. Clarify commit intent and scope
   - Ask the user for a short description of this commit:
     - what behavior it changes or fixes,
     - whether this is feature work, bugfix, refactor, or CI-only.
   - Ask whether:
     - backend tests (`backend-testing`) and frontend/browser tests (`frontend-testing`, `browser-e2e`) have already been run for these changes,
     - or they should be run before finalizing the commit.

2. Inspect working tree and select files

   - Run `git status -sb` and show the list of changed files.
   - Ask the user:
     - which files should be included in this commit,
     - which files (if any) should be left for another commit or reverted.
   - Based on the answer, form an explicit list of paths for this commit and ignore the rest.

3. Review diffs for scope and technical debt

   - For each selected file:
     - show a focused diff chunk-by-chunk (or summarized, if too large),
     - call out any signs of:
       - mixed concerns (API + services + DB changes in one place),
       - duplicated logic or validation,
       - “just one more if” in already complex functions,
       - commented-out code, unused helpers, or obvious shortcuts.
   - For any suspicious change:
     - suggest a cleaner alternative (refactor, helper extraction, moving logic into services),
     - ask the user whether to adjust the code before committing or explicitly mark debt (with `TODO(debt/...)` and issue reference).
   - For security/privacy-sensitive areas (auth, sessions, export, health data):
     - double-check that no PII is added to URLs, logs, or HTML/JSON responses,
     - ensure cookie and error-handling invariants from `AGENTS.md` are still respected.

4. Propose tests to run (if not already done)

   - If the user indicated that tests have not been run, or changes were added since:
     - recommend:
       - targeted backend tests (`backend-testing` skill / `go test` commands),
       - frontend tests (`frontend-testing` skill / `npm run lint`, `npm run build`),
       - browser e2e (`browser-e2e` skill) for UI-sensitive changes.
   - Ask explicitly:
     - which tests should be run now,
     - whether to block commit until they are green.
   - If the user chooses to proceed without rerunning some tests, note this explicitly in the summary.

5. Prepare the commit message

   - Build a commit message with:
     - an imperative subject line (≤72 characters) that describes what the commit does (not how);
     - a short body explaining *why* the change is needed (linking to bug/feature or user impact);
     - a bullet list of key changes, including:
       - domains touched (auth, days/symptoms, settings, stats, export, onboarding),
       - any tests added or updated,
       - any explicit `TODO(debt/...)` entries (with owner and reason),
       - any notable privacy/security-related adjustments.
   - Show the subject and body to the user and ask for edits until they are satisfied.

6. Output staging and commit commands

   - Once the message and scope are approved:
     - if the user explicitly asks you to execute the commit flow, run only the approved `git` commands;
     - otherwise, output the recommended commands:
       - `git add <selected paths>`
       - `git commit -m "<subject>"` (and `-m "<body>"` if needed)
   - Remind the user that:
     - `git push` should only be run when the user explicitly asks for it,
     - further aggregation into a release will be handled by the `release-plan` skill when enough commits accumulate.

## Constraints

- Only run `git commit` or `git push` when the user explicitly asks for them.
- Never silently accept technical debt: always call it out and either refactor or mark it explicitly with `TODO(debt/...)` and mention it in the commit body.
- Always respect `AGENTS.md` and the local AI context set rooted at `AI_CONTEXT.md`:
  - no business logic in `internal/api`,
  - no direct DB access from handlers,
  - no weakening of privacy/security invariants.
