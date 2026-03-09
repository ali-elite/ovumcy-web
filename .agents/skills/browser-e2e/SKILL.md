---
name: browser-e2e
description: Use Playwright to run ovumcy end-to-end browser flows based on a built-in scenario checklist, interpret failures, and summarize UI readiness for commit and release.
---

## Purpose

Help the user validate ovumcy behavior in a real browser by:
- selecting relevant end-to-end flows from the built-in checklist (auth, onboarding, dashboard, calendar, stats, settings, language, navigation, security, etc.),
- mapping these flows to Playwright tests (existing or new),
- running the selected tests, interpreting failures, and suggesting fixes or follow-up implementation work,
- summarizing UI readiness for commit and for inclusion in release regression plans.

## Workflow

1. Clarify scope and target areas
   - Ask the user (or read from the previous skill) what changed or which bug is being verified:
     - affected pages and flows (e.g. login, forgot-password, dashboard, calendar, settings, export, navigation, language),
     - whether this is a narrow re-check for one bug or a broader smoke/regression pass,
     - whether any security/privacy-sensitive behavior (auth, sessions, health data) is involved.
   - Using the scenario checklist below, map the change to one or more areas (Authentication, Recovery, Onboarding, Dashboard, Calendar, Stats, Settings, Language, Privacy, Navigation, Cross-browser, Accessibility, Security).
   - Show the selected areas and ask whether to:
     - run only these areas, or
     - also add a small smoke set from other core areas.

2. Select concrete scenarios and tests

   - For each selected area:
     - choose a small set of representative scenarios from the checklist (3–10 per run), including at least:
       - one happy-path flow,
       - one error/edge-case flow.
   - For each scenario:
     - map it to an existing Playwright test (file + test name), if available; or
     - propose a new test to be generated, with a suggested file path, e.g. `e2e/auth.spec.ts`, `e2e/dashboard.spec.ts`.
   - Present the list of planned tests (files + scenario names) and ask the user to confirm or adjust the selection before running or generating anything.

3. Generate or update Playwright tests (when needed)

   - For scenarios without existing coverage:
     - generate Playwright test code that:
       - sets up state safely (test accounts, fixtures; no real credentials or PII),
       - performs the steps from the scenario (navigation, input, clicks),
       - asserts expected UI behavior via stable selectors or data attributes.
   - For frontend CSP refactors, prefer stable `data-*` hooks over framework-specific attributes when generating or updating selectors. Treat failures caused only by removed `x-*` markup as test-maintenance work first, not product regressions.
   - For settings UI simplifications, include one assertion that the removed control is no longer rendered so browser coverage catches stale templates or stale asset bundles after the refactor.
   - propose the exact test file(s) and test names.
   - For each new or updated test:
     - show the diff (or full snippet) and ask the user to approve it before writing.
     - after approval, apply the change and confirm that the test file is ready.

4. Run Playwright tests and interpret results

   - For reproducible ovumcy runs, execute Playwright through the repository runner (`npm run e2e` or `npm run e2e:ci`) so server startup, temporary DB isolation, and log capture are consistent.
   - If onboarding/stateful flows fail with a generic save error, inspect `.tmp/e2e/app-*.log` before classifying the issue as a UI regression.
   - If a failure looks flaky in CI, rerun with `npm run e2e:ci` at least twice before calling it a product regression, and include worker mode plus `.tmp/e2e/app-*.log` path in the summary.
   - For HTMX actions guarded by the shared confirm modal, explicitly accept the modal before interacting with rerendered controls; otherwise later clicks may be blocked by the overlay and look like false UI regressions.
   - Ask which Playwright command to use, for example:
     - `npx playwright test e2e/auth.spec.ts`,
     - or a narrower command targeting only the selected tests.
   - After explicit approval, run the selected tests and capture:
     - which tests passed,
     - which tests failed, and their error messages.
   - For each failing test:
     - summarize the scenario in plain language,
     - show the key failure reason (timeout, locator not found, assertion mismatch, navigation error, etc.),
     - classify the failure as:
       - likely regression caused by the recent change,
       - pre-existing bug revealed by better coverage,
       - test flakiness or too-brittle assertion.
     - propose concrete next steps:
       - return to `feature-change` (or equivalent implementation skill) to adjust code,
       - strengthen or stabilize the test if it is flaky,
       - or temporarily mark the scenario as a known issue (with TODO and issue reference) if you explicitly choose that.

5. Security, privacy, and cross-browser focus

   - If the change touches auth, sessions, roles, export, or health-related data:
      - prioritize scenarios from the Authentication, Recovery, Settings (Password / Clear Data / Delete Account), Navigation, and Security sections of the checklist.
   - Offer optional cross-browser runs:
      - run critical flows at least in two engines (e.g. Chromium and WebKit),
      - run key screens in a mobile viewport (login, dashboard, calendar, settings).
   - For timezone-sensitive browser flows, run at least one smoke scenario outside Chromium when the change alters cookie format or client-to-server timezone transport.
   - Ask the user whether to include cross-browser and mobile-view runs now or leave them for release-level regression.

