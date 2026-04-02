---
name: governance-update
description: Propose and apply small, concrete updates to AI_CONTEXT, AGENTS, and skill files after non-trivial ovumcy changes.
---

## Purpose

Keep ovumcy governance and context files in sync with reality by:
- detecting when a change introduces new domains, flows, or cross-cutting rules,
- proposing small English snippets for the local AI context set, `AGENTS.md`, and skill files,
- helping the user review and apply these updates without breaking structure or introducing duplication.

## Inputs

- A short summary of recent changes (ideally from `feature-change` / `commit` / `release-plan`):
  - which domains and flows were touched,
  - which invariants or conventions were added or clarified,
  - which recurring patterns were discovered (for example new typical feature workflows).

## Workflow

1. Decide whether governance updates are needed
   - From the summary, determine whether recent changes:
     - added or significantly changed user-visible flows or invariants,
     - introduced new domain concepts or cross-cutting rules,
     - exposed a recurring pattern that deserves a dedicated skill or rule.
   - If nothing affects governance or context, explicitly say so and stop.

2. Propose AI context updates

   - Draft 1–3 small, concrete English snippets for the local AI context set (`AI_CONTEXT.md` or the linked files under `.agents/context/`) that:
     - describe new or updated flows, domains, or invariants,
     - fit into existing sections (e.g. domains, timezone, auth flows).
   - For each snippet:
     - propose the exact section and approximate position where it should be inserted or updated.
   - Show all proposed snippets with their target locations and ask the user which ones to accept or reject.

3. Propose agent rule updates

   - Draft 1–3 small, concrete English snippets for `AGENTS.md` that:
     - capture new rules or restrictions the agent should follow (for example new privacy rule, new layering constraint),
     - or clarify existing rules that were previously ambiguous.
   - For each snippet:
     - propose the section (and whether it should be “add”, “replace”, or “tighten”).
   - Show the snippets and locations, and ask for per-snippet approval.

4. Propose skill updates or new skills

   - Identify any recurring patterns that emerged (for example a new typical auth flow, a new kind of export, a new test ritual).
   - Propose 1–3 snippets for:
     - updating existing skills (e.g. `feature-change`, `backend-testing`, `browser-e2e`),
     - or a short outline of a new skill if the pattern does not fit existing ones.
   - For each suggested change:
     - show the text and target skill file/section,
     - ask whether to accept, adjust, or discard it.

5. Review for duplication and contradictions

   - After the user applies accepted snippets (manually or via approved diffs):
     - re-read `AI_CONTEXT.md`, the linked local context files under `.agents/context/`, `AGENTS.md`, and relevant skill files.
   - Look for:
     - duplicated rules between context, agents, and skills,
     - contradictory statements or outdated parts.
   - Propose a small restructuring plan (1–5 steps) to:
     - de-duplicate overlapping rules,
     - move purely contextual text from `AGENTS.md` into the local AI context set, or vice versa,
     - keep skills focused on workflows rather than restating all rules.
   - Prefer this split when restructuring:
     - `AI_CONTEXT.md` = entry point and concise summary for the local AI context set,
     - `.agents/context/*.md` = product/domain facts and flow-specific exceptions,
     - `AGENTS.md` = reusable agent rules and enforcement,
     - skill files = workflow steps that reference the AI context set / `AGENTS.md` instead of duplicating their wording.
   - Ask for approval before suggesting any further concrete edits.

## Constraints

- Never modify `AI_CONTEXT.md`, files under `.agents/context/`, `AGENTS.md`, or skill files without the user explicitly approving each change (snippet or diff).
- Keep governance updates small and incremental by default: prefer a few focused snippets over large rewrites. A larger approved governance refactor is acceptable when the goal is de-duplication, clearer ownership, or cleaner cross-references without changing policy.
- Always respect existing privacy/security invariants when proposing new rules.
