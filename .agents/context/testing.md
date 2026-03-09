# Ovumcy AI Context: Testing

## CI Expectations

- Browser e2e checks must run through the repository runner (`npm run e2e` locally, `npm run e2e:ci` in CI), which manages app lifecycle and an isolated temporary SQLite database under `.tmp/e2e`.
- Browser e2e remains a SQLite-based regression lane; Postgres confidence comes from dialect-aware backend bootstrap and integration coverage rather than a separate browser matrix.

## Playwright E2E Rules

- Playwright e2e runs in a dedicated CI runner job and must not be part of public image build/push workflows.
- CI Playwright policy is tiered: run a fast smoke subset on `push` to `main`, and run the full `npm run e2e:ci` suite on `pull_request`, published releases, or manual workflow dispatch.
- CI-gating e2e runs use serial Playwright execution for stateful flows to avoid SQLite write contention.
- Playwright selectors should prefer stable `data-*` hooks over framework-specific attributes such as `x-show` or serialized inline state. When strict CSP refactors remove inline framework markup, update browser helpers in the same change.
- When a form intentionally omits HTML `maxlength` to preserve server-side validation UX, add a browser test that submits an over-limit value and asserts the localized error message.

## Test Isolation and Failure Triage

- Browser e2e uses an isolated temporary SQLite database under `.tmp/e2e`.
- Failure triage should inspect `.tmp/e2e/app-*.log` together with Playwright `test-results/` artifacts.