6. Handoff summary for commit and release planning

   - Produce a concise summary:
     - which Playwright commands were run,
     - which areas and scenarios were covered,
     - which tests passed and which failed,
     - which failures (if any) block the current change from being committed.
   - If all selected scenarios are green and there are no blocking issues:
     - state that browser e2e checks are green for this change and it is ready for:
       - the `commit` skill to prepare a commit,
       - and for inclusion in the `release-plan` skill’s regression plan.
   - If there are blocking failures:
     - clearly state that the change is not ready to commit,
     - list the failing scenarios and suggest returning to the implementation skill with these as explicit tasks.

## Constraints

- Always respect `AGENTS.md` and the local AI context set rooted at `AI_CONTEXT.md`:
  - do not use real user data, secrets, or production credentials in tests,
  - never weaken security/privacy invariants just to make tests pass.
- Do not run Playwright or any other commands until the user has approved the e2e scope or directly asked for the command run.
- Never modify application code as part of this skill without a separate, user-approved implementation plan.
- Any new or updated test code must be shown and approved before being written.

## Scenario checklist (for this skill)

Use these as building blocks when selecting scenarios for a given change.

### Authentication & Account

- Register with valid data → verify redirect to recovery code screen  
- Register with already existing email → verify error message  
- Register with mismatched passwords → verify error message  
- Register with weak password (too short, no uppercase, no digit) → verify error message  
- Register with invalid email format → verify validation  
- Register with empty fields → verify validation  
- Login with correct credentials → verify redirect to dashboard or onboarding  
- Login with wrong password → verify generic error (not revealing which field is wrong)  
- Login with non-existent email → verify error message  
- Login with empty fields → verify browser/custom validation  
- Login with "Stay signed in for 30 days" checked → verify session persists after browser close  
- Login without "Stay signed in" → verify session ends on browser close  
- Show/hide password toggle on login, register, settings → verify it works on all fields  
- Logout → verify redirect to login, session destroyed, no back-navigation possible  

### Recovery Code Flow

- Recovery code shown only once after registration → verify it disappears on re-login  
- "Copy code" button → verify clipboard content matches the code  
- "Download code" button → verify file downloads with correct content  
- Cannot proceed without checking the confirmation checkbox → verify blocking behavior  
- Enter correct recovery code on `/forgot-password` → verify next step loads  
- Enter incorrect recovery code → verify error message displayed  
- Enter recovery code with wrong format → verify validation  
- Regenerate recovery code in Settings → verify old code stops working  

### Onboarding

- Onboarding is shown only on first login → verify it does not repeat on subsequent logins  
- Step 1: select date from quick-pick buttons → verify date is reflected in the input field  
- Step 1: manually enter date in input → verify it is accepted  
- Step 1: try to proceed without selecting a date → verify blocking/validation  
- Step 1: dates older than 60 days are not shown in the list → verify boundary  
- Step 2: drag cycle length slider → verify value updates in real time  
- Step 2: drag period length slider → verify value updates in real time  
- Step 2: toggle "Auto-fill period days" on/off → verify setting is saved  
- Step 3: "Start using Ovumcy" → verify redirect to dashboard  
- "Back" button on each step → verify state is preserved (no data loss)  
- Reload mid-onboarding → verify step state is preserved or gracefully reset  

### Dashboard

- Today's date and weekday displayed correctly → verify matches system date  
- Current phase label is correct relative to cycle start date  
- Cycle day number is correct relative to cycle start date  
- "Next period" date is correct (cycle start + cycle length)  
- "Ovulation" date is correct (cycle start + ~14 days)  
- Toggle "Period day" on → verify day is marked, intensity/symptom form becomes active  
- Toggle "Period day" off → verify entry is cleared or unflagged  
- Select flow intensity (None / Light / Medium / Heavy) → verify only one is selected at a time  
- Select multiple symptoms across different categories → verify all are highlighted  
- Deselect a symptom → verify it returns to unselected state  
- Add text to Notes field → verify it is saved  
- Save entry → verify success feedback and data persisted  
- Save entry → re-open dashboard → verify data is still there  
- "Clear today's entry" → verify all fields reset  
- "Clear today's entry" without confirmation dialog → test for accidental data loss (known issue)  
- Cycle summary section: verify stats (average, median, period length, fertile window) match actual data  

### Calendar

- Current month displayed by default  
- Navigate to previous month → verify dates and phase markings update  
- Navigate to next month → verify future days show predictions only (no past data)  
- "Today" button → verify calendar jumps back to current month and highlights today  
- Click a past day → verify correct entry loads in the side panel  
- Click a future day → verify side panel shows predicted phase, no editable data  
- Click today → verify side panel matches dashboard entry  
- Edit an entry for a past day from calendar → save → verify persisted  
- Phase color coding legend → verify all types (period, prediction, fertile window, ovulation) are present and accurate  
- Ovulation day shows correct icon/marker on the correct date  
- Period days (actual) are visually distinct from predicted period days  

