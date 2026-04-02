---
name: "release-plan"
description: "Generate an end-to-end release regression plan for ovumcy, including backend tests, frontend checks, privacy/security-sensitive areas, and manual UI and browser-based scenarios."
---

# Skill: release-plan

## Purpose

Help the user prepare a structured, repeatable release and regression plan for the ovumcy project by:
- enumerating what must be tested before a release,
- separating automated checks from manual and browser-based flows,
- highlighting privacy/security-sensitive areas and high-risk changes,
- linking to existing testing skills and rituals (backend, frontend, browser e2e).

This skill also owns the detailed validation checklist for deployment-only and self-hosted operational changes that would be too heavy to keep inside `full-change`.

## Inputs

- A short description of the release scope:
  - features and bugfixes included,
  - areas touched (backend / frontend / auth / exports / settings / calendar / roles / i18n),
  - known risks, DB migrations, or permission changes (if any).

## Workflow

1. Clarify the release scope
   - Ask the user for:
     - a short summary of what changed in this release (features, fixes, refactors),
     - which modules or domains were touched (auth, days/symptoms, stats, settings, export, onboarding, navigation, language),
     - whether there are:
       - DB migrations,
       - auth/permission changes,
       - export or health-data related changes.
   - Classify the release as:
     - small (few localized changes),
     - medium,
     - or high-risk (cross-cutting, migrations, auth/roles, or many domains).

2. Identify impacted domains

   - Map the described changes into ovumcy domains, for example:
     - Authentication & sessions
     - Roles and access control (owner vs partner)
     - Core tracking flows (logging data, editing, deleting)
     - Calendar and stats
     - Settings (cycle, profile, privacy, password, danger operations)
     - Data export / import
     - Onboarding and first-run setup
     - UI and i18n (ru/en)
   - For each domain, explicitly mark whether it is **privacy/security-critical** according to `AGENTS.md` and the local AI context set rooted at `AI_CONTEXT.md`.

3. Propose automated checks

   - For backend:
     - list concrete `go test` commands:
       - targeted packages for heavily-touched domains,
       - plus `go test ./...` for medium/high-risk releases.
     - note that these can be orchestrated via the backend-testing skill.
   - For frontend:
     - propose running `npm run lint` (or `npm run lint:js`) and `npm run build`.
     - call out any expectations around `web/static/*` changes.
     - if the README explicitly documents the current `main` branch, avoid or remove badges that report only on the latest tagged release, because they can misstate the quality of the release candidate being validated.
     - note that these can be orchestrated via the frontend-testing skill.
   - For browser-based E2E:
     - suggest a small smoke suite using the browser-e2e skill (e.g. login, onboarding, dashboard, calendar, settings, language switch, critical security flows).
     - if CI uses event-based gating, include two explicit lanes in the plan:
       - `push` lane: smoke subset only (fast regression signal),
       - `pull_request` / published `release` / `workflow_dispatch` lanes: full `npm run e2e:ci` suite.
     - name the smoke spec files explicitly so release owners can verify CI configuration and expected runtime.
     - for auth/OIDC releases, browser verification is not complete until the auth/recovery smoke subset and all three OIDC lanes (`fallback`, `hybrid`, `oidc_only`) pass on `chromium`, `firefox`, and `webkit`.
   - For deployment-only and self-hosted validation:
     - prefer compose validation and isolated runtime smoke tests over backend/frontend/e2e suites.
     - when README screenshots are refreshed, capture them from an isolated local instance with English UI state and a clean demo account, then keep the README screenshot list aligned with the actual files under `docs/screenshots/`.
     - when compose files use fixed container names, named volumes, or `env_file` paths, create temporary isolated copies for validation instead of running `docker compose up` against the default files.
     - for self-hosted config-surface changes, validate every documented profile with temporary `.env` files and `docker compose config`, and verify that runtime defaults, `.env.example`, and compose/example stack defaults do not drift.
     - for storage-engine work, decide up front whether the task is docs-only planning or actual runtime support. If it is actual support, require dialect-aware migrations, runtime boot changes, and validation for each supported engine before treating the task as complete.
     - for actual storage-engine support, backend verification must include bootstrap/idempotence coverage for each supported engine and a clear statement about whether cross-engine data migration is or is not part of the current change.
     - for supported alternative database paths, pair backend engine coverage with an operator-facing deployment example and a browser or runtime smoke lane that boots the app under that engine.
     - for Docker-backed Postgres integration tests, do not treat in-container `pg_isready` alone as sufficient readiness. Wait until the database is reachable from the host-side test client before treating bootstrap failures as application regressions.
     - for public self-hosted reverse-proxy example stacks, isolated validation should assert both that `/healthz` succeeds through the proxy and that `docker inspect` reports no host-published ports for the `ovumcy` service.
     - for public self-hosted database-backed reverse-proxy examples, validate compose structure and env contracts even when full runtime smoke is impractical due to real-domain TLS or certificate prerequisites, and call out that limitation explicitly.
     - for backup/restore changes to self-hosted Docker deployments, validate the runbook with an isolated stack: capture the SQLite file hash before backup, restore into a clean volume, bring the app back to healthy, and compare the restored hash before treating the documentation as verified.
     - for CI security automation in ovumcy, keep CodeQL in its own workflow and run `gosec`, Trivy filesystem/image scans, and SBOM generation in a separate workflow against a locally built runtime-only image instead of coupling scans to publish-only jobs.
     - when GitHub code scanning flags workflow supply-chain drift, prefer pinning Actions to immutable SHAs and keeping version comments rather than suppressing the warning or reverting to mutable tags.
   - Ask the user which commands and suites are acceptable in the current environment and mark them as part of the release checklist (not to be run automatically by this skill).

