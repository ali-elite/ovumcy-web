---
name: frontend-testing
description: Plan and run ovumcy frontend checks (lint, build, and manual UI flows) after code changes, and summarize readiness for commit and browser e2e.
---

## Purpose

Help the user validate ovumcy frontend changes by:
- mapping edits in templates/JS/CSS to concrete lint/build commands,
- proposing and interpreting `npm` (or equivalent) frontend checks,
- defining manual UI flows to verify in the browser (desktop and mobile),
- preparing a concise status summary for commit and browser-e2e skills.

## Inputs

- A short description of what was changed on the frontend:
  - which files under `web/` were touched (`web/src/js`, templates, Tailwind/CSS, assets),
  - which pages or components are affected (auth, dashboard, calendar, settings, export, onboarding),
  - whether generated assets under `web/static/*` were updated.

## Workflow

1. Clarify scope of frontend changes
   - Ask which frontend files were edited:
     - JS under `web/src/js/...`,
     - templates under `internal/templates/...` or `web/templates/...`,
     - Tailwind/CSS or layout changes.
   - Identify affected flows by page and role (owner / partner / anonymous), e.g. “owner dashboard calendar”, “auth register/forgot-password”.

2. Select lint and build commands

   - Propose an ordered list of commands, for example:
     - `npm run lint` or `npm run lint:js` for JavaScript/TypeScript;
     - `npm run build` (or the closest available build script) to ensure assets compile.
   - For each command, briefly explain what it checks.
   - Ask the user which commands are available and which ones to run now.

3. Run and interpret frontend checks (interactive)

   - For each approved command:
     - run it and capture output;
     - if it succeeds:
       - confirm success and move to the next command.
     - if it fails:
       - summarize the main errors (file, line, message);
       - classify issues as:
         - clear mistakes in the recent change,
         - existing or non-blocking issues.
       - propose concrete fixes or adjustments to the change before proceeding.
   - If `npm run build` produced changes in `web/static/*`:
     - explicitly call out the diff and ask whether these generated assets should be:
       - included in the upcoming commit, or
       - discarded/reset before staging.

4. Define manual UI flows

   - Based on the affected pages and roles, propose manual test scenarios, for example:
     - Owner (desktop): login → navigate to affected page → perform key actions → verify expected UI and error handling.
     - Owner (mobile viewport): repeat critical flows (auth, dashboard, calendar, settings, export).
     - Partner/viewer (desktop + mobile): open shared views → navigate between states → ensure no private fields are visible.
   - If theme or palette changes are included, require a quick contrast pass in both light and dark modes on `/login`, `/dashboard`, and `/settings`.
   - For strict CSP or browser-hardening work, add a manual smoke pass with DevTools console open on `/login`, `/onboarding`, `/dashboard`, and `/settings` to catch CSP violations that lint/build will not surface.
   - For register form UX changes, include a check that password mismatch keeps both entered fields intact on the same page and shows inline error without URL query leakage.
   - Output these flows as concise, step-by-step checklists the user can follow in a real browser.
   - Ask the user to report any discovered UI issues so they can be turned into new change tasks.

5. Handoff summary for commit and browser-e2e

   - Produce a short summary including:
     - which frontend commands were run,
     - which passed and which (if any) failed,
     - whether `web/static/*` changes are intended to be committed,
     - which manual flows were executed (if the user confirmed them) and any reported issues.
   - If all planned commands are green and no blocking UI issues were found:
     - state that frontend checks are green for the current change and it is ready for:
       - the `commit` skill to prepare a commit,
       - the browser-e2e skill if a deeper automated UI pass is desired.
   - If there are remaining failures or unresolved UI issues:
     - clearly state that the change is not ready to commit and should go back to the implementation skill with concrete notes.

## Constraints

- Always respect project `AGENTS.md`:
  - do not silently include generated assets in commits without asking,
  - pay special attention to privacy-sensitive UI (auth, health data, owner/partner views).
- Do not run commands until the user has approved the frontend check scope or directly asked for the command run.
- Never modify frontend source files as part of this skill without an explicit, user-approved change plan.
