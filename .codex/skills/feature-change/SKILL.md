---
name: feature-change
description: Plan and implement an ovumcy code change (feature or fix) using existing domain services, clean layering, and then hand off to testing, commit, and governance-update skills.
---

## Purpose

Help the user implement a change in ovumcy by:
- mapping it to an existing domain (`auth`, `days/symptoms/cycle/viewer`, `settings/notifications`, `stats/dashboard/calendar`, `export`, `onboarding/setup`),
- changing services first and keeping `internal/api` transport-only,
- preparing clear inputs for backend/frontend testing skills,
- preparing a concise summary for commit and governance-update skills.

## Workflow

1. Clarify change, source, and domain
   - Determine how the change was requested:
     - explicit task: the user describes the desired behavior or feature;
     - error report: the user (or COMET) shows a failing test, stack trace, log snippet, or UI bug description.
   - If this is an error report:
     - restate in English what is currently happening and what should happen instead (from the end-user’s point of view);
     - turn it into an explicit change task before planning any code edits.
   - Ask or infer:
     - which pages or API endpoints are involved,
     - whether this affects only backend or also templates/JS.
   - Choose the primary domain from:
     - `auth/session/recovery/reset`,
     - `days/symptoms/cycle/viewer`,
     - `settings/notifications`,
     - `stats/dashboard/calendar`,
     - `export`,
     - `onboarding/setup`.

2. Map to services and API
   - Locate relevant services in `internal/services` for this domain.
   - Plan to:
     - extend or add service methods in `internal/services`,
     - then adapt `internal/api` handlers as thin transport (parsing, auth, CSRF, error mapping),
     - avoid adding business logic to handlers or templates.
   - When a self-hosted request looks like "admin capabilities" but only needs local operator control, prefer a small CLI under `internal/cli` plus focused services instead of adding a new browser-admin surface.

3. Produce a small numbered plan
   - Before editing files, output a numbered list of small steps (5–10), each with:
     - which files in `internal/services` will change,
     - which files in `internal/api` will change,
     - whether `internal/templates` or `web/src/js` will change.
   - Mark any privacy/security-sensitive steps explicitly.
   - Wait for user approval before editing files.

4. Implement services first
   - Modify or add service methods:
     - reuse existing domain types and error taxonomy,
     - return typed/domain errors to be mapped by the centralized API error mapping layer.
   - Add/extend unit tests in `internal/services` for happy paths, edge cases, and error conditions.

5. Adapt API and transport
   - Update `internal/api` handlers/helpers to:
     - call the new/extended service methods,
     - use centralized API error mapping (no inline status/message switches),
     - use shared content negotiation and status markup helpers (error + dismissible success), following `AGENTS.md` rules (no ad-hoc header parsing or inline status HTML).
   - Add or update API regression tests to assert:
     - status codes + error keys,
     - HTMX vs full-page behavior,
     - redirects/flash where applicable.

6. Frontend impact (if any)
   - If templates or JS under `web/` change:
     - list which files are touched;
     - list manual UI flows to test (by role: owner/partner, and pages affected).
   - If templates or Tailwind/CSS classes are touched, explicitly call out any `web/static/*` changes so the user can decide later (in commit skill) whether to include them.
   - If the change touches any “today” date logic, always check and reuse the request-local timezone propagation (`ovumcy_tz` cookie + `X-Ovumcy-Timezone` header) and outline at least one UTC boundary regression test.

7. Handoff to testing, commit, and governance-update
   - Summarize for downstream skills:
     - what changed in services,
     - what changed in API,
     - what changed in templates/JS (if anything),
     - which tests were added/updated.
   - Recommend next skills to invoke, in order:
     1. backend-testing skill (e.g. suggest targeted `go test` packages or `go test ./...`);
     2. frontend-testing skill (e.g. suggest `npm run lint:js` / `npm run build` and key UI flows);
     3. `commit` skill to prepare a clean commit once tests are green;
     4. governance-update behaviour from `AGENTS.md` (proposing snippets for the local AI context set and skill files when new patterns or invariants were introduced).
   - For auth flows, explicitly check that redirect URLs on validation errors do not contain PII (email, tokens, error messages) in the query string or fragment.

## Constraints

- Always respect `AGENTS.md`:
  - no business logic in `internal/api`,
  - no direct DB access from handlers,
  - use centralized error mapping, negotiation, and markup helpers.
- Prefer extending existing services/domains over creating new ones.
- Never weaken privacy/security invariants; any security-sensitive behavior must be covered by tests.