4. Define manual and browser UI regression flows

   - For each impacted domain, generate concrete flows, separated by role and device type:
     - Owner (desktop): main happy paths + key edge cases on affected pages.
     - Owner (mobile viewport): critical flows (auth, dashboard, calendar, settings, export, onboarding).
     - Partner/viewer (desktop + mobile): ensure read-only invariants and proper privacy sanitization.
   - For each flow:
     - specify whether it should be:
       - manual (human following steps in a browser),
       - or automated (covered by browser-e2e Playwright tests).
   - Use clear, step-by-step scenarios, e.g.:
     - “Owner: sign up → complete onboarding → log a few days → open calendar → verify stats.”
     - “Partner: open shared link → navigate between views → ensure no private fields are visible.”

5. Highlight privacy and security checks

   - For any flow that touches health-related or otherwise sensitive data:
     - add explicit checks:
       - no leakage of private fields in UI (owner-only data not visible to partner/viewer),
       - correct behavior when access is denied (proper redirects/403, no partial data leak),
       - correct handling of language/locale around errors and status messages.
   - For auth/session/roles/export/settings flows:
     - include checks derived from `AGENTS.md` (no PII in URLs, correct cookie flags, no token echoing).
   - Call out any areas where there is currently no automated coverage (unit/API/E2E) and:
     - suggest adding tests in future iterations,
     - mark them as TODO items in the release notes if they are accepted gaps.

6. Produce the final release plan

   - Output the plan as structured sections that can be followed in one sitting:
     - Summary of scope and risk areas
     - Automated checks
       - backend (`go test` commands / backend-testing skill)
       - frontend (`npm` commands / frontend-testing skill)
       - browser e2e (browser-e2e skill + key areas)
     - Manual UI flows
       - by role (owner / partner / anonymous)
       - by device type (desktop / mobile)
     - Privacy/security-specific checks
     - Known limitations and TODOs for future automation and coverage
   - Ensure each item is concrete (command or scenario) and can be checked off.

## Constraints

- Do not assume you can run commands; present all test commands and suites as suggestions to be executed separately.
- Do not promise full coverage if the codebase does not have relevant automated tests; clearly state gaps and propose follow-up work instead.
- Always respect global and project AGENTS instructions, especially around privacy, security, and architectural boundaries, and align domain mapping with the local AI context set rooted at `AI_CONTEXT.md`.