### Statistics

- Page loads without errors for a new account (no cycles yet) → verify empty states  
- Average cycle shown as "-" when fewer than 3 full cycles recorded  
- Median cycle shown as "-" when fewer than 3 full cycles recorded  
- Average period length shown correctly after at least 1 cycle  
- Current phase tile matches dashboard  
- "Cycle length" chart appears after enough data is recorded  
- "Symptom frequency" chart appears after symptoms are logged  
- Baseline (default cycle length) shown as dashed line on chart  

### Settings — Profile

- Enter a display name → save → verify it appears in the nav header instead of email  
- Enter a name longer than allowed → verify validation  
- Email field is read-only → verify it cannot be changed here  
- Save profile with empty name → verify it clears display name (falls back to email)  

### Settings — Cycle Parameters

- Change cycle length slider → save → verify dashboard and calendar update  
- Change period length slider → save → verify calendar period markers update  
- Change last period start date → save → verify all phase predictions recalculate  
- Toggle "Auto-fill period days" → save → verify behavior changes on next period entry  
- Date pickers → verify invalid dates are rejected (e.g. Feb 30)  

### Settings — Change Password

- Change password with correct current password and valid new password → verify success  
- Change password with wrong current password → verify error  
- New password too weak → verify error matching registration rules  
- New password and confirmation do not match → verify error  
- After password change → logout → login with new password → verify works  
- After password change → login with old password → verify rejected  

### Settings — Data Export

- Export CSV → verify file downloads, correct headers and data rows  
- Export JSON → verify file downloads, valid JSON, correct structure  
- Export with date range filter → verify only records within range are included  
- Export with preset "Last 30 days" → verify date range auto-fills correctly  
- Export with preset "All time" → verify all records included  
- Export on account with no data → verify empty file or graceful message  

### Settings — Clear All Data

- Click "Clear all data" → verify confirmation prompt appears (or note absence)  
- Confirm clear → verify calendar is empty, dashboard shows no entries, cycle settings reset  
- Clear data → verify statistics page shows empty state  

### Settings — Delete Account

- Enter wrong password → verify account is NOT deleted, error shown  
- Enter correct password → verify account deleted, redirect to login  
- Try logging in with deleted account → verify rejected  
- After deletion → verify no residual data is accessible  

### Language Switching (RU / EN)

- Switch to EN on login page → verify all text translates (labels, buttons, errors)  
- Switch back to RU → verify correct active state styling on the toggle  
- Switch language while logged in → verify nav items, dashboard labels, calendar translate  
- Switch language on settings page → verify page re-renders in correct language  
- Language preference persists across page reload  

### Privacy Policy Page

- Accessible without login → `/privacy` loads for unauthenticated users  
- "Back" link → verify navigation works  
- GitHub link → verify it opens in new tab and points to correct repo  
- Breadcrumb "Home" link → verify it points to login (unauthenticated) or dashboard (authenticated)  

### Navigation & Routing

- Unauthenticated user accessing `/dashboard` → verify redirect to `/login`  
- Unauthenticated user accessing `/calendar`, `/stats`, `/settings` → verify redirect to `/login`  
- Authenticated user accessing `/login` or `/register` → verify redirect to `/dashboard`  
- Authenticated user accessing `/onboarding` after completing it → verify redirect to `/dashboard`  
- Logo click → verify navigates to dashboard when logged in, login when not  
- Browser back button after logout → verify protected page is not accessible  
- Direct URL to `/recovery-code` without registering → verify graceful redirect or error  

### Cross-Browser & Responsive

- All pages render correctly in at least Chrome and one other engine  
- Login/Register forms usable on mobile viewport (e.g. 375px width)  
- Dashboard symptoms grid reflows correctly on mobile  
- Calendar usable on mobile (side panel collapses/scrolls appropriately)  
- Settings sliders are touch-friendly on mobile  
- Language switcher visible and tappable on small screens  
- Onboarding date picker buttons scroll/wrap correctly on mobile  

### Accessibility (a11y)

- All form fields have associated labels  
- Keyboard navigation: Tab through all form fields in correct order  
- Enter key submits forms where appropriate  
- Show/hide password buttons are keyboard accessible  
- Error messages are announced appropriately (e.g. role="alert" or aria-live)  
- Color is not the only indicator of state (e.g. active symptom, active phase)  
- Language toggle has appropriate labeling (e.g. aria-label or group label)  

### Security

- Passwords are not visible in page source or network responses  
- Recovery code is not stored in URL or localStorage in plain text  
- Error messages do not reveal whether an email exists (login)  
- Session cookie has HttpOnly and Secure flags where applicable  
- Multiple failed logins → verify rate limiting or lockout behavior  
- XSS: entering `<script>alert(1)</script>` in Notes/profile → verify it is escaped  
- CSRF: verify state-changing requests include CSRF token and behave correctly
