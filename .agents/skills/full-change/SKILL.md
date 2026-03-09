---
name: full-change
description: "Take an ovumcy feature or bug report and drive it end-to-end: implement changes, run backend/frontend/browser tests, prepare a commit, and propose AI context and governance updates."
---

## Purpose

Help the user go from “idea or bug” to “ready-to-push commit and updated governance” in one guided flow by:
- clarifying the change from the user’s or reporter’s point of view,
- implementing code changes with clean layering and existing domains (like `feature-change`),
- running backend, frontend, and browser e2e checks,
- auditing test quality when changed or newly added tests look brittle or overly implementation-coupled,
- preparing a scoped commit with a clear message,
- proposing small updates to the local AI context set, `AGENTS.md`, and skills when new patterns or invariants appear.

## Inputs

- A short description of the change request, which can be:
  - an explicit feature/bug description from the user, or
  - an error report from COMET (failing test, stack trace, screenshot, unexpected UI behavior).
- Optional:
  - whether to run a **full** test cycle (backend + frontend + e2e) or a **minimal** one (backend + frontend only),
  - any constraints on commands (for example, “do not run go test ./...” or “no Playwright in this environment”).

## Workflow

1. Understand the change request and domain

   - Before planning or auditing a new non-trivial cycle, re-read `AI_CONTEXT.md` and the linked `.agents/context/*.md` files, not just the root context index, so domain and security assumptions come from the full local AI context set.
   - Determine the source:
     - explicit task (new feature or known bug),
     - error report (test failure, stack trace, UI bug).
   - If it is an error report:
     - restate in English what is currently happening and what should happen instead, from the end-user’s point of view;
     - turn it into an explicit change task.
   - Ask for or infer:
     - affected pages/API endpoints,
     - whether backend, frontend, or both are involved.
   - Map the task to a primary ovumcy domain using the local AI context set rooted at `AI_CONTEXT.md`:
     - `auth/session/recovery/reset`,
     - `days/symptoms/cycle/viewer`,
     - `settings/notifications`,
     - `stats/dashboard/calendar`,
     - `export`,
     - `onboarding/setup`,
     - navigation/i18n if relevant.
   - Mark whether the change is privacy/security-critical according to `AGENTS.md`.

2. Build an end-to-end plan

   - Produce a small numbered plan (7–12 steps) covering the entire flow:
     - steps for implementation (service/API/template/JS changes),
     - steps for backend tests,
     - steps for frontend lint/build,
     - steps for browser e2e checks,
     - a step for test-quality audit when tests are added, updated, or suspected to be brittle,
     - steps for commit preparation,
     - steps for governance/context updates (if needed).
   - For each step, indicate:
     - which files or skills it will touch (e.g. “services: X”, “API: Y”, “tests: Z”, “browser-e2e scenarios: auth + recovery”);
     - whether it is privacy/security-sensitive.
   - Ask the user:
     - to confirm or adjust the plan,
     - to choose test mode: **full cycle** (backend + frontend + browser e2e) or **minimal** (backend + frontend only).
   - Only proceed after explicit approval of the plan.

3. Implement the change (feature-change logic)

   - Follow the implementation steps from the plan, applying the rules from `AGENTS.md` and the local AI context set rooted at `AI_CONTEXT.md`:
     - change or extend `internal/services/...` first, reusing existing types and error taxonomy;
     - adapt `internal/api/...` handlers as thin transport only;
     - change templates/JS under `web/` when needed, respecting timezone and privacy rules.
   - For localization changes, treat locale JSON files, server-side date formatting, template switchers, browser `Intl` helpers, generated JS bundles, and `DEFAULT_LANGUAGE` docs as one cross-cutting change set.
   - When a quality cleanup targets report-card or gocyclo debt, prefer extracting shared test request/response helpers and splitting large regression tests before changing domain behavior.
   - For browser-facing security hardening in ovumcy, add baseline response headers first and treat strict CSP as a separate refactor when templates still rely on inline scripts or Alpine inline expressions.
   - For PWA/mobile install work in ovumcy, treat service workers and offline caching as privacy-sensitive because the app handles health data; prefer manifest/install UX first and require explicit approval before adding offline behavior.
   - For security-driven cookie cryptography fixes, prefer explicit cookie-format version bumps with regression coverage over silent derivation changes that preserve weak legacy behavior.
   - Add or update unit tests in services and API handlers:
     - happy paths,
     - edge cases,
     - unauthorized/forbidden behavior for security-sensitive flows.
   - At significant points (for example potential tech debt or a big refactor), pause and:
     - show the diff for that step,
     - ask the user whether to accept the cleaner refactor or explicitly mark any necessary `TODO(debt/...)`.
   - After completing implementation, summarize:
     - what changed in services/API/web,
     - what tests were added or updated.

4. Run backend tests

   - Based on changed packages, propose a small set of `go test` commands (as in `backend-testing`):
     - targeted packages that match the changed behavior and layer ownership,
     - optional `go test ./...` for broader safety if agreed.
   - Prefer behavior-based backend checks over markup-coupled regressions:
     - service tests for domain rules and persistence,
     - API tests for stable outcomes, typed errors, redirects, flash/session behavior, and unauthorized/forbidden paths.
   - Ask which commands may be run in this environment.
   - For each approved command:
     - run it,
     - interpret failures (which tests failed, why),
     - if failures occur, suggest returning to the implementation step before proceeding further.
   - If the user chooses to continue despite some failing tests, record this explicitly in the final summary.

