# Phase 23: English Documentation Policy (v0.23.0)

## Requirement Snapshot

Restore repository Markdown documentation to English after the previous localization pass, and make English the maintained documentation language for README, requirements, plans, ADRs, and architecture notes.

## Task Checklist

- [x] Restore `README.md` as the full English project overview.
- [x] Remove the separate Chinese README split from the active documentation set.
- [x] Rewrite `requirement.md` in English and include history through v0.23.0.
- [x] Rewrite `docs/architecture` and `docs/adr` Markdown files in English.
- [x] Rewrite historical `.plan/` files in English and add this phase plan.
- [x] Update `VERSION` and the OpenAPI info version to `0.23.0`.
- [x] Run formatting, tests, race tests, vet, JSON validation, and diff checks.

## Non-Goals

- No runtime Go API behavior changes.
- No OpenAPI route semantic changes.

## Follow-Up Candidates

- Add a multilingual documentation site later if project governance requires it.