5. Run frontend checks

   - If any `web/` or template changes occurred:
     - propose `npm run lint` / `npm run lint:js` and `npm run build` (as in `frontend-testing`);
     - ask which commands are available and allowed.
   - Run approved commands and interpret output:
     - summarize errors by file/line/message,
     - suggest concrete fixes.
   - If `npm run build` changed `web/static/*`:
     - show that these are generated assets;
     - ask whether they should be kept for this commit or reset before staging.
   - Define manual UI flows for quick verification (by role + device) based on affected pages.
   - Do not treat thin template smoke tests or raw script-fragment checks as sufficient validation for JS-driven behavior; prefer manual UI flows plus Playwright for client-side outcomes.
   - For deployment-only or self-hosted changes, prefer compose/runtime smoke validation over the normal backend/frontend/e2e cycle and use `release-plan` for the detailed validation checklist.

6. Run browser e2e checks (optional per plan)

   - For Docker-related changes, explicitly verify that public image workflows use a production-only Dockerfile, and keep Playwright/e2e execution in non-publishing CI jobs.
   - For date-sensitive browser checks, derive the persisted target day from rendered action endpoints (for example `/api/days/YYYY-MM-DD`) and assert saved data by that day instead of relying on a subsequent "today" reload.
   - For auth/privacy fixes, require both API assertions (status + stable error code) and HTML/flash assertions (localized message + absence of enumeration phrases/PII in URL).
   - When auth or recovery flows change, add transport tests that assert secrets stay out of JSON and URLs, and browser tests that verify the dedicated issuance page rather than inline secret rendering.
   - For client-side preference features (for example theme toggles), include a Playwright smoke that verifies persistence after reload and in a second page of the same browser context.
   - When a strict CSP refactor removes Alpine or inline handlers, audit existing Playwright helpers for selectors tied to `x-*` attributes before treating browser failures as product regressions.
   - If the user chose **full cycle** and the environment allows Playwright:
     - select relevant scenarios from the built-in browser-e2e checklist (auth, recovery, dashboard, calendar, settings, language, navigation, security);
     - prefer a small set of meaningful user-facing scenarios over broad markup-driven assertions;
     - map them to existing or new Playwright tests;
     - ask which tests to run.
   - Run the selected Playwright tests:
     - list passed and failed scenarios,
     - for each failure, classify it (regression / old bug / flaky test) and suggest next steps.
   - If serious failures are found, ask whether to:
     - fix them in this cycle (return to implementation), or
     - document them as known issues and block the commit.

7. Audit test quality (when relevant)

   - If the change set:
     - adds or updates tests, or
     - exposes brittle regressions, markup-heavy assertions, or implementation-coupled checks,
     call `test-suite-auditor`.
   - Provide the audit with:
     - changed test files,
     - suspected weak assertions,
     - overlap with service tests or Playwright coverage.
   - Use the audit to decide whether a test should:
     - stay as-is,
     - be reduced to thin smoke coverage,
     - be rewritten around behavior,
     - or be moved to browser e2e.
   - If the audit finds weak tests that materially lower confidence in a privacy/security-sensitive flow, return to implementation and test updates before preparing the commit.

8. Prepare the commit

   - Show `git status -sb` and ask which files belong to this commit.
   - Review diffs of selected files:
     - call out any potential tech debt, duplication, or rule violations (layering, privacy, cookies, error mapping);
     - propose cleaner alternatives or explicit `TODO(debt/...)` entries.
   - Confirm that all agreed backend/frontend/e2e tests for this change have been run (or explicitly waived).
   - If step 7 ran, confirm that the test-quality audit findings are either addressed or explicitly deferred.
   - Prepare a commit message:
     - imperative subject (≤72 chars),
     - short “why” explanation,
     - bullet list of key changes, tests, and any `TODO(debt/...)`.
   - Ask the user to edit/approve the message.
   - Unless the user explicitly asks you to execute them, output:
     - `git add <paths>`
     - `git commit -m "<subject>"` (and `-m "<body>"` if needed).

9. Propose governance and context updates

   - If the change introduced new flows, invariants, or patterns:
     - propose 1–3 small English snippets for the local AI context set (`AI_CONTEXT.md` or `.agents/context/*.md`) for updated flows or invariants,
     - propose 1–3 snippets for `AGENTS.md` (new rules or clarifications),
     - propose 1–3 snippets for existing or new skills if a recurring pattern was discovered.
   - For each snippet:
     - suggest the exact target file and section.
   - Show all snippets and locations and ask which to accept.
   - After the user applies accepted changes (manually or via diffs), re-read the files and:
     - highlight any obvious duplication or contradictions,
     - propose a tiny restructuring plan (1–5 steps) if needed.

10. Final summary

   - Present a concise end-to-end summary:
     - domains and files changed,
     - tests run (backend, frontend, browser), with pass/fail status,
     - whether test-quality audit ran and what it concluded,
     - whether a clean commit is ready (with suggested commands),
     - which governance/context updates were proposed and which were accepted.
   - If everything is green and the commit is ready:
     - suggest considering `release-plan` when enough such changes are accumulated.

## Constraints

- Always respect global Codex instructions, the local AI context set rooted at `AI_CONTEXT.md`, and `AGENTS.md`.
- After the user approves the end-to-end plan or directly requests execution, run only the commands that fit the agreed scope and environment constraints.
- Only run `git commit` or `git push` when the user explicitly asks for them.
- Do not silently accept technical debt or weaken privacy/security invariants.
